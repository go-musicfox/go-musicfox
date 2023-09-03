package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
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
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		m.songs = utils.GetSongsOfAlbum(response)
		m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *AlbumDetailMenu) Songs() []structs.Song {
	return m.songs
}
