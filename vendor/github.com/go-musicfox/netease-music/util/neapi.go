package util

import (
	"bytes"
	"crypto/rand"
	"crypto/rc4"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/requests"
	"github.com/klauspost/compress/zstd"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	neapiConfigPublicKeyHex = "3048024100e4e79c1cb27a9fab7d2c740e99e923ec51c009169ec3a380c60404e01b060f0160cdc631d9fb7fe53b228a5ff11bee42b965d28c6c2036193f4ccdc7ed853d110203010001"
	neapiAndroidUA          = "NeteaseMusic/9.5.61.260802021928(9005061);Dalvik/2.1.0 (Linux; U; Android 12; HBN-AL00 Build/cd737a2.0)"
	neapiHost               = "https://interface3.music.163.com"
	neapiConfigURL          = "https://interface.music.163.com/api/gorilla/algorithm/record"
	neapiMagic              = "CSJM"
	neapiNonceLen           = 12
	neapiADLen              = 16
	neapiTagLen             = 16
	neapiHeaderLen          = 4 + 1 + 1 + 2 + neapiNonceLen + 1 + neapiADLen + 4
)

var (
	neapiKeyMu    sync.Mutex
	neapiKeyState *neapiKeyInfo
)

type neapiKeyInfo struct {
	Version    int
	CircleTime int64
	FetchedAt  int64
	KeyBlob    []byte
	EncryptKey []byte
	DecryptKey []byte
}

type neapiReply struct {
	Code int `json:"code"`
	Data struct {
		File       string `json:"file"`
		Version    int    `json:"version"`
		CircleTime int64  `json:"circleTime"`
	} `json:"data"`
}

func neapiZstdCompress(data []byte) []byte {
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		return nil
	}
	defer enc.Close()
	return enc.EncodeAll(data, nil)
}

func neapiZstdDecompress(data []byte) ([]byte, error) {
	decode := func(buf []byte) ([]byte, error) {
		dec, err := zstd.NewReader(bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}
		defer dec.Close()
		return io.ReadAll(dec)
	}
	out, err := decode(data)
	if err == nil {
		return out, nil
	}
	// 当前 gorilla/algorithm/record 下发的帧缺少最后的空块头，补齐后再解一次。
	fixed := append(append([]byte{}, data...), 0x01, 0x00, 0x00)
	if out2, err2 := decode(fixed); err2 == nil {
		return out2, nil
	}
	return nil, err
}

func neapiRSAPublicDecrypt(block []byte) ([]byte, error) {
	der, err := hex.DecodeString(neapiConfigPublicKeyHex)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKCS1PublicKey(der)
	if err != nil {
		return nil, err
	}
	c := new(big.Int).SetBytes(block)
	m := new(big.Int).Exp(c, big.NewInt(int64(pub.E)), pub.N)
	out := m.Bytes()
	if len(out) < pub.Size() {
		out = append(make([]byte, pub.Size()-len(out)), out...)
	}
	if len(out) < 3 || out[0] != 0 || out[1] != 1 {
		return nil, fmt.Errorf("neapi config rsa padding invalid")
	}
	i := 2
	for i < len(out) && out[i] == 0xff {
		i++
	}
	if i >= len(out) || out[i] != 0 {
		return nil, fmt.Errorf("neapi config rsa padding invalid")
	}
	return out[i+1:], nil
}

func neapiRC4(key, data []byte) []byte {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil
	}
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out
}

// --- Lua 5.3 chunk reader, ported from neapiConfig.js ---

type neapiChunkReader struct {
	buf []byte
	pos int
}

func (r *neapiChunkReader) u8() byte {
	if r.pos >= len(r.buf) {
		return 0
	}
	v := r.buf[r.pos]
	r.pos++
	return v
}

func (r *neapiChunkReader) u32() uint32 {
	if r.pos+4 > len(r.buf) {
		r.pos = len(r.buf)
		return 0
	}
	v := binary.LittleEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v
}

func (r *neapiChunkReader) i64() int64 {
	if r.pos+8 > len(r.buf) {
		r.pos = len(r.buf)
		return 0
	}
	v := int64(binary.LittleEndian.Uint64(r.buf[r.pos:]))
	r.pos += 8
	return v
}

func (r *neapiChunkReader) f64() float64 {
	if r.pos+8 > len(r.buf) {
		r.pos = len(r.buf)
		return 0
	}
	v := math.Float64frombits(binary.LittleEndian.Uint64(r.buf[r.pos:]))
	r.pos += 8
	return v
}

