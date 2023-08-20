package kuwo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

type KuWo struct{}

const (
	SearchSongURL = "http://kuwo.cn/api/www/search/searchMusicBykeyWord?key=%s&pn=1&rn=30"
)

var blockSongUrl = map[string]json.Number{
	"2914632520.mp3": "7",
}

var lock sync.Mutex

func (m *KuWo) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song = base.PreSearchSong(song)
	keyWordList := utils.Combination(song.ArtistList)
	wg := sync.WaitGroup{}
	for _, v := range keyWordList {
		wg.Add(1)
		// use goroutine to deal multiple request
		go func(word string) {
			defer wg.Done()
			keyWord := song.Name
			if len(word) != 0 {
				keyWord = fmt.Sprintf("%s %s", song.Name, word)
			}
			// key, value := getTokenInfo(keyWord)
			header := make(http.Header, 4)
			header["referer"] = append(header["referer"], "http://www.kuwo.cn/search/list?key="+url.QueryEscape(keyWord))
			header["cookie"] = append(header["cookie"], "Hm_Iuvt_cdb524f42f0cer9b268e4v7y734w5esq24=fppPsMdS6XRFXthpMGPXycmTsm3Ny8xr")
			header["secret"] = append(header["secret"], "46c96f9b8ab640394da88016b1ed67db57a66c474e041c18b0283cdd53ce101d0465c8a8")
			searchUrl := fmt.Sprintf(SearchSongURL, url.QueryEscape(keyWord))
			result, err := base.Fetch(searchUrl, nil, header, true)
			if err != nil {
				log.Println(err)
				return
			}
			data, ok := result["data"].(common.MapType)
			if ok {
				list, ok := data["list"].([]interface{})
				if ok && len(list) > 0 {
					listLength := len(list)
					maxIndex := listLength/2 + 1
					if maxIndex > 5 {
						maxIndex = 5
					}
					for index, matched := range list {
						if index >= maxIndex { //kuwo list order by score default
							break
						}
						kuWoSong, ok := matched.(common.MapType)
						if ok {
							rid, ok := kuWoSong["rid"].(json.Number)
							rids := ""
							if !ok {
								rids, ok = kuWoSong["rid"].(string)
							} else {
								rids = rid.String()
							}
							if ok {
								songResult := &common.Song{}
								singerName := html.UnescapeString(kuWoSong["artist"].(string))
								songName := html.UnescapeString(kuWoSong["name"].(string))
								//musicSlice := strings.Split(musicrid, "_")
								//musicId := musicSlice[len(musicSlice)-1]
								songResult.PlatformUniqueKey = kuWoSong
								songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword
								songResult.Source = "kuwo"
								songResult.PlatformUniqueKey["header"] = header
								songResult.PlatformUniqueKey["musicId"] = rids
								songResult.Id = rids
								if len(songResult.Id) > 0 {
									songResult.Id = string(common.KuWoTag) + songResult.Id
								}
								songResult.Name = songName
								songResult.Artist = singerName
								songResult.AlbumName = html.UnescapeString(kuWoSong["album"].(string))
								songResult.Artist = strings.ReplaceAll(singerName, " ", "")
								songResult.MatchScore, ok = base.CalScore(song, songName, singerName, index, maxIndex)
								if !ok {
									continue
								}
								// protect slice thread safe
								lock.Lock()
								songs = append(songs, songResult)
								lock.Unlock()
							}
						}
					}
				}
			}
		}(v)
	}
	wg.Wait()
	return base.AfterSearchSong(song, songs)
}
func (m *KuWo) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
	if id, ok := song.PlatformUniqueKey["musicId"]; ok {
		if musicId, ok := id.(string); ok {
			if httpHeader, ok := song.PlatformUniqueKey["header"]; ok {
				if header, ok := httpHeader.(http.Header); ok {
					header["user-agent"] = append(header["user-agent"], "okhttp/3.10.0")
					format := "flac|mp3"
					br := ""
					switch searchSong.Quality {
					case common.Standard:
						format = "mp3"
						br = "&br=128kmp3"
					case common.Higher:
						format = "mp3"
						br = "&br=192kmp3"
					case common.ExHigh:
						format = "mp3"
					case common.Lossless:
						format = "flac|mp3"
					default:
						format = "flac|mp3"
					}

					clientRequest := network.ClientRequest{
						Method:               http.MethodGet,
						ForbiddenEncodeQuery: true,
						RemoteUrl:            "http://mobi.kuwo.cn/mobi.s?f=kuwo&q=" + base64.StdEncoding.EncodeToString(Encrypt([]byte("corp=kuwo&p2p=1&type=convert_url2&sig=0&format="+format+"&rid="+musicId+br))),
						Header:               header,
						Proxy:                true,
					}
					resp, err := network.Request(&clientRequest)
					if err != nil {
						log.Println(err)
						return song
					}
					defer resp.Body.Close()
					body, err := network.GetResponseBody(resp, false)
					reg := regexp.MustCompile(`http[^\s$"]+`)
					address := string(body)
					params := reg.FindStringSubmatch(address)
					if len(params) > 0 {
						if duration, ok := blockSongUrl[filepath.Base(params[0])]; ok && song.PlatformUniqueKey["duration"].(json.Number) == duration {
							log.Println(song.PlatformUniqueKey["UnKeyWord"].(string) + "，该歌曲酷我版权保护")
							return song
						}
						song.Url = params[0]
						return song
					}

				}
			}
		}
	}
	return song
}

