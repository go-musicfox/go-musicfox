package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/ds"
	"go-musicfox/utils"
	"strconv"
	"strings"
)

type AlbumDetailMenu struct {
	menus   []MenuItem
	songs   []ds.Song
	AlbumId int64
}

func NewAlbumDetailMenu(albumId int64) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		AlbumId: albumId,
	}
}

func (m *AlbumDetailMenu) MenuData() interface{} {
	return m.songs
}

func (m *AlbumDetailMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *AlbumDetailMenu) IsPlayable() bool {
	return true
}

func (m *AlbumDetailMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *AlbumDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("album_detail_%d", m.AlbumId)
}

func (m *AlbumDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
	return nil
}

func (m *AlbumDetailMenu) ExtraView() string {
	return ""
}

func (m *AlbumDetailMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *AlbumDetailMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
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
		var menus []MenuItem
		for _, song := range m.songs {
			var artists []string
			for _, artist := range song.Artists {
				artists = append(artists, artist.Name)
			}
			menus = append(menus, MenuItem{utils.ReplaceSpecialStr(song.Name), utils.ReplaceSpecialStr(strings.Join(artists, ","))})
		}
		m.menus = menus

		return true
	}
}

func (m *AlbumDetailMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *AlbumDetailMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}