func (r *neapiChunkReader) str() string {
	size := int(r.u8())
	if size == 0 {
		return ""
	}
	if size == 0xff {
		size = int(r.u32())
	}
	bodyLen := size - 1
	if bodyLen <= 0 {
		return ""
	}
	if r.pos+bodyLen > len(r.buf) {
		r.pos = len(r.buf)
		return ""
	}
	body := r.buf[r.pos : r.pos+bodyLen]
	r.pos += bodyLen
	out := make([]byte, len(body))
	for i := range body {
		out[i] = body[i] ^ neapiStrXOR[i&0x0f]
	}
	return string(out)
}

var neapiCodeXOR = [8]uint32{0x94B4F825, 0x977C6795, 0x327DE32B, 0x16DA5BA1, 0xCF55A2EB, 0x080BB445, 0x021AF055, 0xFC56B8BD}
var neapiStrXOR, _ = hex.DecodeString("aefda2d1832b518a7741c325fb2687f0")

type neapiProto struct {
	code   []uint32
	consts []any
	upvals [][2]byte
	protos []*neapiProto
}

func loadNeapiProto(r *neapiChunkReader) *neapiProto {
	p := &neapiProto{}
	_ = r.str() // source
	_ = r.u32() // linedefined
	_ = r.u32() // lastlinedefined
	_ = r.u8()  // numparams
	_ = r.u8()  // isvararg
	_ = r.u8()  // maxstack

	count := int(r.u32())
	p.code = make([]uint32, count)
	for i := range count {
		p.code[i] = r.u32() ^ neapiCodeXOR[i&7]
	}

	count = int(r.u32())
	p.consts = make([]any, count)
	for i := range count {
		switch r.u8() {
		case 0:
			p.consts[i] = nil
		case 1:
			_ = r.u8()
			p.consts[i] = nil
		case 3:
			_ = r.f64()
			p.consts[i] = nil
		case 19:
			p.consts[i] = r.i64()
		default:
			p.consts[i] = r.str()
		}
	}

	count = int(r.u32())
	p.upvals = make([][2]byte, count)
	for i := range count {
		p.upvals[i] = [2]byte{r.u8(), r.u8()}
	}

	count = int(r.u32())
	p.protos = make([]*neapiProto, count)
	for i := range count {
		p.protos[i] = loadNeapiProto(r)
	}

	count = int(r.u32())
	r.pos += 4 * count
	count = int(r.u32())
	for range count {
		_ = r.str()
		_ = r.u32()
		_ = r.u32()
	}
	count = int(r.u32())
	for range count {
		_ = r.str()
	}
	return p
}

func loadNeapiChunk(buf []byte) (*neapiProto, error) {
	if len(buf) < 34 || !bytes.Equal(buf[:4], []byte{0x7f, 0x45, 0x4c, 0x46}) {
		return nil, fmt.Errorf("neapi config is not a chunk")
	}
	r := &neapiChunkReader{buf: buf, pos: 33}
	_ = r.u8()
	return loadNeapiProto(r), nil
}

func neapiOp(ins uint32) uint32    { return ins & 0x3f }
func neapiArgA(ins uint32) uint32  { return (ins >> 6) & 0xff }
func neapiArgB(ins uint32) uint32  { return (ins >> 23) & 0x1ff }
func neapiArgC(ins uint32) uint32  { return (ins >> 14) & 0x1ff }
func neapiArgBx(ins uint32) uint32 { return (ins >> 14) & 0x3ffff }

func neapiMainRegisters(main *neapiProto) map[uint32]any {
	regs := make(map[uint32]any)
	for _, ins := range main.code {
		if neapiOp(ins) == 1 {
			regs[neapiArgA(ins)] = main.consts[neapiArgBx(ins)]
		}
	}
	return regs
}

func neapiExtractKeyOffsets(script *neapiProto) (map[int64]int64, error) {
	if len(script.protos) == 0 {
		return nil, fmt.Errorf("neapi config has no create_ctx")
	}
	createCtx := script.protos[0]
	regs := neapiMainRegisters(script)
	offsets := make(map[int64]int64)
	var pending int64
	hasPending := false
	for _, ins := range createCtx.code {
		switch neapiOp(ins) {
		case 31: // EQ
			c := neapiArgC(ins)
			if c&0x100 != 0 {
				idx := c & 0xff
				if int(idx) < len(createCtx.consts) {
					if v, ok := createCtx.consts[idx].(int64); ok {
						pending = v
						hasPending = true
					}
				}
			}
		case 5: // GETUPVAL
			if !hasPending {
				continue
			}
			idx := neapiArgB(ins)
			if int(idx) >= len(createCtx.upvals) {
				continue
			}
			up := createCtx.upvals[idx]
			if up[0] != 1 {
				continue
			}
			if v, ok := regs[uint32(up[1])].(int64); ok {
				offsets[pending] = v
				hasPending = false
			}
		}
	}
	if _, ok := offsets[251]; !ok {
		return nil, fmt.Errorf("neapi config missing encrypt direction")
	}
	if _, ok := offsets[503]; !ok {
		return nil, fmt.Errorf("neapi config missing decrypt direction")
	}
	return offsets, nil
}

