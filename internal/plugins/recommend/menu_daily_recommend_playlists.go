package recommend

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

type DailyRecommendPlaylistsMenu struct {
	ui.BaseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
	fetchTime time.Time
}

func NewDailyRecommendPlaylistMenu(base ui.BaseMenu) *DailyRecommendPlaylistsMenu {
	return &DailyRecommendPlaylistsMenu{
		BaseMenu: base,
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
	playlistMenu, err := ui.BuildMenu("playlist_detail", m.BaseMenu, ui.PlaylistDetailOpts{PlaylistID: m.playlists[index].Id})
	if err != nil {
		return nil
	}
	return playlistMenu
}

func (m *DailyRecommendPlaylistsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if _struct.CheckUserInfo(m.User()) == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
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
			page, _ := m.ToLoginPage(enterMenuCallback(main))
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
