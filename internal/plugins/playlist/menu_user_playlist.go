package playlist

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/netease"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type UserPlaylistMenu struct {
	ui.BaseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
	userID    int64
	offset    int
	limit     int
	hasMore   bool
}

func NewUserPlaylistMenu(base ui.BaseMenu, userID int64) *UserPlaylistMenu {
	return &UserPlaylistMenu{
		BaseMenu: base,
		userID:   userID,
		offset:   0,
		limit:    100,
	}
}

func (m *UserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *UserPlaylistMenu) GetMenuKey() string {
	return fmt.Sprintf("user_playlist_%d", m.userID)
}

func (m *UserPlaylistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *UserPlaylistMenu) Playlists() []structs.Playlist {
	return m.playlists
}

func (m *UserPlaylistMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.playlists) < index {
		return nil
	}
	playlistMenu, err := ui.BuildMenu("playlist_detail", m.BaseMenu, ui.PlaylistDetailOpts{PlaylistID: m.playlists[index].Id})
	if err != nil {
		return nil
	}
	return playlistMenu
}

func (m *UserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 等于0，获取当前用户歌单
		if m.userID == ui.CurUser && _struct.CheckUserInfo(m.User()) == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		}

		userID := m.userID
		if m.userID == ui.CurUser {
			// 等于0，获取当前用户歌单
			userID = m.User().UserId
		}

		codeType, playlists, hasMore := netease.FetchUserPlaylists(userID, m.limit, m.offset)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		m.playlists = playlists
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus
		m.hasMore = hasMore

		return true, nil
	}
}

func (m *UserPlaylistMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		userID := m.userID
		if m.userID == ui.CurUser {
			// 等于0，获取当前用户歌单
			userID = m.User().UserId
		}

		m.offset = m.offset + len(m.menus)
		codeType, playlists, hasMore := netease.FetchUserPlaylists(userID, m.limit, m.offset)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(func() model.Page {
				main.RefreshMenuList()
				return nil
			})
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		m.playlists = append(m.playlists, playlists...)
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus
		m.hasMore = hasMore

		return true, nil
	}
}
