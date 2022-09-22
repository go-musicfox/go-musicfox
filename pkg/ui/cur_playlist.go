package ui

import (
	"go-musicfox/pkg/structs"
)

const CurPlaylistKey = "cur_playlist"

type CurPlaylist struct {
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

func (m *CurPlaylist) BeforeBackMenuHook() Hook {
	return nil
}

func (m *CurPlaylist) IsPlayable() bool {
	return true
}

func (m *CurPlaylist) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *CurPlaylist) GetMenuKey() string {
	return CurPlaylistKey
}

func (m *CurPlaylist) MenuViews() []MenuItem {
	return m.menus
}

func (m *CurPlaylist) SubMenu(_ *NeteaseModel, _ int) IMenu {
	return nil
}

func (m *CurPlaylist) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *CurPlaylist) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *CurPlaylist) BeforeEnterMenuHook() Hook {
	// Nothing to do
	return nil
}

func (m *CurPlaylist) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *CurPlaylist) TopOutHook() Hook {
	// Nothing to do
	return nil
}
