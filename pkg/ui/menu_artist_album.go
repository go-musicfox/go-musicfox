package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
)

type ArtistAlbumMenu struct {
    menus    []MenuItem
    albums   []structs.Album
    artistId int64
}

func NewArtistAlbumMenu(artistId int64) *ArtistAlbumMenu {
    return &ArtistAlbumMenu{
        artistId: artistId,
    }
}

func (m *ArtistAlbumMenu) MenuData() interface{} {
    return m.albums
}

func (m *ArtistAlbumMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *ArtistAlbumMenu) IsPlayable() bool {
    return false
}

func (m *ArtistAlbumMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *ArtistAlbumMenu) GetMenuKey() string {
    return fmt.Sprintf("artist_album_%d", m.artistId)
}

func (m *ArtistAlbumMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *ArtistAlbumMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if len(m.albums) < index {
        return nil
    }

    return NewAlbumDetailMenu(m.albums[index].Id)
}

func (m *ArtistAlbumMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *ArtistAlbumMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
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

func (m *ArtistAlbumMenu) BottomOutHook() Hook {
    return nil
}

func (m *ArtistAlbumMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
