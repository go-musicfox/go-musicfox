package song

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// ui.CurUser is the sentinel user ID meaning "the current logged-in user"
// (userID == 0). It stays defined in ui (menu_add_to_user_playlist.go, the
// user_playlist plugin entry and operate.go use it); this menu references
// ui.CurUser and resolves it to User().UserId at runtime.

type AddToUserPlaylistMenu struct {
	ui.BaseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
	song      structs.Song
	userID    int64
	offset    int
	limit     int
	hasMore   bool
	action    bool // true for add, false for del
}

func NewAddToUserPlaylistMenu(base ui.BaseMenu, userID int64, song structs.Song, action bool) *AddToUserPlaylistMenu {
	return &AddToUserPlaylistMenu{
		BaseMenu: base,
		userID:   userID,
		offset:   0,
		limit:    100,
		action:   action,
		song:     song,
	}
}

func (m *AddToUserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *AddToUserPlaylistMenu) GetMenuKey() string {
	return fmt.Sprintf("add_to_user_playlist_%d", m.userID)
}

func (m *AddToUserPlaylistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AddToUserPlaylistMenu) Playlists() []structs.Playlist {
	return m.playlists
}

// Song returns the song payload of the add/remove operation. ui's
// event_handler.go / operate.go reach it through the
// AddToUserPlaylistGetter interface.
func (m *AddToUserPlaylistMenu) Song() structs.Song {
	return m.song
}

// IsAdd reports whether the menu runs an add (true) or remove (false). The
// name avoids colliding with model.Menu.Action(app, index), which every menu
// must also implement.
func (m *AddToUserPlaylistMenu) IsAdd() bool {
	return m.action
}

func (m *AddToUserPlaylistMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *AddToUserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 等于0，获取当前用户歌单
		if m.userID == ui.CurUser && _struct.CheckUserInfo(m.User()) == _struct.NeedLogin {
			page, _ := m.ToLoginPage(ui.EnterMenuCallback(main))
			return false, page
		}

		userID := m.userID
		if m.userID == ui.CurUser {
			// 等于0，获取当前用户歌单
			userID = m.User().UserId
		}

		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(userID, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(ui.EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		var menus []model.MenuItem
		m.playlists = _struct.GetPlaylists(response)
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true, nil
	}
}

func (m *AddToUserPlaylistMenu) BottomOutHook() model.Hook {
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
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(userID, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(nil)
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		list := _struct.GetPlaylists(response)
		for _, playlist := range list {
			m.menus = append(m.menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
		}

		m.playlists = append(m.playlists, list...)

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true, nil
	}
}