func (m *KuWo) ParseSong(searchSong common.SearchSong) *common.Song {
	song := &common.Song{}
	songs := m.SearchSong(searchSong)
	if len(songs) > 0 {
		song = m.GetSongUrl(common.SearchMusic{Quality: searchSong.Quality}, songs[0])
	}
	return song
}

// func getTokenInfo(keyword string) (string, string) {
// 	clientRequest := network.ClientRequest{
// 		Method:    http.MethodGet,
// 		RemoteUrl: "http://kuwo.cn/search/list?key=" + keyword,
// 		Host:      "kuwo.cn",
// 		Header:    nil,
// 		Proxy:     false,
// 	}
// 	resp, err := network.Request(&clientRequest)
// 	if err != nil {
// 		log.Println(err)
// 		return "", ""
// 	}
// 	defer resp.Body.Close()

// 	for _, v := range resp.Header.Values("set-cookie") {
// 		if !strings.HasPrefix(v, "Hm_") {
// 			continue
// 		}
// 		v = utils.ReplaceAll(v, ";.*", "")
// 		res := strings.Split(v, "=")
// 		if len(res) >= 2 {
// 			return res[0], res[1]
// 		}
// 	}
// 	return "", ""
// }

// func genSecret(key, value string) (secret string) {
// 	if key == "" {
// 		return
// 	}
// 	var n string
// 	for i := 0; i < len(key); i++ {
// 		n += strconv.Itoa(int(key[i]))
// 	}
// 	var (
// 		r = int(math.Floor(float64(len(n)) / 5))
// 		o int64
// 		l = int64(math.Ceil(float64(len(key)) / 2))
// 		c = int64(math.Pow(2, 31) - 1)
// 	)
// 	if r*5 >= len(n) {
// 		o, _ = strconv.ParseInt(string([]byte{n[r], n[2*r], n[3*r], n[4*r]}), 10, 64)
// 	} else {
// 		o, _ = strconv.ParseInt(string([]byte{n[r], n[2*r], n[3*r], n[4*r], n[5*r]}), 10, 64)
// 	}

// 	if o < 2 {
// 		return
// 	}
// 	var (
// 		d    = int(math.Round(1e9*rand.Float64())) % 1e8
// 		dStr = strconv.Itoa(d)
// 		nNum int64
// 	)
// 	for n += dStr; len(n) > 10; {
// 		a, _ := new(big.Int).SetString(trimInvalidIntChar(n[:10]), 10)
// 		b, _ := new(big.Int).SetString(trimInvalidIntChar(n[10:]), 10)
// 		fmt.Println(n, a, b)
// 		if len(n)-10 >= 22 {
// 			n = new(big.Float).SetInt(a.Add(a, b)).Text('e', 16)
// 		} else {
// 			n = a.Add(a, b).String()
// 		}
// 	}
// 	nNum, _ = strconv.ParseInt(n, 10, 64)
// 	nNum = (o*nNum + l) % c

// 	var f string
// 	for i := 0; i < len(value); i++ {
// 		h := int64(value[i]) ^ int64(math.Floor(float64(nNum)/float64(c)*255))
// 		if h < 16 {
// 			f += "0" + strconv.FormatInt(h, 16)
// 		} else {
// 			f += strconv.FormatInt(h, 16)
// 		}
// 		nNum = (o*nNum + l) % c
// 	}
// 	for dStr = strconv.FormatInt(int64(d), 16); len(dStr) < 8; {
// 		dStr = "0" + dStr
// 	}
// 	secret = f + dStr
// 	return
// }

// func trimInvalidIntChar(s string) string {
// 	var i int
// 	for i = 0; i < len(s); i++ {
// 		if s[i] < '0' || s[i] > '9' {
// 			break
// 		}
// 	}
// 	return s[:i]
// }
