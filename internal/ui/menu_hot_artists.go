package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
)

type HotArtistsMenu struct {
	baseMenu
	menus   []model.MenuItem
	artists []structs.Artist
}

func NewHotArtistsMenu(base baseMenu) *HotArtistsMenu {
	return &HotArtistsMenu{
		baseMenu: base,
	}
}

func (m *HotArtistsMenu) IsSearchable() bool {
	return true
}

func (m *HotArtistsMenu) GetMenuKey() string {
	return "hot_artists"
}

func (m *HotArtistsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *HotArtistsMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.artists) {
		return nil
	}
	return NewArtistDetailMenu(m.baseMenu, m.artists[index].Id, m.artists[index].Name)
}

func (m *HotArtistsMenu) BeforeEnterMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		// 不重复请求
		if len(m.menus) > 0 && len(m.artists) > 0 {
			return true, nil
		}

		artistService := service.TopArtistsService{
			Limit: "80",
		}
		code, response := artistService.TopArtists()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}
		m.artists = utils.GetArtistsOfTopArtists(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(artist.Name)})
		}

		return true, nil
	}
}

func (m *HotArtistsMenu) Artists() []structs.Artist {
	return m.artists
}
