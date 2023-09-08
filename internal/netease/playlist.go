package netease

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/netease-music/service"
)

func FetchUserPlaylists(userId int64, limit int, offset int) (codeType utils.ResCode, playlists []structs.Playlist, hasMore bool) {
	userPlaylists := service.UserPlaylistService{
		Uid:    strconv.FormatInt(userId, 10),
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}
	code, response := userPlaylists.UserPlaylist()
	codeType = utils.CheckCode(code)
	if codeType != utils.Success {
		return
	}

	playlists = utils.GetPlaylists(response)
	menus := make([]model.MenuItem, len(playlists))
	for i := range playlists {
		menus[i] = model.MenuItem{Title: utils.ReplaceSpecialStr(playlists[i].Name)}
	}

	// 是否有更多
	var err error
	if hasMore, err = jsonparser.GetBoolean(response, "more"); err != nil {
		hasMore = false
	}

	return
}

func FetchUserPlaylistByName(userId int64, playlistName string, getAll bool) (songs []structs.Song, err error) {
	var (
		playlistId int64
		offset     = 0
		codeType   utils.ResCode
		playlists  []structs.Playlist
		hasMore    bool
	)
	// 寻找歌单
Loop:
	for {
		codeType, playlists, hasMore = FetchUserPlaylists(userId, 30, offset)
		if codeType != utils.Success {
			err = NetworkErr
			return
		}
		offset += len(playlists)
		for _, p := range playlists {
			if p.Name == playlistName {
				playlistId = p.Id
				break Loop
			}
		}
		if !hasMore {
			err = Error{Msg: "未找到歌单:" + playlistName}
			return
		}
	}
	codeType, songs = FetchSongsOfPlaylist(playlistId, getAll)
	if codeType != utils.Success {
		err = NetworkErr
		return
	}
	return
}
