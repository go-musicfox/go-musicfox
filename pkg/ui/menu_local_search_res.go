package ui

import (
	"strings"

	ds2 "go-musicfox/pkg/structs"
)

type searchRes struct {
	item  MenuItem
	index int
}

type LocalSearchMenu struct {
	Menu
	resItems []searchRes
}

func NewSearchMenu(originMenu Menu, search string) *LocalSearchMenu {
	menu := &LocalSearchMenu{
		Menu: originMenu,
	}

	for i, item := range originMenu.MenuViews() {
		if strings.Contains(item.Title, search) || strings.Contains(item.Subtitle, search) {
			menu.resItems = append(menu.resItems, searchRes{
				item:  item,
				index: i,
			})
		}
	}
	return menu
}

func (m *LocalSearchMenu) IsLocatable() bool {
	return false
}

func (m *LocalSearchMenu) MenuViews() []MenuItem {
	var items []MenuItem
	for _, item := range m.resItems {
		items = append(items, item.item)
	}
	return items
}

func (m *LocalSearchMenu) SubMenu(model *NeteaseModel, index int) Menu {
	if index > len(m.resItems)-1 {
		return nil
	}

	return m.Menu.SubMenu(model, m.resItems[index].index)
}

func (m *LocalSearchMenu) RealDataIndex(index int) int {
	if index > len(m.resItems)-1 {
		return 0
	}

	return m.resItems[index].index
}

func (m *LocalSearchMenu) BottomOutHook() Hook {
	return nil
}

func (m *LocalSearchMenu) TopOutHook() Hook {
	return nil
}

func (m *LocalSearchMenu) Songs() []ds2.Song {
	if menu, ok := m.Menu.(SongsMenu); ok {
		return menu.Songs()
	}
	return nil
}

func (m *LocalSearchMenu) Playlists() []ds2.Playlist {
	if menu, ok := m.Menu.(PlaylistsMenu); ok {
		return menu.Playlists()
	}
	return nil
}

func (m *LocalSearchMenu) Albums() []ds2.Album {
	if menu, ok := m.Menu.(AlbumsMenu); ok {
		return menu.Albums()
	}
	return nil
}

func (m *LocalSearchMenu) Artists() []ds2.Artist {
	if menu, ok := m.Menu.(ArtistsMenu); ok {
		return menu.Artists()
	}
	return nil
}
