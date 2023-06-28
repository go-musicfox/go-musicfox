package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/structs"
)

const CurPlaylistKey = "cur_playlist"

type CurPlaylist struct {
	DefaultMenu
	menus []MenuItem
	songs []structs.Song
}

func NewCurPlaylist(songs []structs.Song) *CurPlaylist {
	return &CurPlaylist{
		songs: songs,
		menus: GetViewFromSongs(songs),
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

func (m *CurPlaylist) MenuViews() []MenuItem {
	return m.menus
}

func (m *CurPlaylist) Songs() []structs.Song {
	return m.songs
}

func (m *CurPlaylist) BottomOutHook() Hook {
	return func(model *NeteaseModel) bool {
		if model.player.playingMenu == nil || model.player.playingMenu.GetMenuKey() == CurPlaylistKey {
			return true
		}
		hook := model.player.playingMenu.BottomOutHook()
		if hook == nil {
			return true
		}
		res := hook(model)
		m.songs = model.player.playlist
		m.menus = GetViewFromSongs(m.songs)
		return res
	}
}
