package provider

import (
	"fmt"
	"log"

	"github.com/cnsilvan/UnblockNeteaseMusic/cache"
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	kugou "github.com/cnsilvan/UnblockNeteaseMusic/provider/kugou"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/kuwo"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/migu"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/qq"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

type Provider interface {
	SearchSong(song common.SearchSong) (songs []*common.Song)
	GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song
	//search&get url
	ParseSong(searchSong common.SearchSong) *common.Song
}

var providers map[string]Provider

func Init() {
	providers = make(map[string]Provider)
	for _, source := range common.Source {
		providers[source] = NewProvider(source)
	}
}
func NewProvider(kind string) Provider {
	switch kind {
	case "kuwo":
		return &kuwo.KuWo{}
	case "kugou":
		return &kugou.KuGou{}
	case "migu":
		return &migu.Migu{}
	case "qq":
		return &qq.QQ{}
	default:
		return &kuwo.KuWo{}
	}
}
func GetProvider(kind string) Provider {
	if p, ok := providers[kind]; ok {
		return p
	}
	return NewProvider(kind)

}

func UpdateCacheMd5(music common.SearchMusic, md5 string) {
	if song, ok := cache.GetSong(music); ok {
		song.Md5 = md5
		cache.PutSong(music, song)
	}
}
func Find(music common.SearchMusic) common.Song {
	log.Println(fmt.Sprintf("find song info :%+v", music))
	if song, ok := cache.GetSong(music); ok {
		if len(song.Url) > 0 {
			log.Println("hit cache:", utils.ToJson(song))
			if checkCache(song) {
				return *song
			} else if strings.Index(music.Id, string(common.StartTag)) == 0 {
				log.Println("but cache invalid")
			} else {
				cache.Delete(music)
				log.Println("but cache invalid")
			}
		}
		//log.Println("search:", utils.ToJson(song))
		if strings.Index(music.Id, string(common.StartTag)) == 0 {
			now := time.Now()
			var re *common.Song
			if strings.Index(music.Id, string(common.KuGouTag)) == 0 {
				re = calculateSongInfo(GetProvider("kugou").GetSongUrl(music, song))
			} else if strings.Index(music.Id, string(common.KuWoTag)) == 0 {
				re = calculateSongInfo(GetProvider("kuwo").GetSongUrl(music, song))
			} else if strings.Index(music.Id, string(common.MiGuTag)) == 0 {
				re = calculateSongInfo(GetProvider("migu").GetSongUrl(music, song))
			} else if strings.Index(music.Id, string(common.QQTag)) == 0 {
				re = calculateSongInfo(GetProvider("qq").GetSongUrl(music, song))
			} else {

			}
			log.Println("consumed:", time.Since(now))
			log.Println(utils.ToJson(re))
			if re != nil && len(re.Url) > 0 {
				cache.PutSong(music, re)
				return *re
			}
		}
	}
	var songT common.Song
	songT.Id = music.Id
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "https://" + common.HostDomain["music.163.com"] + "/api/song/detail?ids=[" + songT.Id + "]",
		Host:      "music.163.com",
		Header:    nil,
		Proxy:     false,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		return songT
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err2 := network.StealResponseBody(resp)
		if err2 != nil {
			log.Println("GetResponseBody fail")
			return songT
		}
		oJson := utils.ParseJsonV2(body)
		if songs, ok := oJson["songs"].(common.SliceType); ok && len(songs) > 0 {
			song := songs[0]
			var searchSong = make(common.MapType, 8)
			searchSong["songId"] = songT.Id
			var artists []string
			switch song.(type) {
			case common.MapType:
				neteaseSong := song.(common.MapType)
				searchSong["id"] = neteaseSong["id"]
				searchSong["name"] = neteaseSong["name"]
				searchSong["alias"] = neteaseSong["alias"]
				searchSong["duration"] = neteaseSong["duration"]
				searchSong["album"] = make(common.MapType, 2)
				searchSong["album"].(common.MapType)["id"], ok = neteaseSong["album"].(common.MapType)["id"]
				searchSong["album"].(common.MapType)["name"], ok = neteaseSong["album"].(common.MapType)["name"]
				switch neteaseSong["artists"].(type) {
				case common.SliceType:
					length := len(neteaseSong["artists"].(common.SliceType))
					searchSong["artists"] = make(common.SliceType, length)
					artists = make([]string, length)
					for index, value := range neteaseSong["artists"].(common.SliceType) {
						if searchSong["artists"].(common.SliceType)[index] == nil {
							searchSong["artists"].(common.SliceType)[index] = make(common.MapType, 2)
						}
						searchSong["artists"].(common.SliceType)[index].(common.MapType)["id"], ok = value.(common.MapType)["id"]
						searchSong["artists"].(common.SliceType)[index].(common.MapType)["name"], ok = value.(common.MapType)["name"]
						artists[index], ok = value.(common.MapType)["name"].(string)
					}

				}
			default:

			}
			//if searchSong["name"] != nil {
			//	searchSong["name"] = utils.ReplaceAll(searchSong["name"].(string), `\s*cover[:：\s][^）]+）`, "")
			//	searchSong["name"] = utils.ReplaceAll(searchSong["name"].(string), `(\s*cover[:：\s][^\)]+)`, "")
			//}
			searchSong["artistsName"] = strings.Join(artists, " ")
			searchSong["artistList"] = artists
			searchSong["keyword"] = searchSong["name"].(string) + " " + searchSong["artistsName"].(string)
			log.Println("search song:" + searchSong["keyword"].(string))
			songT = *parseSongFn(searchSong, music)
			log.Println(utils.ToJson(songT))
			return songT

		} else {
			return songT
		}
	} else {
		return songT
	}
}

