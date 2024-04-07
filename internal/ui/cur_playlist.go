package ui

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
)

const CurPlaylistKey = "cur_playlist"

type CurPlaylist struct {
	baseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewCurPlaylist(base baseMenu, songs []structs.Song) *CurPlaylist {
	return &CurPlaylist{
		baseMenu: base,
		songs:    songs,
		menus:    utils.GetViewFromSongs(songs),
	}
}

func (m *CurPlaylist) IsSearchable() bool {
	return true
}

func (m *CurPlaylist) IsPlayable() bool {
	return true
}

func (m *CurPlaylist) GetMenuKey() string {
	return CurPlaylistKey
}

func (m *CurPlaylist) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *CurPlaylist) Songs() []structs.Song {
	return m.songs
}

func (m *CurPlaylist) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.netease.player.playingMenu == nil || m.netease.player.playingMenu.GetMenuKey() == CurPlaylistKey {
			return true, nil
		}
		hook := m.netease.player.playingMenu.BottomOutHook()
		if hook == nil {
			return true, nil
		}
		res, page := hook(main)
		m.songs = m.netease.player.playlist
		m.menus = utils.GetViewFromSongs(m.songs)
		return res, page
	}
}
