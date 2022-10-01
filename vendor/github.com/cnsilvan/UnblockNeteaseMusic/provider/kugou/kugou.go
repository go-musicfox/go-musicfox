package kugou

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

const (
	APIGetSongURL = "http://trackercdn.kugou.com/i/v2/?"
	SearchSongURL = "http://mobilecdn.kugou.com/api/v3/search/song?keyword=%s&page=1&pagesize=10"
)

type KuGou struct{}

var lock sync.Mutex

func (m *KuGou) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song = base.PreSearchSong(song)
	cookies := getCookies()
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
			searchUrl := fmt.Sprintf(SearchSongURL, keyWord)
			result, err := base.Fetch(searchUrl, cookies, nil, true)
			if err != nil {
				log.Println(err)
				return
			}
			if listSlice, ok := transSearchResponse(result); ok {
				listLength := len(listSlice)
				if listLength > 0 {
					maxIndex := listLength/2 + 1
					if maxIndex > 5 {
						maxIndex = 5
					}
					for index, matched := range listSlice {
						if index >= maxIndex {
							break
						}
						if kugouSong, ok := matched.(common.MapType); ok {
							if _, ok := kugouSong["hash"].(string); ok {
								songResult := &common.Song{}
								singerName, _ := kugouSong["singername"].(string)
								songName, _ := kugouSong["songname"].(string)
								songResult.PlatformUniqueKey = kugouSong
								songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword
								songResult.Source = "kugou"
								songResult.Name = songName
								songResult.Artist = singerName
								songResult.Artist = strings.ReplaceAll(singerName, " ", "")
								songResult.AlbumName, _ = kugouSong["album_name"].(string)
								audioId, ok := kugouSong["audio_id"].(json.Number)
								songResult.Id = audioId.String()
								if ok && len(songResult.Id) > 0 {
									songResult.Id = string(common.KuGouTag) + songResult.Id
								}
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
func (m *KuGou) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
	hashKey := "hash"
	switch searchSong.Quality {
	case common.Standard:
		hashKey = "hash"
	case common.Higher:
		hashKey = "hash"
	case common.ExHigh:
		hashKey = "320hash"
	case common.Lossless:
		hashKey = "sqhash"
	default:
		hashKey = "hash"
	}
	fileHash, ok := song.PlatformUniqueKey[hashKey].(string)
	if !ok || fileHash == "" {
		fileHash, ok = song.PlatformUniqueKey["hash"].(string)
	}
	albumId, ok := song.PlatformUniqueKey["album_id"].(string)
	if ok && len(fileHash) > 0 {
		clientRequest := network.ClientRequest{
			Method: http.MethodGet,
			RemoteUrl: APIGetSongURL + "key=" + utils.MD5([]byte(fileHash+"kgcloudv2")) + "&hash=" +
				fileHash + "&appid=1005&pid=2&cmd=25&behavior=play&album_id=" + albumId,
			//Host:      "trackercdnbj.kugou.com",
			//Cookies:              cookies,
			Header:               nil,
			ForbiddenEncodeQuery: true,
			Proxy:                false,
		}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			log.Println(err)
			return song
		}
		defer resp.Body.Close()
		body, err := network.StealResponseBody(resp)
		songData := utils.ParseJsonV2(body)
		status, ok := songData["status"].(json.Number)
		if !ok || status.String() != "1" {
			log.Println(song.PlatformUniqueKey["UnKeyWord"].(string) + "，该歌曲酷狗版权保护")
			//log.Println(utils.ToJson(songData))
			return song
		}
		songUrls, ok := songData["url"].(common.SliceType)
		if ok && len(songUrls) > 0 {
			songUrl, ok := songUrls[0].(string)
			if ok && strings.Index(songUrl, "http") == 0 {
				song.Url = songUrl
				if br, ok := songData["bitRate"]; ok {
					switch br.(type) {
					case json.Number:
						song.Br, _ = strconv.Atoi(br.(json.Number).String())
					case int:
						song.Br = br.(int)
					}
				}
				return song

			}
		}
	}
	return song
}

func (m *KuGou) ParseSong(searchSong common.SearchSong) *common.Song {
	song := &common.Song{}
	songs := m.SearchSong(searchSong)
	if len(songs) > 0 {
		song = m.GetSongUrl(common.SearchMusic{Quality: searchSong.Quality}, songs[0])
	}
	return song
}

func getCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 1)
	cookie := &http.Cookie{Name: "kg_mid", Value: createGuid(), Path: "kugou.com", Domain: "kugou.com"}
	cookies[0] = cookie
	return cookies
}
func createGuid() string {
	guid := s4() + s4() + "-" + s4() + "-" + s4() + "-" + s4() + "-" + s4() + s4() + s4()
	return utils.MD5(bytes.NewBufferString(guid).Bytes())
}
func s4() string {
	num := uint64((1 + common.Rand.Float64()) * 0x10000)
	num = num | 0
	return strconv.FormatUint(num, 16)[1:]
}

func transSearchResponse(obj map[string]interface{}) (common.SliceType, bool) {
	data := obj["data"]
	if data != nil {
		if dMap, ok := data.(common.MapType); ok {
			if lists, ok := dMap["info"]; ok {
				if listSlice, ok := lists.(common.SliceType); ok {
					return listSlice, true
				}
			}
		}
	}
	return nil, false
}
