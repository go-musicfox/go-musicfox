package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type DailyRecommendPlaylistsMenu struct {
	DefaultMenu
	menus     []MenuItem
	playlists []structs.Playlist
}

func NewDailyRecommendPlaylistMenu() *DailyRecommendPlaylistsMenu {
	return new(DailyRecommendPlaylistsMenu)
}

func (m *DailyRecommendPlaylistsMenu) MenuData() interface{} {
	return m.playlists
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
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}

		return true
	}
}
