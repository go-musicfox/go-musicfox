package netease

import (
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

const (
	playlistDetailEAPIPath = "/api/v6/playlist/detail"
	playlistDetailEAPIURL  = "https://music.163.com/eapi/v6/playlist/detail"
	songDetailBatchSize    = 500
)

var (
	requestPlaylistDetail = func(playlistId int64) (float64, []byte) {
		options := &util.Options{
			Crypto: "eapi",
			Url:    playlistDetailEAPIPath,
		}
		data := map[string]string{
			"id": strconv.FormatInt(playlistId, 10),
			"n":  "100000",
			"s":  "0",
		}
		code, response, _ := util.CreateRequest("POST", playlistDetailEAPIURL, data, options)
		return code, response
	}

	requestSongDetail = func(ids string) (float64, []byte) {
		songDetail := service.SongDetailService{Ids: ids}
		return songDetail.SongDetail()
	}
)

// resolveResponseCode validates both the transport/client status code and the
// JSON business code in the response body. util.CreateRequest defaults to 200
// when "code" is missing or unparseable (e.g. a 502 HTML page), so we verify
// the JSON business code explicitly.
func resolveResponseCode(code float64, response []byte) _struct.ResCode {
	if codeType := _struct.CheckCode(code); codeType != _struct.Success {
		return codeType
	}
	parsedCode, err := jsonparser.GetFloat(response, "code")
	if err != nil {
		return _struct.NetworkError
	}
	return _struct.CheckCode(parsedCode)
}

func extractPlaylistTrackIDs(response []byte) []int64 {
	var trackIDs []int64
	_, _ = jsonparser.ArrayEach(response, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
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
		batch := ids[start:end]

		idStrs := make([]string, len(batch))
		for i, id := range batch {
			idStrs[i] = strconv.FormatInt(id, 10)
		}

		code, response := requestSongDetail(strings.Join(idStrs, ","))
		if codeType := resolveResponseCode(code, response); codeType != _struct.Success {
			return codeType, nil
		}

		_, _ = jsonparser.ArrayEach(response, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if song, err := structs.NewSongFromShortNameSongsJson(value); err == nil {
				songs = append(songs, song)
			}
		}, "songs")
	}
	return _struct.Success, songs
}

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

func FetchSongsOfPlaylist(playlistId int64, getAll bool) (codeType _struct.ResCode, songs []structs.Song) {
	code, response := requestPlaylistDetail(playlistId)
	codeType = resolveResponseCode(code, response)
	if codeType != _struct.Success {
		return codeType, nil
	}
	if _, dataType, _, err := jsonparser.Get(response, "playlist"); err != nil || dataType != jsonparser.Object {
		return _struct.UnknownError, nil
	}

	tracks := _struct.GetSongsOfPlaylist(response)
	trackIDs := extractPlaylistTrackIDs(response)
	if len(trackIDs) == 0 {
		return _struct.Success, tracks
	}
	if len(tracks) > 0 && !getAll {
		return _struct.Success, tracks
	}

	songsByID := make(map[int64]structs.Song, len(tracks))
	for _, song := range tracks {
		songsByID[song.Id] = song
	}

	missingIDs := make([]int64, 0, len(trackIDs))
	seenMissing := make(map[int64]struct{}, len(trackIDs))
	for _, id := range trackIDs {
		if _, ok := songsByID[id]; ok {
			continue
		}
		if _, seen := seenMissing[id]; seen {
			continue
		}
		seenMissing[id] = struct{}{}
		missingIDs = append(missingIDs, id)
	}
	if len(missingIDs) == 0 {
		return _struct.Success, tracks
	}

	fetchedCode, fetchedSongs := fetchSongsByIDs(missingIDs)
	if fetchedCode != _struct.Success {
		return fetchedCode, nil
	}
	for _, song := range fetchedSongs {
		songsByID[song.Id] = song
	}

	songs = make([]structs.Song, 0, len(trackIDs))
	for _, id := range trackIDs {
		if song, ok := songsByID[id]; ok {
			songs = append(songs, song)
		}
	}
	return _struct.Success, songs
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
