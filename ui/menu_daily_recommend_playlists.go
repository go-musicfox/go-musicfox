package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/ds"
	"go-musicfox/utils"
)

type DailyRecommendPlaylistsMenu struct {
	menus     []MenuItem
	playlists []ds.Playlist
}

func NewDailyRecommendPlaylistMenu() *DailyRecommendPlaylistsMenu {
	return new(DailyRecommendPlaylistsMenu)
}

func (m *DailyRecommendPlaylistsMenu) MenuData() interface{} {
	return m.playlists
}

func (m *DailyRecommendPlaylistsMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *DailyRecommendPlaylistsMenu) IsPlayable() bool {
	return false
}

func (m *DailyRecommendPlaylistsMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *DailyRecommendPlaylistsMenu) GetMenuKey() string {
	return "daily_playlists"
}

func (m *DailyRecommendPlaylistsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DailyRecommendPlaylistsMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if index >= len(m.playlists) {
		return nil
	}
	return NewPlaylistDetailMenu(m.playlists[index].Id)
}

func (m *DailyRecommendPlaylistsMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *DailyRecommendPlaylistsMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *DailyRecommendPlaylistsMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		// 不重复请求
		if len(m.menus) > 0 && len(m.playlists) > 0 {
			return true
		}

		recommendPlaylists := service.RecommendResourceService{}
		code, response := recommendPlaylists.RecommendResource()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}
		m.playlists = utils.GetDailyPlaylists(response)
		for _, playlist := range m.playlists {
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(playlist.Name), ""})
		}

		return true
	}
}

func (m *DailyRecommendPlaylistsMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *DailyRecommendPlaylistsMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}

