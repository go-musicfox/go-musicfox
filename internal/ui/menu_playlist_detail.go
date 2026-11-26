package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/netease"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

var fetchSongsOfPlaylist = netease.FetchSongsOfPlaylist

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
		codeType, songs := fetchSongsOfPlaylist(m.playlistId, configs.AppConfig.Player.ShowAllSongsOfPlaylist)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			notify.Notify(notify.NotifyContent{
				Title:   model.T(MsgOperationLoadPlaylistFailed),
				Text:    playlistLoadErrorText(codeType),
				GroupId: types.GroupID,
				Level:   notify.ToastError,
			})
			return false, nil
		}
		m.songs = songs
		m.menus = menux.GetViewFromSongs(songs)

		return true, nil
	}
}

func playlistLoadErrorText(codeType _struct.ResCode) string {
	switch codeType {
	case _struct.NetworkError:
		return model.T(MsgSearchNetworkError)
	case _struct.TooManyRequests:
		return model.T(MsgLoginTooManyRequests)
	default:
		return model.T(MsgSearchUnknownError)
	}
}

func (m *PlaylistDetailMenu) Songs() []structs.Song {
	return m.songs
}

func (m *PlaylistDetailMenu) PlaylistId() int64 {
	return m.playlistId
}
