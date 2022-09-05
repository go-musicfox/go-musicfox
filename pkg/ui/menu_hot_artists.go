package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type HotArtistsMenu struct {
    menus   []MenuItem
    artists []structs.Artist
}

func NewHotArtistsMenu() *HotArtistsMenu {
    return new(HotArtistsMenu)
}

func (m *HotArtistsMenu) MenuData() interface{} {
    return m.artists
}

func (m *HotArtistsMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *HotArtistsMenu) IsPlayable() bool {
    return false
}

func (m *HotArtistsMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *HotArtistsMenu) GetMenuKey() string {
    return "hot_artists"
}

func (m *HotArtistsMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *HotArtistsMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index >= len(m.artists) {
        return nil
    }
    return NewArtistDetailMenu(m.artists[index].Id)
}

func (m *HotArtistsMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *HotArtistsMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
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
            m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(artist.Name), ""})
        }

        return true
    }
}

func (m *HotArtistsMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *HotArtistsMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
