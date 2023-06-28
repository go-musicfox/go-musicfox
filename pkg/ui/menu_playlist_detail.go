package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type PlaylistDetailMenu struct {
	DefaultMenu
	menus      []MenuItem
	songs      []structs.Song
	playlistId int64
}

func NewPlaylistDetailMenu(playlistId int64) *PlaylistDetailMenu {
	return &PlaylistDetailMenu{
		playlistId: playlistId,
	}
}

func (m *PlaylistDetailMenu) IsSearchable() bool {
	return true
}

func (m *PlaylistDetailMenu) IsPlayable() bool {
	return true
}

func (m *PlaylistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("playlist_detail_%d", m.playlistId)
}

func (m *PlaylistDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *PlaylistDetailMenu) SubMenu(_ *NeteaseModel, _ int) Menu {
	return nil
}

func getSongsInPlaylist(playlistId int64, getAll bool) (codeType utils.ResCode, songs []structs.Song) {
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

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		codeType, songs := getSongsInPlaylist(m.playlistId, configs.ConfigRegistry.MainShowAllSongsOfPlaylist)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}
		m.songs = songs
		m.menus = GetViewFromSongs(songs)

		return true
	}
}

func (m *PlaylistDetailMenu) Songs() []structs.Song {
	return m.songs
}
