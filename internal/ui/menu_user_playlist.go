package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

const CurUser int64 = 0

type UserPlaylistMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []structs.Playlist
	userId    int64
	offset    int
	limit     int
	hasMore   bool
}

func NewUserPlaylistMenu(base baseMenu, userId int64) *UserPlaylistMenu {
	return &UserPlaylistMenu{
		baseMenu: base,
		userId:   userId,
		offset:   0,
		limit:    100,
	}
}

func (m *UserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *UserPlaylistMenu) GetMenuKey() string {
	return fmt.Sprintf("user_playlist_%d", m.userId)
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
	return NewPlaylistDetailMenu(m.baseMenu, m.playlists[index].Id)
}

// TODO optimize
func getUserPlaylists(userId int64, limit int, offset int) (codeType utils.ResCode, playlists []structs.Playlist, hasMore bool) {
	userPlaylists := service.UserPlaylistService{
		Uid:    strconv.FormatInt(userId, 10),
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}
	code, response := userPlaylists.UserPlaylist()
	codeType = utils.CheckCode(code)
	if codeType != utils.Success {
		return
	}

	playlists = utils.GetPlaylists(response)
	menus := make([]model.MenuItem, len(playlists))
	for i := range playlists {
		menus[i] = model.MenuItem{Title: utils.ReplaceSpecialStr(playlists[i].Name)}
	}

	// 是否有更多
	var err error
	if hasMore, err = jsonparser.GetBoolean(response, "more"); err != nil {
		hasMore = false
	}

	return
}

func (m *UserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 等于0，获取当前用户歌单
		if m.userId == CurUser && utils.CheckUserInfo(m.netease.user) == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		userId := m.userId
		if m.userId == CurUser {
			// 等于0，获取当前用户歌单
			userId = m.netease.user.UserId
		}

		codeType, playlists, hasMore := getUserPlaylists(userId, m.limit, m.offset)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		m.playlists = playlists
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
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
		userId := m.userId
		if m.userId == CurUser {
			// 等于0，获取当前用户歌单
			userId = m.netease.user.UserId
		}

		m.offset = m.offset + len(m.menus)
		codeType, playlists, hasMore := getUserPlaylists(userId, m.limit, m.offset)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(func() model.Page {
				main.RefreshMenuList()
				return nil
			})
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		m.playlists = append(m.playlists, playlists...)
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus
		m.hasMore = hasMore

		return true, nil
	}
}
