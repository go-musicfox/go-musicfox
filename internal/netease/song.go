package netease

import (
	"strconv"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/netease-music/service"
)

func FetchDailySongs() (playlist []structs.Song, err error) {
	recommendSongs := service.RecommendSongsService{}
	code, response := recommendSongs.RecommendSongs()
	codeType := utils.CheckCode(code)
	if codeType != utils.Success {
		err = NetworkErr
		return
	}
	playlist = utils.GetDailySongs(response)
	return
}

func FetchSongsOfPlaylist(playlistId int64, getAll bool) (codeType utils.ResCode, songs []structs.Song) {
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
	codeType = utils.CheckCode(code)
	if codeType != utils.Success {
		return
	}
	songs = utils.GetSongsOfPlaylist(response)

	return
}

func FetchLikeSongs(userId int64, getAll bool) (playlist []structs.Song, err error) {
	var (
		codeType  utils.ResCode
		playlists []structs.Playlist
		songs     []structs.Song
	)
	codeType, playlists, _ = FetchUserPlaylists(userId, 1, 0)
	if codeType != utils.Success {
		err = NetworkErr
		return
	}
	codeType, songs = FetchSongsOfPlaylist(playlists[0].Id, getAll)
	if codeType != utils.Success {
		err = NetworkErr
		return
	}
	playlist = songs
	return
}
