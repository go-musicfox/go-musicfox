package ui

import (
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DailyRecommendPlaylistsMenu struct {
	DefaultMenu
	menus     []MenuItem
	playlists []structs.Playlist
	fetchTime time.Time
}

func NewDailyRecommendPlaylistMenu() *DailyRecommendPlaylistsMenu {
	return new(DailyRecommendPlaylistsMenu)
}

func (m *DailyRecommendPlaylistsMenu) IsSearchable() bool {
	return true
}

func (m *DailyRecommendPlaylistsMenu) GetMenuKey() string {
	return "daily_playlists"
}

func (m *DailyRecommendPlaylistsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DailyRecommendPlaylistsMenu) SubMenu(_ *NeteaseModel, index int) Menu {
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
		now := time.Now()
		if len(m.menus) > 0 && len(m.playlists) > 0 && utils.IsSameDate(m.fetchTime, now) {
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
		m.fetchTime = now

		return true
	}
}

func (m *DailyRecommendPlaylistsMenu) Playlists() []structs.Playlist {
	return m.playlists
}
