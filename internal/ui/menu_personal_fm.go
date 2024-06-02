package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type PersonalFmMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewPersonalFmMenu(base baseMenu) *PersonalFmMenu {
	return &PersonalFmMenu{
		baseMenu: base,
	}
}

func (m *PersonalFmMenu) IsSearchable() bool {
	return true
}

func (m *PersonalFmMenu) IsPlayable() bool {
	return true
}

func (m *PersonalFmMenu) GetMenuKey() string {
	return "personal_fm"
}

func (m *PersonalFmMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *PersonalFmMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 已有数据
		if len(m.menus) > 0 && len(m.songs) > 0 {
			return true, nil
		}

		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		// 响应中获取数据
		m.songs = _struct.GetFmSongs(response)
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *PersonalFmMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		songs := _struct.GetFmSongs(response)
		menus := menux.GetViewFromSongs(songs)

		m.menus = append(m.menus, menus...)
		m.songs = append(m.songs, songs...)
		m.netease.player.playlist = m.songs
		m.netease.player.playlistUpdateAt = time.Now()

		return true, nil
	}
}

func (m *PersonalFmMenu) Songs() []structs.Song {
	return m.songs
}