func parseSongFn(key common.MapType, music common.SearchMusic) *common.Song {
	id := "0"
	searchSongName := key["name"].(string)
	searchSongName = strings.ToUpper(searchSongName)
	searchArtistsName := key["artistsName"].(string)
	searchArtistsName = strings.ToUpper(searchArtistsName)
	searchKeyword := key["keyword"].(string)
	if songId, ok := key["songId"]; ok {
		id = songId.(string)
	}
	searchArtistList := key["artistList"].([]string)
	key["musicQuality"] = music.Quality
	var ch = make(chan *common.Song)
	now := time.Now()
	searchSong := common.SearchSong{
		Keyword: searchKeyword, Name: searchSongName,
		ArtistsName: searchArtistsName, Quality: music.Quality,
		OrderBy: common.MatchedScoreDesc, Limit: 1,
		ArtistList: searchArtistList,
	}
	songs := getSongFromAllSource(searchSong, ch)
	log.Println("consumed:", time.Since(now))
	result := &common.Song{}
	result.Size = 0
	for _, song := range songs {

		if song.MatchScore > result.MatchScore {
			result = song
		} else if song.MatchScore == result.MatchScore && song.Size > result.Size {
			result = song
		}
	}

	if id != "0" {
		result.Id = id
		if len(result.Url) > 0 {
			result.PlatformUniqueKey = nil
			cache.PutSong(music, result)
		}
	}
	return result

}

func getSongFromAllSource(key common.SearchSong, ch chan *common.Song) []*common.Song {
	var songs []*common.Song
	sum := 0
	for _, p := range providers {
		pt := p
		go utils.PanicWrapper(func() {
			ch <- calculateSongInfo(pt.ParseSong(key))
		})
		sum++
	}
	for {
		select {
		case song, _ := <-ch:
			if len(song.Url) > 0 {
				songs = append(songs, song)
			}
			sum--
			if sum <= 0 {
				return songs
			}
		case <-time.After(time.Second * 6):
			return songs
		}
	}
}

