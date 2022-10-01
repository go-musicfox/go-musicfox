package migu

import (
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/processor/crypto"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

type Migu struct{}

var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC8asrfSaoOb4je+DSmKdriQJKW
VJ2oDZrs3wi5W67m3LwTB9QVR+cE3XWU21Nx+YBxS0yun8wDcjgQvYt625ZCcgin
2ro/eOkNyUOTBIbuj9CvMnhUYiR61lC1f1IGbrSYYimqBVSjpifVufxtx/I3exRe
ZosTByYp4Xwpb1+WAQIDAQAB
-----END PUBLIC KEY-----
`)
var rsaPublicKey *rsa.PublicKey

func getRsaPublicKey() (*rsa.PublicKey, error) {
	var err error = nil
	if rsaPublicKey == nil {
		rsaPublicKey, err = crypto.ParsePublicKey(publicKey)
	}
	return rsaPublicKey, err
}

func (m *Migu) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song = base.PreSearchSong(song)
	result, err := base.Fetch(
		"http://m.music.migu.cn/migu/remoting/scr_search_tag?keyword="+song.Keyword+"&type=2&rows=20&pgc=1",
		nil, nil, false)
	if err != nil {
		log.Println(err)
		return songs
	}
	data, ok := result["musics"].(common.SliceType)
	if ok {
		list := data
		if ok && len(list) > 0 {
			listLength := len(list)
			maxIndex := listLength/2 + 1
			if maxIndex > 10 {
				maxIndex = 10
			}
			for index, matched := range list {
				miguSong, ok := matched.(common.MapType)
				if ok {
					_, ok := miguSong["copyrightId"].(string)
					if ok {
						if index >= maxIndex {
							break
						}
						songResult := &common.Song{}
						singerName, _ := miguSong["singerName"].(string)
						songName, _ := miguSong["songName"].(string)
						songResult.PlatformUniqueKey = miguSong
						songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword
						songResult.Source = "migu"
						songResult.Id, ok = miguSong["id"].(string)
						if len(songResult.Id) > 0 {
							songResult.Id = string(common.MiGuTag) + songResult.Id
						}
						songResult.Name = songName
						songResult.Artist = singerName
						songResult.AlbumName, _ = miguSong["albumName"].(string)
						songResult.Artist = strings.ReplaceAll(singerName, " ", "")
						songResult.MatchScore, ok = base.CalScore(song, songName, singerName, index, maxIndex)
						if !ok {
							continue
						}
						songs = append(songs, songResult)

					}

				}

			}

		}
	}
	return base.AfterSearchSong(song, songs)
}
func (m *Migu) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
	if cId, ok := song.PlatformUniqueKey["copyrightId"]; ok {
		if copyrightId := cId.(string); ok {
			type MiguFormat struct {
				CopyrightId string `json:"copyrightId"`
				Type        int    `json:"type"`
			}
			formatType := 1
			switch searchSong.Quality {
			case common.Standard:
				formatType = 1
			case common.Higher:
				formatType = 1
			case common.ExHigh:
				formatType = 2
			case common.Lossless:
				formatType = 3
			default:
				formatType = 3
			}
			en := utils.ToJson(&MiguFormat{CopyrightId: copyrightId, Type: formatType})
			if len(copyrightId) > 0 {
				header := make(http.Header, 2)
				header["origin"] = append(header["origin"], "http://music.migu.cn/")
				header["referer"] = append(header["referer"], "http://music.migu.cn/")
				clientRequest := network.ClientRequest{
					Method:               http.MethodGet,
					RemoteUrl:            "http://music.migu.cn/v3/api/music/audioPlayer/getPlayInfo?dataType=2&" + encrypt(en),
					//Host:                 "music.migu.cn",
					Header:               header,
					Proxy:                true,
					ForbiddenEncodeQuery: true, //dataType first must
				}
				resp, err := network.Request(&clientRequest)
				if err != nil {
					log.Println(err)
					return song
				}
				defer resp.Body.Close()
				body, _ := network.StealResponseBody(resp)
				//data := utils.ParseJsonV2(body)
				type MiguResult struct {
					PlayUrl string
					//FormatId string
					//SalePrice string
					//BizType string
					//BizCode string
					//AuditionsLength int64
				}
				type MiguResponse struct {
					ReturnCode string
					Msg        string
					Data       *MiguResult
				}
				miguResponse := &MiguResponse{}
				err = utils.ParseJsonV4(body, miguResponse)
				if err != nil {
					log.Println(err)
				} else if miguResponse.Data != nil {
					if strings.Index(miguResponse.Data.PlayUrl, "http") == 0 {
						song.Url = miguResponse.Data.PlayUrl
						return song
					} else if strings.Index(miguResponse.Data.PlayUrl, "//") == 0 {
						song.Url = "http:" + miguResponse.Data.PlayUrl
						return song
					}

				}
			}
		}
	}

	return song
}
func (m *Migu) ParseSong(searchSong common.SearchSong) *common.Song {
	song := &common.Song{}
	songs := m.SearchSong(searchSong)
	if len(songs) > 0 {
		song = m.GetSongUrl(common.SearchMusic{Quality: searchSong.Quality}, songs[0])
	}
	return song
}
func encrypt(text string) string {
	encryptedData := ""
	//log.Println(text)
	text = utils.ToJson(utils.ParseJson(bytes.NewBufferString(text).Bytes()))
	randomBytes, err := utils.GenRandomBytes(32)
	if err != nil {
		log.Println(err)
		return encryptedData
	}
	pwd := bytes.NewBufferString(hex.EncodeToString(randomBytes)).Bytes()
	salt, err := utils.GenRandomBytes(8)
	if err != nil {
		log.Println(err)
		return encryptedData
	}
	//key = []byte{0xaf, 0xb3, 0xac, 0x50, 0xcd, 0x1d, 0x23, 0x81, 0x58, 0x5f, 0xa7, 0xbc, 0xbd, 0x8c, 0xbe, 0x02, 0x56, 0x0f, 0xad, 0xe7, 0xd1, 0x7e, 0x2e, 0xb1, 0x14, 0x81, 0x6f, 0x27, 0xab, 0x7b, 0x6a, 0x75}
	//iv = []byte{0xfb, 0x10, 0x89, 0xb0, 0x13, 0x32, 0xf2, 0xa7, 0x02, 0x51, 0x49, 0xff, 0xbc, 0x16, 0xf0, 0x40}
	//pwd = bytes.NewBufferString("d8e28215ed6573e0fd5eb8b8ae8062542589e96f669bee6503af003c63cdfbd4").Bytes()
	//salt = []byte{0xde, 0xfc, 0x9f, 0x26, 0x29, 0xdd, 0xec, 0x37}
	key, iv := derive(pwd, salt, 256, 16)
	var data []byte
	data = append(data, bytes.NewBufferString("Salted__").Bytes()...)
	data = append(data, salt...)
	encryptedD := crypto.AesEncryptCBCWithIv(bytes.NewBufferString(text).Bytes(), key, iv)
	data = append(data, encryptedD...)
	dat := base64.StdEncoding.EncodeToString(data)
	var rsaB []byte
	pubKey, err := getRsaPublicKey()
	if err == nil {
		rsaB = crypto.RSAEncryptV2(pwd, pubKey)
	} else {
		rsaB = crypto.RSAEncrypt(pwd, publicKey)
	}
	sec := base64.StdEncoding.EncodeToString(rsaB)
	//log.Println("data:", dat)
	//log.Println("sec:", sec)
	encryptedData = "data=" + url.QueryEscape(dat)
	encryptedData = encryptedData + "&secKey=" + url.QueryEscape(sec)
	return encryptedData
}
func derive(password []byte, salt []byte, keyLength int, ivSize int) ([]byte, []byte) {
	keySize := keyLength / 8
	repeat := math.Ceil(float64(keySize+ivSize*8) / 32)
	var data []byte
	var lastData []byte
	for i := 0.0; i < repeat; i++ {
		var md5Data []byte
		md5Data = append(md5Data, lastData...)
		md5Data = append(md5Data, password...)
		md5Data = append(md5Data, salt...)
		h := md5.New()
		h.Write(md5Data)
		md5Data = h.Sum(nil)
		data = append(data, md5Data...)
		lastData = md5Data
	}
	return data[:keySize], data[keySize : keySize+ivSize]
}
