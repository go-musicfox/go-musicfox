package ui

import (
	"go-musicfox/pkg/structs"
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

func (m *CurPlaylist) MenuData() interface{} {
	return m.songs
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