func SearchSongFromAllSource(key common.SearchSong) []*common.Song {
	var songs []*common.Song
	ch := make(chan []*common.Song)
	sum := 0
	for _, p := range providers {
		pt := p
		go utils.PanicWrapper(func() {
			ch <- pt.SearchSong(key)
		})
		sum++
	}
	for {
		select {
		case song, _ := <-ch:
			songs = append(songs, song...)
			sum--
			if sum <= 0 {
				return songs
			}
		case <-time.After(time.Second * 6):
			return songs
		}
	}
}
func calculateSongInfo(song *common.Song) *common.Song {
	if len(song.Url) > 0 {
		if len(song.Md5) > 0 && song.Br > 0 && song.Size > 0 {
			return song
		}
		if song.Br > 0 && song.Size > 0 && !strings.Contains(song.Url, "qq.com") && !strings.Contains(song.Url, "xiami.net") && !strings.Contains(song.Url, "qianqian.com") {
			return song
		}
		header := make(http.Header, 1)
		header["range"] = append(header["range"], "bytes=0-8191")
		uri, err := url.Parse(song.Url)
		if err == nil {
			song.Url = uri.String()
		}
		clientRequest := network.ClientRequest{
			Method:         http.MethodGet,
			RemoteUrl:      song.Url,
			Header:         header,
			ConnectTimeout: time.Second * 2,
			Proxy:          true,
		}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			log.Println("processSong fail:", err)
			return song
		}
		defer resp.Body.Close()
		if resp.StatusCode > 199 && resp.StatusCode < 300 {
			if strings.Contains(song.Url, "qq.com") {
				song.Md5 = resp.Header.Get("server-md5")
			} else if strings.Contains(song.Url, "xiami.net") || strings.Contains(song.Url, "qianqian.com") {
				song.Md5 = strings.ToLower(utils.ReplaceAll(resp.Header.Get("etag"), `/"/g`, ""))
				//.replace(/"/g, '').toLowerCase()
			}
			if song.Size == 0 {
				size := resp.Header.Get("content-range")
				if len(size) > 0 {
					sizeSlice := strings.Split(size, "/")
					if len(sizeSlice) > 0 {
						size = sizeSlice[len(sizeSlice)-1]
					}
				} else {
					size = resp.Header.Get("content-length")
					if len(size) < 1 {
						size = "0"
					}
				}
				song.Size, _ = strconv.ParseInt(size, 10, 64)
			}
			if song.Br == 0 {
				//log.Println(utils.LogInterface(resp.Header))
				if resp.Header.Get("content-length") == "8192" || resp.ContentLength == 8192 {
					body, err := network.GetResponseBody(resp, false)
					if err != nil {
						log.Println("song GetResponseBody error:", err)
						return song
					}
					bitrate := decodeBitrate(body)
					if bitrate == 999 || (bitrate > 0 && bitrate < 500) {
						song.Br = bitrate * 1000
					}
				}
			}
		} else {
			return &common.Song{}
		}
	}
	return song
}
func decodeBitrate(data []byte) int {
	bitRateMap := map[int]map[int][]int{
		0: {
			3: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 500},
			2: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
			1: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
		},
		3: {
			3: {0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 500},
			2: {0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 500},
			1: {0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 500},
		},
		2: {
			3: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 500},
			2: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
			1: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
		},
	}

	var pointer = 0
	if strings.EqualFold(string(data[0:4]), "fLaC") {
		return 999
	}
	if strings.EqualFold(string(data[0:3]), "ID3") {
		pointer = 6
		var size = 0
		for index, value := range data[pointer : pointer+4] {
			size = size + int((value&0x7f)<<(7*(3-index)))
		}
		pointer = 10 + size
	}

	for i := 0; i < len(data); i++ { //fix migu mp3
		if data[pointer] != 0xff { //fail
			pointer = pointer + 1
			continue
		} else {
			break
		}
	}
	if pointer > len(data)-4 {
		return 0
	}
	header := data[pointer : pointer+4]
	// https://www.allegro.cc/forums/thread/591512/674023
	if len(header) == 4 &&
		header[0] == 0xff &&
		((header[1]>>5)&0x7) == 0x7 &&
		((header[1]>>1)&0x3) != 0 &&
		((header[2]>>4)&0xf) != 0xf &&
		((header[2]>>2)&0x3) != 0x3 {
		version := (header[1] >> 3) & 0x3
		layer := (header[1] >> 1) & 0x3
		bitrate := header[2] >> 4
		return bitRateMap[int(version)][int(layer)][int(bitrate)]
	}
	return 0
}
func checkCache(song *common.Song) bool {
	header := make(http.Header, 1)
	header["range"] = append(header["range"], "bytes=0-1")
	clientRequest := network.ClientRequest{
		Method:         http.MethodGet,
		RemoteUrl:      song.Url,
		Header:         header,
		Proxy:          true,
		ConnectTimeout: time.Second * 2,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		log.Println("checkCache fail:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return false
	} else {
		return true
	}
}
