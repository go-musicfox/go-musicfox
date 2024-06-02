package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type ArtistAlbumMenu struct {
	baseMenu
	menus    []model.MenuItem
	albums   []structs.Album
	artistId int64
}

func NewArtistAlbumMenu(base baseMenu, artistId int64) *ArtistAlbumMenu {
	return &ArtistAlbumMenu{
		baseMenu: base,
		artistId: artistId,
	}
}

func (m *ArtistAlbumMenu) IsSearchable() bool {
	return true
}

func (m *ArtistAlbumMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_album_%d", m.artistId)
}

func (m *ArtistAlbumMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistAlbumMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *ArtistAlbumMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		artistAlbumService := service.ArtistAlbumService{
			ID:     strconv.FormatInt(m.artistId, 10),
			Offset: "0",
			Limit:  "50",
		}
		code, response := artistAlbumService.ArtistAlbum()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.albums = _struct.GetArtistHotAlbums(response)
		m.menus = menux.GetViewFromAlbums(m.albums)

		return true, nil
	}
}

func (m *ArtistAlbumMenu) Albums() []structs.Album {
	return m.albums
}
