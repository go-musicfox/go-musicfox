package ui

import (
	"github.com/anhoder/netease-music/service"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"
)

type RecentSongsMenu struct {
	DefaultMenu
	menus []MenuItem
	songs []structs.Song
}

func NewRecentSongsMenu() *RecentSongsMenu {
	return new(RecentSongsMenu)
}

func (m *RecentSongsMenu) IsSearchable() bool {
	return true
}

func (m *RecentSongsMenu) IsPlayable() bool {
	return true
}

func (m *RecentSongsMenu) GetMenuKey() string {
	return "recent_songs"
}

func (m *RecentSongsMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *RecentSongsMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		recentSongService := service.RecordRecentSongsService{}
		code, response := recentSongService.RecordRecentSongs()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetRecentSongs(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *RecentSongsMenu) Songs() []structs.Song {
	return m.songs
}
