package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/utils"
)

type DailyRecommendSongsMenu struct {
	menus []string
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

func (m *DailyRecommendSongsMenu) MenuViews() []string {
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
		response := recommendSongs.RecommendSongs()
		if _, ok := response["code"]; !ok {
			return false
		}
		code := utils.CheckCodeFromResponse(response)
		if code == utils.NeedLogin {
			model.showLogin = true
			return false
		}

		m.menus = []string{
			"dasdsad",
			"dsadsad",
			"dasdsad",
			"dsadsad",
		}

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

