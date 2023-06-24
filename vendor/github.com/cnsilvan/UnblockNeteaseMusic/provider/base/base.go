package base

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

func PreSearchSong(song common.SearchSong) common.SearchSong {
	song.Keyword = strings.ToUpper(song.Keyword)
	song.Name = strings.ToUpper(song.Name)
	song.ArtistsName = strings.ToUpper(song.ArtistsName)
	return song
}
func Fetch(url string, cookies []*http.Cookie, header http.Header, proxy bool) (result map[string]interface{}, err error) {
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: url,
		Cookies:   cookies,
		Header:    header,
		Proxy:     proxy,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("StatusCode :" + strconv.Itoa(resp.StatusCode))
		return nil, err
	}
	defer resp.Body.Close()
	body, err := network.StealResponseBody(resp)
	if err != nil {
		return nil, err
	}
	result = utils.ParseJsonV2(body)
	return result, nil
}
func CalScore(song common.SearchSong, songName string, singerName string, index int, maxIndex int) (float32, bool) {
	if song.OrderBy == common.MatchedScoreDesc {
		if strings.Contains(songName, "伴奏") && !strings.Contains(song.Keyword, "伴奏") {
			return 0, false
		}
		if strings.Contains(strings.ToUpper(songName), "DJ") &&
			!strings.Contains(strings.ToUpper(song.Keyword), "DJ") {
			return 0, false
		}
		if strings.Contains(strings.ToUpper(songName), "COVER") &&
			!strings.Contains(strings.ToUpper(song.Keyword), "COVER") {
			return 0, false
		}
		var songNameSores float32 = 0.0
		if len(songName) > 0 {
			for _, name := range strings.Split(songName, "-") {
				score := utils.CalMatchScoresV2(song.Name, name, "songName")
				if score > songNameSores {
					songNameSores = score
				}
			}
		}
		var artistsNameSores float32 = 0.0
		if len(singerName) > 0 {
			singerName = strings.ReplaceAll(singerName, "&", "、")
			singerName = strings.ReplaceAll(singerName, "·", "、")
			artistsNameSores = utils.CalMatchScoresV2(song.ArtistsName, singerName, "singerName")
		}
		songMatchScore := songNameSores*0.55 + artistsNameSores*0.35 + 0.1*float32(maxIndex-index)/float32(maxIndex)
		return songMatchScore, true
	} else if song.OrderBy == common.PlatformDefault {

	}
	return 0, true
}
func AfterSearchSong(song common.SearchSong, songs []*common.Song) []*common.Song {
	if song.OrderBy == common.MatchedScoreDesc && len(songs) > 1 {
		sort.Sort(common.SongSlice(songs))
	}
	if song.Limit > 0 && len(songs) > song.Limit {
		songs = songs[:song.Limit]
	}
	return songs
}