func neapiExtractKeyBlob(script *neapiProto) []byte {
	var candidates []string
	for _, c := range script.consts {
		if s, ok := c.(string); ok && len(s) >= 64 {
			if _, err := base64.StdEncoding.DecodeString(s); err == nil {
				candidates = append(candidates, s)
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return len(candidates[i]) > len(candidates[j]) })
	if len(candidates) == 0 {
		return nil
	}
	blob, _ := base64.StdEncoding.DecodeString(candidates[0])
	return blob
}

func neapiDecodeConfig(raw []byte) (*neapiKeyInfo, error) {
	if len(raw) < 68 {
		return nil, fmt.Errorf("neapi config too short")
	}
	key, err := neapiRSAPublicDecrypt(raw[:64])
	if err != nil {
		return nil, err
	}
	size := int(binary.LittleEndian.Uint32(raw[64:68]))
	if 68+size > len(raw) {
		return nil, fmt.Errorf("neapi config payload out of range")
	}
	ecfg, err := neapiZstdDecompress(neapiRC4(key, raw[68:68+size]))
	if err != nil {
		return nil, err
	}
	if len(ecfg) < 6 {
		return nil, fmt.Errorf("neapi config container too short")
	}
	version := int(binary.BigEndian.Uint16(ecfg[4:6]))
	script, err := loadNeapiChunk(ecfg[6:])
	if err != nil {
		return nil, err
	}
	offsets, err := neapiExtractKeyOffsets(script)
	if err != nil {
		return nil, err
	}
	blob := neapiExtractKeyBlob(script)
	if len(blob) < 32 {
		return nil, fmt.Errorf("neapi config key blob missing")
	}
	slice := func(dir int64) []byte {
		off := offsets[dir]
		if off < 0 || int(off)+32 > len(blob) {
			return nil
		}
		return blob[off : off+32]
	}
	return &neapiKeyInfo{
		Version:    version,
		KeyBlob:    blob,
		EncryptKey: slice(251),
		DecryptKey: slice(503),
	}, nil
}

func neapiFetchConfig() (*neapiKeyInfo, error) {
	req := requests.Requests()
	req.Client.Timeout = 15 * time.Second
	req.Header.Set("User-Agent", neapiAndroidUA)
	resp, err := req.Get(neapiConfigURL + "?version=0&behavior=open")
	if err != nil {
		return nil, err
	}
	var reply neapiReply
	if err := json.Unmarshal(resp.Content(), &reply); err != nil {
		return nil, err
	}
	if reply.Code != 200 || reply.Data.File == "" {
		return nil, fmt.Errorf("neapi config request failed")
	}
	raw, err := base64.StdEncoding.DecodeString(reply.Data.File)
	if err != nil {
		return nil, err
	}
	info, err := neapiDecodeConfig(raw)
	if err != nil {
		return nil, err
	}
	info.CircleTime = reply.Data.CircleTime
	info.FetchedAt = time.Now().UnixMilli()
	return info, nil
}

func getNeapiKey() (*neapiKeyInfo, error) {
	neapiKeyMu.Lock()
	defer neapiKeyMu.Unlock()
	if neapiKeyState != nil && neapiKeyState.EncryptKey != nil {
		if neapiKeyState.CircleTime <= 0 || time.Now().UnixMilli()-neapiKeyState.FetchedAt < neapiKeyState.CircleTime {
			return neapiKeyState, nil
		}
	}
	info, err := neapiFetchConfig()
	if err != nil {
		if neapiKeyState != nil {
			return neapiKeyState, nil
		}
		return nil, err
	}
	neapiKeyState = info
	return info, nil
}

func neapiEncrypt(data string, info *neapiKeyInfo) (string, error) {
	compressed := neapiZstdCompress([]byte(data))
	aead, err := chacha20poly1305.New(info.EncryptKey)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, neapiNonceLen)
	ad := make([]byte, neapiADLen)
	_, _ = randRead(nonce)
	_, _ = randRead(ad)
	sealed := aead.Seal(nil, nonce, compressed, ad)

	header := make([]byte, neapiHeaderLen)
	copy(header, neapiMagic)
	header[4] = byte(info.Version >> 8)
	header[5] = byte(info.Version & 0xff)
	binary.LittleEndian.PutUint16(header[6:], uint16(len(nonce)))
	copy(header[8:], nonce)
	header[8+len(nonce)] = byte(len(ad))
	copy(header[9+len(nonce):], ad)
	binary.LittleEndian.PutUint32(header[9+len(nonce)+len(ad):], uint32(len(sealed)))
	out := append(header, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

func neapiDecrypt(body string, info *neapiKeyInfo) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(body))
	if err != nil {
		return nil, err
	}
	if len(raw) < neapiHeaderLen {
		return nil, fmt.Errorf("neapi response too short")
	}
	nonceLen := int(binary.LittleEndian.Uint16(raw[6:]))
	if len(raw) < 8+nonceLen+1 {
		return nil, fmt.Errorf("neapi response invalid")
	}
	nonce := raw[8 : 8+nonceLen]
	adLen := int(raw[8+nonceLen])
	adStart := 9 + nonceLen
	if adStart+adLen+4 > len(raw) {
		return nil, fmt.Errorf("neapi response invalid")
	}
	ad := raw[adStart : adStart+adLen]
	sealedLen := int(binary.LittleEndian.Uint32(raw[adStart+adLen:]))
	sealedStart := adStart + adLen + 4
	if sealedStart+sealedLen > len(raw) {
		return nil, fmt.Errorf("neapi response invalid")
	}
	sealed := raw[sealedStart : sealedStart+sealedLen]
	aead, err := chacha20poly1305.New(info.DecryptKey)
	if err != nil {
		return nil, err
	}
	compressed, err := aead.Open(nil, nonce, sealed, ad)
	if err != nil {
		return nil, err
	}
	return neapiZstdDecompress(compressed)
}

