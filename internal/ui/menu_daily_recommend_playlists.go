package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
	"github.com/go-musicfox/go-musicfox/utils/timex"
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
		if _struct.CheckUserInfo(m.netease.user) == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		// 不重复请求
		now := time.Now()
		if len(m.menus) > 0 && len(m.playlists) > 0 && timex.IsSameDate(m.fetchTime, now) {
			return true, nil
		}

		recommendPlaylists := service.RecommendResourceService{}
		code, response := recommendPlaylists.RecommendResource()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}
		m.playlists = _struct.GetDailyPlaylists(response)
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus
		m.fetchTime = now

		return true, nil
	}
}

func (m *DailyRecommendPlaylistsMenu) Playlists() []structs.Playlist {
	return m.playlists
}
