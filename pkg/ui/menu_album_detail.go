package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type AlbumDetailMenu struct {
	DefaultMenu
	menus   []MenuItem
	songs   []structs.Song
	albumId int64
}

func NewAlbumDetailMenu(albumId int64) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		albumId: albumId,
	}
}

func (m *AlbumDetailMenu) IsSearchable() bool {
	return true
}

func (m *AlbumDetailMenu) IsPlayable() bool {
	return true
}

func (m *AlbumDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("album_detail_%d", m.albumId)
}

func (m *AlbumDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		albumService := service.AlbumService{
			ID: strconv.FormatInt(m.albumId, 10),
		}
		code, response := albumService.Album()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}

		m.songs = utils.GetSongsOfAlbum(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *AlbumDetailMenu) Songs() []structs.Song {
	return m.songs
}
