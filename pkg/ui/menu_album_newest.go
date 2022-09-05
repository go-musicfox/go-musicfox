package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strings"
)

type AlbumNewestMenu struct {
	menus  []MenuItem
	albums []structs.Album
}

func NewAlbumNewestMenu() *AlbumNewestMenu {
	return new(AlbumNewestMenu)
}

func (m *AlbumNewestMenu) MenuData() interface{} {
	return m.albums
}

func (m *AlbumNewestMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *AlbumNewestMenu) IsPlayable() bool {
	return false
}

func (m *AlbumNewestMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *AlbumNewestMenu) GetMenuKey() string {
	return "album_new_hot"
}

func (m *AlbumNewestMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumNewestMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.albums[index].Id)
}

func (m *AlbumNewestMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *AlbumNewestMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *AlbumNewestMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true
		}

		albumService := service.AlbumNewestService{}
		code, response := albumService.AlbumNewest()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		m.albums = utils.GetNewAlbums(response)

		for _, album := range m.albums {
			var artists []string
			for _, artist := range album.Artists {
				artists = append(artists, artist.Name)
			}
			artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(album.Name), utils.ReplaceSpecialStr(artistsStr)})
		}

		return true
	}
}

func (m *AlbumNewestMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *AlbumNewestMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}
