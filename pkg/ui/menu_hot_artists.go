package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type HotArtistsMenu struct {
	DefaultMenu
	menus   []MenuItem
	artists []structs.Artist
}

func NewHotArtistsMenu() *HotArtistsMenu {
	return new(HotArtistsMenu)
}

func (m *HotArtistsMenu) IsSearchable() bool {
	return true
}

func (m *HotArtistsMenu) GetMenuKey() string {
	return "hot_artists"
}

func (m *HotArtistsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *HotArtistsMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.artists) {
		return nil
	}
	return NewArtistDetailMenu(m.artists[index].Id, m.artists[index].Name)
}

func (m *HotArtistsMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		// 不重复请求
		if len(m.menus) > 0 && len(m.artists) > 0 {
			return true
		}

		artistService := service.TopArtistsService{
			Limit: "80",
		}
		code, response := artistService.TopArtists()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		m.artists = utils.GetArtistsOfTopArtists(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(artist.Name)})
		}

		return true
	}
}

func (m *HotArtistsMenu) Artists() []structs.Artist {
	return m.artists
}
