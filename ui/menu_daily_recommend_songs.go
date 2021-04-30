package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/ds"
	"go-musicfox/utils"
	"strings"
)

type DailyRecommendSongsMenu struct {
	menus []MenuItem
	songs []ds.Song
}

func NewDailyRecommendSongsMenu() *DailyRecommendSongsMenu {
	return new(DailyRecommendSongsMenu)
}

func (m *DailyRecommendSongsMenu) MenuData() interface{} {
	return m.songs
}

func (m *DailyRecommendSongsMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *DailyRecommendSongsMenu) IsPlayable() bool {
	return true
}

func (m *DailyRecommendSongsMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *DailyRecommendSongsMenu) GetMenuKey() string {
	return "daily_songs"
}

func (m *DailyRecommendSongsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DailyRecommendSongsMenu) SubMenu(*NeteaseModel, int) IMenu {
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
		if utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		// 不重复请求
		if len(m.menus) > 0 && len(m.songs) > 0 {
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
		for _, song := range m.songs {
			var artists []string
			for _, artist := range song.Artists {
				artists = append(artists, artist.Name)
			}
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(song.Name), utils.ReplaceSpecialStr(strings.Join(artists, ","))})
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

