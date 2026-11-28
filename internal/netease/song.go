package netease

import (
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

const songDetailBatchSize = 500

func FetchDailySongs() (playlist []structs.Song, err error) {
	recommendSongs := service.RecommendSongsService{}
	code, response := recommendSongs.RecommendSongs()
	codeType := _struct.CheckCode(code)
	if codeType != _struct.Success {
		err = NetworkErr
		return
	}
	playlist = _struct.GetDailySongs(response)
	return
}

// resolveResponseCode validates both the transport code and the JSON business
// code. CreateRequest returns 200 when a response has no business code (for
// example a 502 HTML page), so the body must be checked explicitly.
func resolveResponseCode(code float64, response []byte) _struct.ResCode {
	if codeType := _struct.CheckCode(code); codeType != _struct.Success {
		return codeType
	}
	businessCode, err := jsonparser.GetFloat(response, "code")
	if err != nil {
		return _struct.NetworkError
	}
	return _struct.CheckCode(businessCode)
}

func extractPlaylistTrackIDs(response []byte) []int64 {
	trackIDs := make([]int64, 0)
	_, _ = jsonparser.ArrayEach(response, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		if id, err := jsonparser.GetInt(value, "id"); err == nil && id > 0 {
			trackIDs = append(trackIDs, id)
		}
	}, "playlist", "trackIds")
	return trackIDs
}

func fetchSongsByIDs(ids []int64) (_struct.ResCode, []structs.Song) {
	songs := make([]structs.Song, 0, len(ids))
	for start := 0; start < len(ids); start += songDetailBatchSize {
		end := min(start+songDetailBatchSize, len(ids))
		idStrings := make([]string, 0, end-start)
		for _, id := range ids[start:end] {
			idStrings = append(idStrings, strconv.FormatInt(id, 10))
		}
		detail := service.SongDetailService{Ids: strings.Join(idStrings, ",")}
		code, response := detail.SongDetail()
		if codeType := resolveResponseCode(code, response); codeType != _struct.Success {
			return codeType, nil
		}
		_, _ = jsonparser.ArrayEach(response, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
			if song, err := structs.NewSongFromShortNameSongsJson(value); err == nil {
				songs = append(songs, song)
			}
		}, "songs")
	}
	return _struct.Success, songs
}

func FetchSongsOfPlaylist(playlistId int64, getAll bool) (codeType _struct.ResCode, songs []structs.Song) {
	var (
		code     float64
		response []byte
	)
	if !getAll {
		playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
		code, response = playlistDetail.PlaylistDetail()
	} else {
		allTrack := service.PlaylistTrackAllService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
		code, response = allTrack.AllTracks()
	}
	if codeType = resolveResponseCode(code, response); codeType != _struct.Success {
		return
	}
	if _, _, _, err := jsonparser.Get(response, "playlist"); err != nil {
		return _struct.NetworkError, nil
	}
	songs = _struct.GetSongsOfPlaylist(response)
	if len(songs) == 0 {
		if trackIDs := extractPlaylistTrackIDs(response); len(trackIDs) > 0 {
			codeType, songs = fetchSongsByIDs(trackIDs)
		}
	}
	return
}

func FetchLikeSongs(userId int64, getAll bool) (playlist []structs.Song, err error) {
	var (
		codeType  _struct.ResCode
		playlists []structs.Playlist
		songs     []structs.Song
	)
	codeType, playlists, _ = FetchUserPlaylists(userId, 1, 0)
	if codeType != _struct.Success {
		err = NetworkErr
		return
	}
	codeType, songs = FetchSongsOfPlaylist(playlists[0].Id, getAll)
	if codeType != _struct.Success {
		err = NetworkErr
		return
	}
	playlist = songs
	return
}
