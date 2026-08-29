package album

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type AlbumDetailMenu struct {
	ui.BaseMenu
	menus   []model.MenuItem
	songs   []structs.Song
	albumID int64
}

func NewAlbumDetailMenu(base ui.BaseMenu, albumID int64) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		BaseMenu: base,
		albumID:  albumID,
	}
}

func (m *AlbumDetailMenu) IsSearchable() bool {
	return true
}

func (m *AlbumDetailMenu) IsPlayable() bool {
	return true
}

func (m *AlbumDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("album_detail_%d", m.albumID)
}

func (m *AlbumDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		albumService := service.AlbumService{
			ID: strconv.FormatInt(m.albumID, 10),
		}
		code, response := albumService.Album()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
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

func (m *AlbumDetailMenu) AlbumID() int64 {
	return m.albumID
}
