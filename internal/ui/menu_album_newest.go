package ui

import (
	"fmt"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type AlbumNewestMenu struct {
	baseMenu
	menus  []model.MenuItem
	albums []structs.Album
}

func NewAlbumNewestMenu(base baseMenu) *AlbumNewestMenu {
	return &AlbumNewestMenu{
		baseMenu: base,
	}
}

func (m *AlbumNewestMenu) IsSearchable() bool {
	return true
}

func (m *AlbumNewestMenu) GetMenuKey() string {
	return "album_new_hot"
}

func (m *AlbumNewestMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumNewestMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *AlbumNewestMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true, nil
		}

		albumService := service.AlbumNewestService{}
		code, response := albumService.AlbumNewest()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.albums = _struct.GetNewAlbums(response)

		for _, album := range m.albums {
			var artists []string
			for _, artist := range album.Artists {
				artists = append(artists, artist.Name)
			}
			artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
			m.menus = append(m.menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(album.Name), Subtitle: _struct.ReplaceSpecialStr(artistsStr)})
		}

		return true, nil
	}
}

func (m *AlbumNewestMenu) Albums() []structs.Album {
	return m.albums
}
