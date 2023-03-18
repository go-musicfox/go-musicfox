package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type ArtistAlbumMenu struct {
	DefaultMenu
	menus    []MenuItem
	albums   []structs.Album
	artistId int64
}

func NewArtistAlbumMenu(artistId int64) *ArtistAlbumMenu {
	return &ArtistAlbumMenu{
		artistId: artistId,
	}
}

func (m *ArtistAlbumMenu) IsSearchable() bool {
	return true
}

func (m *ArtistAlbumMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_album_%d", m.artistId)
}

func (m *ArtistAlbumMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *ArtistAlbumMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.albums[index].Id)
}

func (m *ArtistAlbumMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		artistAlbumService := service.ArtistAlbumService{
			ID:     strconv.FormatInt(m.artistId, 10),
			Offset: "0",
			Limit:  "50",
		}
		code, response := artistAlbumService.ArtistAlbum()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		m.albums = utils.GetArtistHotAlbums(response)
		m.menus = GetViewFromAlbums(m.albums)

		return true
	}
}

func (m *ArtistAlbumMenu) Albums() []structs.Album {
	return m.albums
}
