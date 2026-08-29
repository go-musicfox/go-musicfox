package artist

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// HotArtistsMenu 展示热门歌手（主菜单「热门歌手」入口对应的插件菜单）。
type HotArtistsMenu struct {
	ui.BaseMenu
	menus   []model.MenuItem
	artists []structs.Artist
}

func NewHotArtistsMenu(base ui.BaseMenu) *HotArtistsMenu {
	return &HotArtistsMenu{
		BaseMenu: base,
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
	artistMenu, err := ui.BuildMenu("artist_detail", m.BaseMenu, ui.ArtistDetailOpts{ArtistID: m.artists[index].Id, Name: m.artists[index].Name})
	if err != nil {
		return nil
	}
	return artistMenu
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
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		m.artists = _struct.GetArtistsOfTopArtists(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(artist.Name)})
		}

		return true, nil
	}
}

func (m *HotArtistsMenu) Artists() []structs.Artist {
	return m.artists
}
