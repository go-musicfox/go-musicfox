package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/utils"
	"strings"
)

type DailyRecommendSongsMenu struct {
	menus []MenuItem
}

func (m *DailyRecommendSongsMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *DailyRecommendSongsMenu) IsPlayable() bool {
	return false
}

func (m *DailyRecommendSongsMenu) ResetPlaylistWhenEnter() bool {
	return false
}

func (m *DailyRecommendSongsMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *DailyRecommendSongsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DailyRecommendSongsMenu) SubMenu(index int) IMenu {
	return nil
}

func (m *DailyRecommendSongsMenu) ExtraView() string {
	return ""
}

func (m *DailyRecommendSongsMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *DailyRecommendSongsMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *DailyRecommendSongsMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		recommendSongs := service.RecommendSongsService{}
		code, response := recommendSongs.RecommendSongs()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			model.showLogin = true
			return false
		}
		list := utils.GetListFromSongs(response)
		for _, song := range list {
			var artists []string
			for _, artist := range song.Artists {
				artists = append(artists, artist.Name)
			}
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(song.Name), utils.ReplaceSpecialStr(strings.Join(artists, ","))})
		}

		model.menuData = list

		return true
	}
}

func (m *DailyRecommendSongsMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *DailyRecommendSongsMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}

