package netease

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// maxPlaylistSearchPages 限制按名称查找歌单的最大翻页数，防止服务端异常
// 时无限请求
const maxPlaylistSearchPages = 100

func FetchUserPlaylists(userId int64, limit int, offset int) (codeType _struct.ResCode, playlists []structs.Playlist, hasMore bool) {
	userPlaylists := service.UserPlaylistService{
		Uid:    strconv.FormatInt(userId, 10),
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}
	code, response := userPlaylists.UserPlaylist()
	codeType = _struct.CheckCode(code)
	if codeType != _struct.Success {
		return
	}

	playlists = _struct.GetPlaylists(response)
	menus := make([]model.MenuItem, len(playlists))
	for i := range playlists {
		menus[i] = model.MenuItem{Title: _struct.ReplaceSpecialStr(playlists[i].Name)}
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
		codeType   _struct.ResCode
		playlists  []structs.Playlist
		hasMore    bool
	)
	// 寻找歌单
Loop:
	for page := 0; ; page++ {
		codeType, playlists, hasMore = FetchUserPlaylists(userId, 30, offset)
		if codeType != _struct.Success {
			err = mapResCodeToErr(codeType)
			return
		}
		offset += len(playlists)
		for _, p := range playlists {
			if p.Name == playlistName {
				playlistId = p.Id
				break Loop
			}
		}
		if !hasMore || len(playlists) == 0 || page >= maxPlaylistSearchPages {
			// len==0 防服务端异常持续返回"空列表+more=true"导致的死循环
			err = Error{Msg: "未找到歌单:" + playlistName}
			return
		}
	}
	codeType, songs = FetchSongsOfPlaylist(playlistId, getAll)
	if codeType != _struct.Success {
		err = mapResCodeToErr(codeType)
		return
	}
	return
}
