package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// CurUser is the sentinel user ID meaning "the current logged-in user"
// (userID == 0). The user_playlist menu (now in the internal/plugins/playlist
// plugin) and this menu resolve it to svc.User().UserId at runtime; the
// plugin's parameterized main-menu entry builds user_playlist with CurUser,
// exactly like the built-in entry used to.
const CurUser int64 = 0

type AddToUserPlaylistMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
	song      structs.Song
	userID    int64
	offset    int
	limit     int
	hasMore   bool
	action    bool // true for add, false for del
}

func NewAddToUserPlaylistMenu(base baseMenu, userID int64, song structs.Song, action bool) *AddToUserPlaylistMenu {
	return &AddToUserPlaylistMenu{
		baseMenu: base,
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

func (m *AddToUserPlaylistMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *AddToUserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 等于0，获取当前用户歌单
		if m.userID == CurUser && _struct.CheckUserInfo(m.svc.User()) == _struct.NeedLogin {
			page, _ := m.svc.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		userID := m.userID
		if m.userID == CurUser {
			// 等于0，获取当前用户歌单
			userID = m.svc.User().UserId
		}

		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(userID, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.svc.ToLoginPage(EnterMenuCallback(main))
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
		if m.userID == CurUser {
			// 等于0，获取当前用户歌单
			userID = m.svc.User().UserId
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
			page, _ := m.svc.ToLoginPage(nil)
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
