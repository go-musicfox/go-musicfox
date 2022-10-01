package qq

import (
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/config"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
	"github.com/mitchellh/mapstructure"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type QQ struct{}

type typeSong struct {
	Album struct {
		Name string
	}
	File struct {
		Media_Mid string
	}
	Mid    string
	Name   string
	Singer []struct {
		Name string
	}
}

type typeSongResult struct {
	Code int
	Data struct {
		Sip        []string
		Midurlinfo []struct {
			Purl string
		}
	}
}

type getSongConfig struct {
	fmid    string
	mid     string
	cookies []*http.Cookie
	song    *common.Song
	br      string
	format  string
}

func (m *QQ) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song = base.PreSearchSong(song)
	cookies := getCookies()
	result, err := base.Fetch(
		"https://c.y.qq.com/soso/fcgi-bin/client_search_cp?"+
			"ct=24&qqmusic_ver=1298&new_json=1&remoteplace=txt.yqq.center&"+
			"t=0&aggr=1&cr=1&catZhida=1&lossless=0&flag_qc=0&p=1&n=20&w="+
			song.Keyword+
			"&"+
			"g_tk=5381&loginUin=0&hostUin=0&"+
			"format=json&inCharset=utf8&outCharset=utf-8&notice=0&platform=yqq&needNewCode=0",
		cookies, nil, true)
	if err != nil {
		log.Println(err)
		return songs
	}
	data := result["data"]
	if data != nil {
		if dMap, ok := data.(common.MapType); ok {
			if dSong, ok := dMap["song"].(common.MapType); ok {
				if list, ok := dSong["list"].(common.SliceType); ok {
					if ok && len(list) > 0 {
						listLength := len(list)
						maxIndex := listLength/2 + 1
						if maxIndex > 10 {
							maxIndex = 10
						}
						for index, matched := range list {
							if index >= maxIndex {
								break
							}

							qqSong := &typeSong{}
							if err = mapstructure.Decode(matched, &qqSong); err == nil {
								artists := make([]string, 2)
								for _, singer := range qqSong.Singer {
									artists = append(artists, singer.Name)
								}
								songResult := &common.Song{}
								songResult.PlatformUniqueKey = matched.(common.MapType)
								songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword
								songResult.PlatformUniqueKey["Mid"] = qqSong.File.Media_Mid
								songResult.PlatformUniqueKey["MusicId"] = qqSong.Mid
								songResult.Source = "qq"
								songResult.Name = qqSong.Name
								songResult.Artist = strings.Join(artists, " & ")
								songResult.AlbumName = qqSong.Album.Name
								songResult.Id = string(common.QQTag) + qqSong.Mid
								songResult.MatchScore, ok = base.CalScore(song, qqSong.Name, songResult.Artist, index, maxIndex)
								if !ok {
									continue
								}
								songs = append(songs, songResult)
							}
						}
					}
				}
			}
		}
	}
	return base.AfterSearchSong(song, songs)
}

func (m *QQ) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
	if fmid, ok := song.PlatformUniqueKey["Mid"].(string); ok {
		if mid, ok := song.PlatformUniqueKey["MusicId"].(string); ok {
			cookies := getCookies()
			if cookies == nil {
				format := "mp3"
				br := "M800"
				searchSong.Quality = common.Standard
				conf := &getSongConfig{
					fmid,
					mid,
					nil,
					song,
					br,
					format,
				}
				if gotSong := getSong(conf); gotSong != nil {
					song = gotSong
				}
			} else {
				wg := sync.WaitGroup{}
				wg.Add(3)
				rand.Seed(time.Now().UnixNano())
				songCh := make(chan *common.Song, 3)
				for _, quality := range []map[string]string{{"M500": "mp3"}, {"M800": "mp3"}, {"F000": "flac"}} {
					for br, format := range quality {
						conf := &getSongConfig{
							fmid,
							mid,
							cookies,
							song,
							br,
							format,
						}
						go func(conf *getSongConfig) {
							if gotSong := getSong(conf); gotSong != nil {
								gotSong.PlatformUniqueKey["Quality"] = conf.br
								songCh <- gotSong
							}
							wg.Done()
						}(conf)
					}
				}
				wg.Wait()
				songs := make(map[string]*common.Song, 3)
				for gotSong := range songCh {
					songs[gotSong.PlatformUniqueKey["Quality"].(string)] = gotSong
					if len(songCh) == 0 {
						break
					}
				}
				quality := "M500"
				finished := false
				for !finished {
					switch searchSong.Quality {
					case common.Standard:
						quality = "M500"
					case common.Higher:
						fallthrough
					case common.ExHigh:
						quality = "M800"
					case common.Lossless:
						quality = "F000"
					default:
						quality = "M500"
					}
					if gotSong, ok := songs[quality]; ok {
						song = gotSong
						finished = true
					}
					searchSong.Quality--
				}
			}
		}
	}
	return song
}

func getSong(config *getSongConfig) *common.Song {
	guid := utils.ToFixed(rand.Float64()*10000000, 0)
	rawQueryData := `{"req_0":{"module":"vkey.GetVkeyServer","method":"CgiGetVkey","param":{"guid":"` +
		strconv.Itoa(int(guid)) +
		`","loginflag":1,"filename":["` + config.br + config.fmid + "." + config.format +
		`"],"songmid":["` + config.mid + `"],"songtype":[0],"uin":"0","platform":"20"}}}`
	clientRequest := network.ClientRequest{
		Method:               http.MethodGet,
		ForbiddenEncodeQuery: true,
		RemoteUrl:            "https://u.y.qq.com/cgi-bin/musicu.fcg?data=" + url.QueryEscape(rawQueryData),
		Proxy:                true,
		Cookies:              config.cookies,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := network.StealResponseBody(resp)
	songData := utils.ParseJsonV2(body)
	songResult := &typeSongResult{}
	if err = mapstructure.Decode(songData["req_0"], &songResult); err == nil {
		if songResult.Data.Midurlinfo[0].Purl != "" {
			config.song.Url = songResult.Data.Sip[0] + songResult.Data.Midurlinfo[0].Purl
			return config.song
		} else {
			log.Println(config.song.PlatformUniqueKey["UnKeyWord"].(string) + "，该歌曲QQ音乐版权保护")
			// log.Println(utils.ToJson(songData))
		}
	}

	return nil
}

func (m *QQ) ParseSong(searchSong common.SearchSong) *common.Song {
	song := &common.Song{}
	songs := m.SearchSong(searchSong)
	if len(songs) > 0 {
		song = m.GetSongUrl(common.SearchMusic{Quality: searchSong.Quality}, songs[0])
	}
	return song
}

func getCookies() []*http.Cookie {
	if _, err := os.Stat(*config.QQCookieFile); os.IsNotExist(err) {
		return nil
	}
	return utils.ParseCookies(*config.QQCookieFile)
}
