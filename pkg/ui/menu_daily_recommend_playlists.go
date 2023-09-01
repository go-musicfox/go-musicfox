package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DailyRecommendPlaylistsMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
	fetchTime time.Time
}

func NewDailyRecommendPlaylistMenu(baseMenu baseMenu) *DailyRecommendPlaylistsMenu {
	return &DailyRecommendPlaylistsMenu{
		baseMenu: baseMenu,
	}
}

func (m *DailyRecommendPlaylistsMenu) IsSearchable() bool {
	return true
}

func (m *DailyRecommendPlaylistsMenu) GetMenuKey() string {
	return "daily_playlists"
}

func (m *DailyRecommendPlaylistsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DailyRecommendPlaylistsMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.playlists) {
		return nil
	}
	return NewPlaylistDetailMenu(m.baseMenu, m.playlists[index].Id)
}

func (m *DailyRecommendPlaylistsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if utils.CheckUserInfo(m.netease.user) == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		// 不重复请求
		now := time.Now()
		if len(m.menus) > 0 && len(m.playlists) > 0 && utils.IsSameDate(m.fetchTime, now) {
			return true, nil
		}

		recommendPlaylists := service.RecommendResourceService{}
		code, response := recommendPlaylists.RecommendResource()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}
		m.playlists = utils.GetDailyPlaylists(response)
		for _, playlist := range m.playlists {
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}
		m.fetchTime = now

		return true, nil
	}
}

func (m *DailyRecommendPlaylistsMenu) Playlists() []structs.Playlist {
	return m.playlists
}
