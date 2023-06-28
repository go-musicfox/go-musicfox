package ui

import (
	"fmt"
	"strings"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type AlbumNewestMenu struct {
	DefaultMenu
	menus  []MenuItem
	albums []structs.Album
}

func NewAlbumNewestMenu() *AlbumNewestMenu {
	return new(AlbumNewestMenu)
}

func (m *AlbumNewestMenu) IsSearchable() bool {
	return true
}

func (m *AlbumNewestMenu) GetMenuKey() string {
	return "album_new_hot"
}

func (m *AlbumNewestMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumNewestMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.albums[index].Id)
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
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		return true
	}
}

func (m *AlbumNewestMenu) Albums() []structs.Album {
	return m.albums
}