func randRead(b []byte) (int, error) {
	return rand.Read(b)
}

func neapiParseBody(text string) []byte {
	if json.Valid([]byte(text)) {
		return []byte(text)
	}
	if i := strings.Index(text, "}{"); i >= 0 {
		return []byte(text[:i+1])
	}
	return []byte(text)
}

func createNeapiRequest(method, rawURL string, data map[string]string, options *Options) (float64, []byte, []*http.Cookie) {
	info, err := getNeapiKey()
	if err != nil {
		return 520, []byte(err.Error()), nil
	}
	uri := "/api"
	if options.Url != "" {
		uri = options.Url
	}
	body := url.Values{}
	for k, v := range data {
		body.Set(k, v)
	}
	encrypted, err := neapiEncrypt(body.Encode(), info)
	if err != nil {
		return 520, []byte(err.Error()), nil
	}

	globalCookieJar = GetGlobalCookieJar()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if len(UNMProxyURL) > 0 {
		if proxyURL, err := url.Parse(UNMProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	client := &http.Client{Jar: globalCookieJar, Transport: transport}
	if HTTPClientTimeout > 0 {
		client.Timeout = HTTPClientTimeout
	}
	cookies := append([]*http.Cookie{}, options.Cookies...)
	if u, err := url.Parse(rawURL); err == nil {
		cookies = append(cookies, globalCookieJar.Cookies(u)...)
	}

	target := neapiHost + "/neapi" + strings.TrimPrefix(uri, "/api")
	submit := func() (float64, []byte, []*http.Cookie) {
		httpReq, err := http.NewRequest("POST", target, strings.NewReader(encrypted))
		if err != nil {
			return 520, []byte(err.Error()), nil
		}
		httpReq.Header.Set("User-Agent", neapiAndroidUA)
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
		httpReq.Header.Set("X-ER", strconv.Itoa(info.Version))
		for _, cookie := range cookies {
			httpReq.AddCookie(cookie)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return 520, []byte(err.Error()), nil
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.Header.Get("X-USER-STATUS") == "777" || resp.Header.Get("X-USER-STATUS") == "999" {
			return 777, []byte("neapi config rotated"), resp.Cookies()
		}
		plain, err := neapiDecrypt(string(raw), info)
		if err != nil {
			return float64(resp.StatusCode), raw, resp.Cookies()
		}
		plain = neapiParseBody(string(plain))
		code := float64(resp.StatusCode)
		var parsed map[string]any
		if json.Unmarshal(plain, &parsed) == nil {
			if v, ok := parsed["code"].(float64); ok {
				code = v
			}
		}
		return code, plain, resp.Cookies()
	}
	code, plain, cookies := submit()
	if code == 777 {
		neapiKeyMu.Lock()
		neapiKeyState = nil
		neapiKeyMu.Unlock()
		if info, err = getNeapiKey(); err == nil {
			encrypted, _ = neapiEncrypt(body.Encode(), info)
			code, plain, cookies = submit()
		}
	}
	return code, plain, cookies
}
