package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type AddToUserPlaylistMenu struct {
	DefaultMenu
	menus     []MenuItem
	playlists []structs.Playlist
	song      structs.Song
	userId    int64
	offset    int
	limit     int
	hasMore   bool
	action    bool // true for add, false for del
}

func NewAddToUserPlaylistMenu(userId int64, song structs.Song, action bool) *AddToUserPlaylistMenu {
	return &AddToUserPlaylistMenu{
		userId: userId,
		offset: 0,
		limit:  100,
		action: action,
		song:   song,
	}
}

func (m *AddToUserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *AddToUserPlaylistMenu) GetMenuKey() string {
	return fmt.Sprintf("add_to_user_playlist_%d", m.userId)
}

func (m *AddToUserPlaylistMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AddToUserPlaylistMenu) Playlists() []structs.Playlist {
	return m.playlists
}

func (m *AddToUserPlaylistMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	return nil
}

func (m *AddToUserPlaylistMenu) BeforeEnterMenuHook() Hook {
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

		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(userId, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		} else if codeType != utils.Success {
			return false
		}

		var menus []MenuItem
		m.playlists = utils.GetPlaylists(response)
		for _, playlist := range m.playlists {
			menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true
	}
}

func (m *AddToUserPlaylistMenu) BottomOutHook() Hook {
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
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(userId, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, nil)
			return false
		} else if codeType != utils.Success {
			return false
		}

		list := utils.GetPlaylists(response)
		for _, playlist := range list {
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}

		m.playlists = append(m.playlists, list...)

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true
	}
}
