package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
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
		if utils.CheckUserInfo(m.netease.user) == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		now := time.Now()
		if len(m.menus) > 0 && len(m.songs) > 0 && utils.IsSameDate(m.fetchTime, now) {
			return true, nil
		}
		recommendSongs := service.RecommendSongsService{}
		code, response := recommendSongs.RecommendSongs()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}
		m.songs = utils.GetDailySongs(response)
		m.menus = utils.GetViewFromSongs(m.songs)
		m.fetchTime = now

		return true, nil
	}
}

func (m *DailyRecommendSongsMenu) Songs() []structs.Song {
	return m.songs
}
