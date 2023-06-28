package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

const CurUser int64 = 0

type UserPlaylistMenu struct {
	DefaultMenu
	menus     []MenuItem
	playlists []structs.Playlist
	userId    int64
	offset    int
	limit     int
	hasMore   bool
}

func NewUserPlaylistMenu(userId int64) *UserPlaylistMenu {
	return &UserPlaylistMenu{
		userId: userId,
		offset: 0,
		limit:  100,
	}
}

func (m *UserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *UserPlaylistMenu) GetMenuKey() string {
	return fmt.Sprintf("user_playlist_%d", m.userId)
}

func (m *UserPlaylistMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *UserPlaylistMenu) Playlists() []structs.Playlist {
	return m.playlists
}

func (m *UserPlaylistMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if len(m.playlists) < index {
		return nil
	}
	return NewPlaylistDetailMenu(m.playlists[index].Id)
}

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
	menus := make([]MenuItem, len(playlists))
	for i := range playlists {
		menus[i] = MenuItem{Title: utils.ReplaceSpecialStr(playlists[i].Name)}
	}

	// 是否有更多
	var err error = nil
	if hasMore, err = jsonparser.GetBoolean(response, "more"); err != nil {
		hasMore = false
	}

	return
}

func (m *UserPlaylistMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		// 等于0，获取当前用户歌单
		if m.userId == CurUser && utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		userId := m.userId
		if m.userId == CurUser {
			// 等于0，获取当前用户歌单
			userId = model.user.UserId
		}

		codeType, playlists, hasMore := getUserPlaylists(userId, m.limit, m.offset)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}

		m.playlists = playlists
		var menus []MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus
		m.hasMore = hasMore

		return true
	}
}

func (m *UserPlaylistMenu) BottomOutHook() Hook {
	if !m.hasMore {
		return nil
	}
	return func(model *NeteaseModel) bool {
		userId := m.userId
		if m.userId == CurUser {
			// 等于0，获取当前用户歌单
			userId = model.user.UserId
		}

		m.offset = m.offset + len(m.menus)
		codeType, playlists, hasMore := getUserPlaylists(userId, m.limit, m.offset)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}

		m.playlists = playlists
		var menus []MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus
		m.hasMore = hasMore

		return true
	}
}
