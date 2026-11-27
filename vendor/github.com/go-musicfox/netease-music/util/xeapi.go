package util

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/requests"
)

const (
	xeapiStaticKeyHex = "ab1d5a430f6bb04a3f01e81ddd72bd916d5ce591248ac128714806d7f8fb1b84"
	xeapiSignKey      = "mUHCwVNWJbunMqAHf5MImuirT6plvs6VSFW62MGHstFQxhBGdEoIhLItH3djc4+FB/OKty3+lL2rGeoFBpVe5g=="
	xeapiAndroidUA    = "NeteaseMusic/9.5.61.260802021928(9005061);Dalvik/2.1.0 (Linux; U; Android 12; HBN-AL00 Build/cd737a2.0)"
	xeapiHost         = "https://interface3.music.163.com"
)

var (
	xeapiStaticKey  = mustHexDecode(xeapiStaticKeyHex)
	xeapiKeyMu      sync.Mutex
	xeapiKeyState   *xeapiKeyInfo
	xeapiSessionID  string
	xeapiSessionKey string
)

type xeapiKeyInfo struct {
	PublicKey string `json:"publicKey"`
	SK        string `json:"sk"`
	Version   string `json:"version"`
	DeviceID  string `json:"deviceId"`
}

func mustHexDecode(s string) []byte {
	v, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return v
}

func xeapiAESEncrypt(key, plaintext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	padding := block.BlockSize() - len(plaintext)%block.BlockSize()
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	out := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += block.BlockSize() {
		block.Encrypt(out[i:i+block.BlockSize()], plaintext[i:i+block.BlockSize()])
	}
	return out
}

func xeapiAESDecrypt(key, ciphertext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil || len(ciphertext) == 0 || len(ciphertext)%block.BlockSize() != 0 {
		return nil
	}
	out := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += block.BlockSize() {
		block.Decrypt(out[i:i+block.BlockSize()], ciphertext[i:i+block.BlockSize()])
	}
	if pad := int(out[len(out)-1]); pad > 0 && pad <= block.BlockSize() {
		out = out[:len(out)-pad]
	}
	return out
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func xeapiSign(timestamp, nonce string) string {
	return base64.StdEncoding.EncodeToString(hmacSHA256([]byte(xeapiSignKey), []byte(timestamp+nonce)))
}

func xeapiMidTransform(in []byte) []byte {
	random := make([]byte, 16)
	_, _ = rand.Read(random)
	xored := make([]byte, len(in))
	for i := range in {
		xored[i] = in[i] ^ random[i&0x0f]
	}
	rot := 0
	if len(xored) > 0 {
		rot = int(random[0]&0x0f) % len(xored)
	}
	out := make([]byte, 0, 16+len(xored))
	out = append(out, random...)
	out = append(out, xored[rot:]...)
	out = append(out, xored[:rot]...)
	return out
}

func deriveX25519AESKey(shared, ephemeralPublic []byte) []byte {
	prk := hmacSHA256(make([]byte, 32), shared)
	return hmacSHA256(prk, append(append([]byte{}, ephemeralPublic...), 1))[:16]
}

func xeapiEncryptS(dynamicKey []byte, state *xeapiKeyInfo, osName string) ([]byte, error) {
	peerRaw, err := base64.StdEncoding.DecodeString(state.PublicKey)
	if err != nil {
		return nil, err
	}
	peer, err := ecdh.X25519().NewPublicKey(peerRaw)
	if err != nil {
		return nil, err
	}
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	shared, err := priv.ECDH(peer)
	if err != nil {
		return nil, err
	}
	ephemeral := priv.PublicKey().Bytes()
	key := deriveX25519AESKey(shared, ephemeral)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, 12)
	_, _ = rand.Read(iv)
	plaintext := []byte(base64.StdEncoding.EncodeToString(dynamicKey) + "|" + osName + "|" + state.SK)
	sealed := gcm.Seal(nil, iv, plaintext, nil)

	out := make([]byte, 0, len(ephemeral)+len(iv)+len(sealed))
	out = append(out, ephemeral...)
	out = append(out, iv...)
	out = append(out, sealed...)
	return out, nil
}

