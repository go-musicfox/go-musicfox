package ui

import (
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type PersonalFmMenu struct {
	DefaultMenu
	menus []MenuItem
	songs []structs.Song
}

func NewPersonalFmMenu() *PersonalFmMenu {
	return new(PersonalFmMenu)
}

func (m *PersonalFmMenu) IsSearchable() bool {
	return true
}

func (m *PersonalFmMenu) IsPlayable() bool {
	return true
}

func (m *PersonalFmMenu) ResetPlaylistWhenPlay() bool {
	return true
}

func (m *PersonalFmMenu) GetMenuKey() string {
	return "personal_fm"
}

func (m *PersonalFmMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *PersonalFmMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		// 已有数据
		if len(m.menus) > 0 && len(m.songs) > 0 {
			return true
		}

		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		// 响应中获取数据
		m.songs = utils.GetFmSongs(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *PersonalFmMenu) BottomOutHook() Hook {
	return func(model *NeteaseModel) bool {
		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		songs := utils.GetFmSongs(response)
		menus := GetViewFromSongs(songs)

		m.menus = append(m.menus, menus...)
		m.songs = append(m.songs, songs...)
		model.player.playlist = m.songs
		model.player.playlistUpdateAt = time.Now()

		return true
	}
}

func (m *PersonalFmMenu) Songs() []structs.Song {
	return m.songs
}
