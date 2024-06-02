package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

type DailyRecommendSongsMenu struct {
	baseMenu
	menus     []model.MenuItem
	songs     []structs.Song
	fetchTime time.Time
}

func NewDailyRecommendSongsMenu(baseMenu baseMenu) *DailyRecommendSongsMenu {
	return &DailyRecommendSongsMenu{
		baseMenu: baseMenu,
	}
}

func (m *DailyRecommendSongsMenu) IsSearchable() bool {
	return true
}

func (m *DailyRecommendSongsMenu) IsPlayable() bool {
	return true
}

func (m *DailyRecommendSongsMenu) GetMenuKey() string {
	return "daily_songs"
}

func (m *DailyRecommendSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DailyRecommendSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if _struct.CheckUserInfo(m.netease.user) == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		now := time.Now()
		if len(m.menus) > 0 && len(m.songs) > 0 && timex.IsSameDate(m.fetchTime, now) {
			return true, nil
		}
		recommendSongs := service.RecommendSongsService{}
		code, response := recommendSongs.RecommendSongs()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}
		m.songs = _struct.GetDailySongs(response)
		m.menus = menux.GetViewFromSongs(m.songs)
		m.fetchTime = now

		return true, nil
	}
}

func (m *DailyRecommendSongsMenu) Songs() []structs.Song {
	return m.songs
}
