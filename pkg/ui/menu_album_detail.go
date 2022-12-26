package ui

import (
	"fmt"
	"strconv"

	"go-musicfox/pkg/structs"
	"go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type AlbumDetailMenu struct {
	DefaultMenu
	menus   []MenuItem
	songs   []structs.Song
	AlbumId int64
}

func NewAlbumDetailMenu(albumId int64) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		AlbumId: albumId,
	}
}

func (m *AlbumDetailMenu) IsSearchable() bool {
	return true
}

func (m *AlbumDetailMenu) MenuData() interface{} {
	return m.songs
}

func (m *AlbumDetailMenu) IsPlayable() bool {
	return true
}

func (m *AlbumDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("album_detail_%d", m.AlbumId)
}

func (m *AlbumDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		albumService := service.AlbumService{
			ID: strconv.FormatInt(m.AlbumId, 10),
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
