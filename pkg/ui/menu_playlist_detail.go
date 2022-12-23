package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/configs"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
)

type PlaylistDetailMenu struct {
	DefaultMenu
	menus      []MenuItem
	songs      []structs.Song
	PlaylistId int64
}

func NewPlaylistDetailMenu(playlistId int64) *PlaylistDetailMenu {
	return &PlaylistDetailMenu{
		PlaylistId: playlistId,
	}
}

func (m *PlaylistDetailMenu) IsSearchable() bool {
	return true
}

func (m *PlaylistDetailMenu) MenuData() interface{} {
	return m.songs
}

func (m *PlaylistDetailMenu) IsPlayable() bool {
	return true
}

func (m *PlaylistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("playlist_detail_%d", m.PlaylistId)
}

func (m *PlaylistDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *PlaylistDetailMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
	return nil
}

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		var (
			code     float64
			response []byte
		)
		if !configs.ConfigRegistry.MainShowAllSongsOfPlaylist {
			playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(m.PlaylistId, 10), S: "0"} // 最近S个收藏者，设为0
			code, response = playlistDetail.PlaylistDetail()
		} else {
			allTrack := service.PlaylistTrackAllService{Id: strconv.FormatInt(m.PlaylistId, 10), S: "0"} // 最近S个收藏者，设为0
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
