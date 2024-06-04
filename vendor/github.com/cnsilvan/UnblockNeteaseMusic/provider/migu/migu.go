package migu

import (
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/tidwall/gjson"

	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
)

type Migu struct{}

var header = http.Header{
	"Origin":     []string{"https://music.migu.cn/"},
	"Referer":    []string{"https://m.music.migu.cn/v3/"},
	"Aversionid": []string{"null"},
	"Channel":    []string{"0146921"},
	"User-Agent": []string{"curl/8.8.0"},
}

func (m *Migu) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song = base.PreSearchSong(song)

	result, err := base.FetchV2(
		"https://m.music.migu.cn/migu/remoting/scr_search_tag?keyword="+url.QueryEscape(song.Keyword)+"&type=2&rows=20&pgc=1",
		nil, header, true)
	if err != nil {
		log.Println(err)
		return songs
	}

	var list [][]byte
	jsonparser.ArrayEach(result, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		list = append(list, value)
	}, "musics")
	listLength := len(list)
	maxIndex := listLength/2 + 1
	if maxIndex > 10 {
		maxIndex = 10
	}

	for index, matched := range list {
		if utils.StringFromJSON(matched, "copyrightId") == "" {
			continue
		}

		if index >= maxIndex {
			break
		}
		songResult := &common.Song{}
		songResult.Artist = strings.ReplaceAll(utils.StringFromJSON(matched, "singerName"), " ", "")
		songResult.Name = utils.StringFromJSON(matched, "songName")
		songResult.AlbumName = utils.StringFromJSON(matched, "albumName")
		songResult.Source = "migu"
		songResult.Id = utils.StringFromJSON(matched, "id")
		if len(songResult.Id) > 0 {
			songResult.Id = string(common.MiGuTag) + songResult.Id
		}

		songResult.PlatformUniqueKey = make(map[string]any)
		_ = jsonparser.ObjectEach(matched, func(key, value []byte, dataType jsonparser.ValueType, offset int) error {
			k := string(key)
			switch dataType {
			case jsonparser.String:
				songResult.PlatformUniqueKey[k], _ = jsonparser.ParseString(value)
			case jsonparser.Number:
				songResult.PlatformUniqueKey[k], _ = jsonparser.ParseFloat(value)
			case jsonparser.Object, jsonparser.Array:
				songResult.PlatformUniqueKey[k] = string(value)
			case jsonparser.Boolean:
				songResult.PlatformUniqueKey[k], _ = jsonparser.ParseBoolean(value)
			case jsonparser.Null:
			default:
			}
			return nil
		})
		songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword

		var ok bool
		songResult.MatchScore, ok = base.CalScore(song, songResult.Name, songResult.Artist, index, maxIndex)
		if !ok {
			continue
		}
		songs = append(songs, songResult)
	}

	return base.AfterSearchSong(song, songs)
}

func (m *Migu) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
	songId, ok := song.PlatformUniqueKey["id"].(string)
	if !ok {
		return song
	}
	types := []string{"PQ"}
	switch searchSong.Quality {
	case common.Standard:
		types = []string{"PQ"}
	case common.Higher:
		types = []string{"SQ", "PQ"}
	case common.ExHigh:
		types = []string{"HQ", "SQ", "PQ"}
	case common.Lossless:
		types = []string{"ZQ24", "HQ", "SQ", "PQ"}
	default:
		types = []string{"ZQ24", "HQ", "SQ", "PQ"}
	}
	urlChan := make(chan string, len(types))

	for _, formatType := range types {
		go func(song *common.Song, formatType string) {
			url := "https://app.c.nf.migu.cn/MIGUM2.0/strategy/listen-url/v2.4?netType=01&resourceType=2&songId=" + songId + "&toneFlag=" + formatType
			res, err := base.FetchV2(url, nil, header, true)
			if err != nil {
				slog.Error("migu request error", slog.Any("error", err))
				urlChan <- ""
				return
			}

			urlChan <- gjson.GetBytes(res, "data.url").String()
		}(song, formatType)
	}

	i := 0
	for u := range urlChan {
		i++
		if u == "" || song.Url != "" {
			continue
		}

		if strings.HasPrefix(u, "http") {
			song.Url = u
			continue
		}

		if strings.HasPrefix(u, "//") {
			song.Url = "http:" + u
			continue
		}

		if i >= len(types) {
			break
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
