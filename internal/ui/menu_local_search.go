package ui

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type LocalSearchMenu struct {
	model.LocalSearchMenuImpl
	base baseMenu
}

func NewLocalSearchMenu(netease *Netease) *LocalSearchMenu {
	return &LocalSearchMenu{
		base:                newBaseMenu(netease),
		LocalSearchMenuImpl: *model.DefaultSearchMenu(),
	}
}

func (m *LocalSearchMenu) Songs() []structs.Song {
	if menu, ok := m.Menu.(SongsMenu); ok {
		return menu.Songs()
	}
	return nil
}

func (m *LocalSearchMenu) Playlists() []structs.Playlist {
	if menu, ok := m.Menu.(PlaylistsMenu); ok {
		return menu.Playlists()
	}
	return nil
}

func (m *LocalSearchMenu) Albums() []structs.Album {
	if menu, ok := m.Menu.(AlbumsMenu); ok {
		return menu.Albums()
	}
	return nil
}

func (m *LocalSearchMenu) Artists() []structs.Artist {
	if menu, ok := m.Menu.(ArtistsMenu); ok {
		return menu.Artists()
	}
	return nil
}

func (m *LocalSearchMenu) IsPlayable() bool {
	me, ok := m.Menu.(Menu)
	return ok && me.IsPlayable()
}

func (m *LocalSearchMenu) IsLocatable() bool {
	return false
}

func (m *LocalSearchMenu) IsSearchable() bool {
	return false
}
