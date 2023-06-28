package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type HighQualityPlaylistsMenu struct {
	DefaultMenu
	menus     []MenuItem
	playlists []structs.Playlist
}

func NewHighQualityPlaylistsMenu() *HighQualityPlaylistsMenu {
	return new(HighQualityPlaylistsMenu)
}

func (m *HighQualityPlaylistsMenu) IsSearchable() bool {
	return true
}

func (m *HighQualityPlaylistsMenu) GetMenuKey() string {
	return "high_quality_playlists"
}

func (m *HighQualityPlaylistsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *HighQualityPlaylistsMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.playlists) {
		return nil
	}
	return NewPlaylistDetailMenu(m.playlists[index].Id)
}

func (m *HighQualityPlaylistsMenu) Playlists() []structs.Playlist {
	return m.playlists
}

func (m *HighQualityPlaylistsMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		// 不重复请求
		if len(m.menus) > 0 && len(m.playlists) > 0 {
			return true
		}

		highQualityPlaylists := service.TopPlaylistHighqualityService{
			Limit: "80",
		}
		code, response := highQualityPlaylists.TopPlaylistHighquality()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		m.playlists = utils.GetPlaylistsFromHighQuality(response)
		for _, playlist := range m.playlists {
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}

		return true
	}
}
