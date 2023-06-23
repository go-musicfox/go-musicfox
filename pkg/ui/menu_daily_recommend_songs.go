package ui

import (
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DailyRecommendSongsMenu struct {
	DefaultMenu
	menus     []MenuItem
	songs     []structs.Song
	fetchTime time.Time
}

func NewDailyRecommendSongsMenu() *DailyRecommendSongsMenu {
	return new(DailyRecommendSongsMenu)
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

func (m *DailyRecommendSongsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DailyRecommendSongsMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		now := time.Now()
		if len(m.menus) > 0 && len(m.songs) > 0 && utils.IsSameDate(m.fetchTime, now) {
			return true
		}
		recommendSongs := service.RecommendSongsService{}
		code, response := recommendSongs.RecommendSongs()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetDailySongs(response)
		m.menus = GetViewFromSongs(m.songs)
		m.fetchTime = now

		return true
	}
}

func (m *DailyRecommendSongsMenu) Songs() []structs.Song {
	return m.songs
}
