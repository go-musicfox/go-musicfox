package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type HighQualityPlaylistsMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
}

func NewHighQualityPlaylistsMenu(base baseMenu) *HighQualityPlaylistsMenu {
	return &HighQualityPlaylistsMenu{
		baseMenu: base,
	}
}

func (m *HighQualityPlaylistsMenu) IsSearchable() bool {
	return true
}

func (m *HighQualityPlaylistsMenu) GetMenuKey() string {
	return "high_quality_playlists"
}

func (m *HighQualityPlaylistsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *HighQualityPlaylistsMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.playlists) {
		return nil
	}
	return NewPlaylistDetailMenu(m.baseMenu, m.playlists[index].Id)
}

func (m *HighQualityPlaylistsMenu) Playlists() []structs.Playlist {
	return m.playlists
}

func (m *HighQualityPlaylistsMenu) BeforeEnterMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		// 不重复请求
		if len(m.menus) > 0 && len(m.playlists) > 0 {
			return true, nil
		}

		highQualityPlaylists := service.TopPlaylistHighqualityService{
			Limit: "80",
		}
		code, response := highQualityPlaylists.TopPlaylistHighquality()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		m.playlists = _struct.GetPlaylistsFromHighQuality(response)
		for _, playlist := range m.playlists {
			m.menus = append(m.menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
		}

		return true, nil
	}
}
