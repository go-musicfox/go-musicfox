package migu

import (
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"

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

	list := gjson.GetBytes(result, "musics").Array()
	listLength := len(list)
	maxIndex := listLength/2 + 1
	if maxIndex > 10 {
		maxIndex = 10
	}

	for index, matched := range list {
		if matched.Get("copyrightId").String() == "" {
			continue
		}

		if index >= maxIndex {
			break
		}
		songResult := &common.Song{}
		songResult.Artist = strings.ReplaceAll(matched.Get("singerName").String(), " ", "")
		songResult.Name = matched.Get("songName").String()
		songResult.AlbumName = matched.Get("albumName").String()
		songResult.Source = "migu"
		songResult.Id = matched.Get("id").String()
		if len(songResult.Id) > 0 {
			songResult.Id = string(common.MiGuTag) + songResult.Id
		}

		songResult.PlatformUniqueKey = make(map[string]any)
		for k, v := range matched.Map() {
			switch {
			case v.IsArray():
				songResult.PlatformUniqueKey[k] = v.Array()
			case v.IsObject():
				songResult.PlatformUniqueKey[k] = v.Map()
			case v.IsBool():
				songResult.PlatformUniqueKey[k] = v.Bool()
			default:
				songResult.PlatformUniqueKey[k] = v.String()
			}
		}
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
	types := []string{"ZQ24", "SQ", "HQ", "PQ"}
	urlChan := make(chan string, len(types))
	defer close(urlChan)

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
		if u == "" {
			continue
		}

		if strings.HasPrefix(u, "http") {
			song.Url = u
			return song
		}

		if strings.HasPrefix(u, "//") {
			song.Url = "http:" + u
			return song
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
