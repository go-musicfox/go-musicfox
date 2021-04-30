package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/ds"
	"go-musicfox/utils"
	"strconv"
	"strings"
)

type PlaylistDetailMenu struct {
	menus 	   []MenuItem
	songs      []ds.Song
	PlaylistId int64
}

func (m *PlaylistDetailMenu) MenuData() interface{} {
	return m.songs
}

func (m *PlaylistDetailMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *PlaylistDetailMenu) IsPlayable() bool {
	return true
}

func (m *PlaylistDetailMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *PlaylistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("playlist_detail_%d", m.PlaylistId)
}

func (m *PlaylistDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *PlaylistDetailMenu) SubMenu(model *NeteaseModel, index int) IMenu {
	return nil
}

func (m *PlaylistDetailMenu) ExtraView() string {
	return ""
}

func (m *PlaylistDetailMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *PlaylistDetailMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(m.PlaylistId, 10), S: "0"}	// 最近S个收藏者，设为0
		code, response := playlistDetail.PlaylistDetail()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetSongsOfPlaylist(response)
		var menus []MenuItem
		for _, song := range m.songs {
			var artists []string
			for _, artist := range song.Artists {
				artists = append(artists, artist.Name)
			}
			menus = append(menus, MenuItem{utils.ReplaceSpecialStr(song.Name), utils.ReplaceSpecialStr(strings.Join(artists, ","))})
		}
		m.menus = menus

		return true
	}
}

func (m *PlaylistDetailMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *PlaylistDetailMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}

