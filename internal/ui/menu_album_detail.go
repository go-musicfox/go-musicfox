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

type AlbumDetailMenu struct {
	baseMenu
	menus   []model.MenuItem
	songs   []structs.Song
	albumId int64
}

func NewAlbumDetailMenu(base baseMenu, albumId int64) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		baseMenu: base,
		albumId:  albumId,
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

func (m *AlbumDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		albumService := service.AlbumService{
			ID: strconv.FormatInt(m.albumId, 10),
		}
		code, response := albumService.Album()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		m.songs = _struct.GetSongsOfAlbum(response)
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *AlbumDetailMenu) Songs() []structs.Song {
	return m.songs
}

func (m *AlbumDetailMenu) AlbumId() int64 {
	return m.albumId
}