func xeapiDecryptPublicKey(encrypted string) (*xeapiKeyInfo, error) {
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}
	plain := xeapiAESDecrypt(xeapiStaticKey, raw)
	if plain == nil {
		return nil, fmt.Errorf("xeapi public key decrypt failed")
	}
	var state xeapiKeyInfo
	if err := json.Unmarshal(plain, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func xeapiDeviceID() string {
	deviceIdOnce.Do(func() {
		idx := time.Now().Nanosecond() % (len(deviceIds) - 1)
		globalDeviceId = deviceIds[idx]
	})
	return globalDeviceId
}

func fetchXeapiKey(deviceID, currentVersion string) (*xeapiKeyInfo, error) {
	nonce := ""
	for i := range 16 {
		nonce += strconv.Itoa(int(time.Now().UnixNano()>>uint(i)) % 10)
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	form := requests.Datas{
		"appVersion":        "9.5.61",
		"currentKeyVersion": currentVersion,
		"deviceId":          deviceID,
		"nonce":             nonce,
		"os":                "android",
		"requestType":       "active",
		"signature":         xeapiSign(timestamp, nonce),
		"t1":                "",
		"t2":                "",
		"timestamp":         timestamp,
		"uid":               "",
	}

	req := requests.Requests()
	req.Client.Timeout = 15 * time.Second
	req.Header.Set("User-Agent", xeapiAndroidUA)
	if deviceID != "" {
		req.Header.Set("Cookie", "deviceId="+url.QueryEscape(deviceID))
	}
	resp, err := req.Post("https://interface.music.163.com/api/bsr/sk/get", requests.Datas(form))
	if err != nil {
		return nil, err
	}
	body := resp.Content()
	var reply struct {
		Code int `json:"code"`
		Data struct {
			EncryptedData string          `json:"encryptedData"`
			Signature     string          `json:"signature"`
			Timestamp     json.RawMessage `json:"timestamp"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, err
	}
	if reply.Code != 200 || reply.Data.EncryptedData == "" {
		return nil, fmt.Errorf("xeapi key request failed: %s", string(body))
	}
	responseTimestamp := strings.Trim(string(reply.Data.Timestamp), `"`)
	if reply.Data.Signature != "" && xeapiSign(responseTimestamp, nonce) != reply.Data.Signature {
		return nil, fmt.Errorf("xeapi key response signature mismatch")
	}
	state, err := xeapiDecryptPublicKey(reply.Data.EncryptedData)
	if err != nil {
		return nil, err
	}
	if state.SK == "" {
		return nil, fmt.Errorf("xeapi key response missing sk")
	}
	state.DeviceID = deviceID
	return state, nil
}

func getXeapiKey() (*xeapiKeyInfo, error) {
	xeapiKeyMu.Lock()
	defer xeapiKeyMu.Unlock()
	if xeapiKeyState != nil && xeapiKeyState.PublicKey != "" {
		return xeapiKeyState, nil
	}
	deviceID := xeapiDeviceID()
	currentVersion := ""
	if xeapiKeyState != nil {
		currentVersion = xeapiKeyState.Version
	}
	state, err := fetchXeapiKey(deviceID, currentVersion)
	if err != nil {
		return nil, err
	}
	xeapiKeyState = state
	return state, nil
}

func xeapiBuildPlaintext(uri string, data map[string]string, contentType, method string) []byte {
	fields := make(map[string]string)
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded;charset=utf-8"
	}
	mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	if mediaType != "application/x-www-form-urlencoded" {
		fields["contentType"] = contentType
	}
	if method = strings.ToUpper(method); method != "POST" {
		fields["method"] = method
	}
	u, _ := url.Parse(uri)
	if u != nil && u.RawQuery != "" {
		fields["queryString"] = u.RawQuery
	}
	if data != nil {
		bodyData := make(url.Values, len(data))
		for k, v := range data {
			if k == "e_r" {
				continue
			}
			bodyData.Set(k, v)
		}
		body := []byte(bodyData.Encode())
		if len(body) > 0 {
			fields["content"] = string(body)
		}
	}
	if fields["queryString"] != "" {
		fields["queryString"] += "&e_r=true"
	} else {
		fields["queryString"] = "e_r=true"
	}
	out, _ := json.Marshal(fields)
	return out
}

func xeapiEncrypt(uri string, data map[string]string, state *xeapiKeyInfo, contentType, method string) (map[string]string, error) {
	plaintext := xeapiBuildPlaintext(uri, data, contentType, method)
	dynamicKey := make([]byte, 16)
	_, _ = rand.Read(dynamicKey)

	inner := xeapiAESEncrypt(xeapiStaticKey, plaintext)
	middle := xeapiMidTransform(inner)
	c := xeapiAESEncrypt(dynamicKey, middle)

	s, err := xeapiEncryptS(dynamicKey, state, "android")
	if err != nil {
		return nil, err
	}
	sessionID := ""
	if xeapiSessionKey != "" {
		sessionID = xeapiSessionID
	}
	r := xeapiAESEncrypt(xeapiStaticKey, []byte(fmt.Sprintf("%s|%s", state.Version, sessionID)))
	return map[string]string{
		"C": base64.RawURLEncoding.EncodeToString(c),
		"S": base64.StdEncoding.EncodeToString(s),
		"R": base64.StdEncoding.EncodeToString(r),
	}, nil
}

func xeapiDecryptResponse(body []byte) []byte {
	plain := xeapiAESDecrypt([]byte(eapiKey), body)
	if plain == nil {
		return body
	}
	if len(plain) >= 2 && plain[0] == 0x1f && plain[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(plain))
		if err == nil {
			defer zr.Close()
			if out, err := io.ReadAll(zr); err == nil {
				return out
			}
		}
	}
	return plain
}

func xeapiAPIuri(options *Options) string {
	if options.Url != "" {
		return options.Url
	}
	return "/api"
}

func createXeapiRequest(method, rawURL string, data map[string]string, options *Options) (float64, []byte, []*http.Cookie) {
	state, err := getXeapiKey()
	if err != nil {
		return 520, []byte(err.Error()), nil
	}
	uri := xeapiAPIuri(options)
	contentType := "application/x-www-form-urlencoded;charset=utf-8"
	form, err := xeapiEncrypt(uri, data, state, contentType, method)
	if err != nil {
		return 520, []byte(err.Error()), nil
	}

	globalCookieJar = GetGlobalCookieJar()
	req := requests.Requests()
	if len(UNMProxyURL) > 0 {
		req.Proxy(UNMProxyURL)
	}
	if HTTPClientTimeout > 0 {
		req.Client.Timeout = HTTPClientTimeout
	}
	req.Client.Jar = globalCookieJar
	req.Header.Set("User-Agent", xeapiAndroidUA)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Client-Enc-State", "ENCRYPTED")
	req.Header.Set("x-aeapi", "true")
	req.Header.Set("x-os", "android")
	req.Header.Set("x-osver", "16")
	req.Header.Set("x-appver", "9.5.61")
	req.Header.Set("x-buildver", strconv.FormatInt(time.Now().Unix(), 10))

	cookies := append([]*http.Cookie{}, options.Cookies...)
	if u, err := url.Parse(rawURL); err == nil {
		cookies = append(cookies, globalCookieJar.Cookies(u)...)
	}
	deviceID := CookieValueByName(cookies, "deviceId", state.DeviceID)
	sDeviceID := CookieValueByName(cookies, "sDeviceId", deviceID)
	req.Header.Set("x-deviceid", deviceID)
	req.Header.Set("x-sdeviceid", sDeviceID)
	musicU := CookieValueByName(cookies, "MUSIC_U", "")
	if musicU != "" {
		req.Header.Set("x-music-u", musicU)
	}
	if options.CheckToken != "" {
		if token := fetchCheckToken(options.CheckToken); token != "" {
			req.Header.Set("X-antiCheatToken", token)
		}
	}
	if options.Token != "" {
		req.Header.Set("X-antiCheatToken", options.Token)
	}
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.SetCookie(cookie)
	}

	resp, err := req.Post(xeapiHost+"/xeapi"+strings.TrimPrefix(uri, "/api"), requests.Datas(form))
	if err != nil {
		return 520, []byte(err.Error()), nil
	}
	if resp.R != nil {
		if ssid := resp.R.Header.Get("x-encr-ssid"); ssid != "" {
			xeapiSessionID = ssid
		}
		if sskey := resp.R.Header.Get("x-encr-sskey"); sskey != "" {
			xeapiSessionKey = sskey
		}
	}
	body := xeapiDecryptResponse(resp.Content())
	code := float64(resp.R.StatusCode)
	var parsed map[string]any
	if json.Unmarshal(body, &parsed) == nil {
		if v, ok := parsed["code"].(float64); ok {
			code = v
		}
	}
	return code, body, resp.Cookies()
}
