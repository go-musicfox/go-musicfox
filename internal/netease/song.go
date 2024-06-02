package netease

import (
	"strconv"

	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

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
	codeType = _struct.CheckCode(code)
	if codeType != _struct.Success {
		return
	}
	songs = _struct.GetSongsOfPlaylist(response)

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
