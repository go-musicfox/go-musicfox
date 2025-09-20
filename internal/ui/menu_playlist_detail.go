package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/netease"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type PlaylistDetailMenu struct {
	baseMenu
	menus      []model.MenuItem
	songs      []structs.Song
	playlistId int64
}

func NewPlaylistDetailMenu(base baseMenu, playlistId int64) *PlaylistDetailMenu {
	return &PlaylistDetailMenu{
		baseMenu:   base,
		playlistId: playlistId,
	}
}

func (m *PlaylistDetailMenu) IsSearchable() bool {
	return true
}

func (m *PlaylistDetailMenu) IsPlayable() bool {
	return true
}

func (m *PlaylistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("playlist_detail_%d", m.playlistId)
}

func (m *PlaylistDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *PlaylistDetailMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		codeType, songs := netease.FetchSongsOfPlaylist(m.playlistId, configs.AppConfig.Player.ShowAllSongsOfPlaylist)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}
		m.songs = songs
		m.menus = menux.GetViewFromSongs(songs)

		return true, nil
	}
}

func (m *PlaylistDetailMenu) Songs() []structs.Song {
	return m.songs
}

func (m *PlaylistDetailMenu) PlaylistId() int64 {
	return m.playlistId
}
