package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
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

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		var (
			code     float64
			response []byte
		)
		if !configs.ConfigRegistry.MainShowAllSongsOfPlaylist {
			playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(m.playlistId, 10), S: "0"} // 最近S个收藏者，设为0
			code, response = playlistDetail.PlaylistDetail()
		} else {
			allTrack := service.PlaylistTrackAllService{Id: strconv.FormatInt(m.playlistId, 10), S: "0"} // 最近S个收藏者，设为0
			code, response = allTrack.AllTracks()
		}
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetSongsOfPlaylist(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *PlaylistDetailMenu) Songs() []structs.Song {
	return m.songs
}
