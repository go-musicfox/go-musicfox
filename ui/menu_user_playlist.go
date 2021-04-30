package ui

import (
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"go-musicfox/ds"
	"go-musicfox/utils"
	"strconv"
)

type UserPlaylistMenu struct {
	menus   []MenuItem
	uid     string
	offset  int
	limit   int
	hasMore bool
}

func NewUserPlaylistMenu(user *ds.User) *UserPlaylistMenu {
	if user == nil {
		return &UserPlaylistMenu{}
	}

	return &UserPlaylistMenu{
		uid: strconv.FormatInt(user.UserId, 10),
		offset: 0,
		limit: 100,
	}
}

func (m *UserPlaylistMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *UserPlaylistMenu) IsPlayable() bool {
	return false
}

func (m *UserPlaylistMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *UserPlaylistMenu) GetMenuKey() string {
	return "my_playlist"
}

func (m *UserPlaylistMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *UserPlaylistMenu) SubMenu(model *NeteaseModel, index int) IMenu {
	playlists, ok := model.menuData.([]ds.Playlist)
	if !ok || len(playlists) < index {
		return nil
	}
	return &PlaylistDetailMenu{PlaylistId: playlists[index].Id}
}

func (m *UserPlaylistMenu) ExtraView() string {
	return ""
}

func (m *UserPlaylistMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *UserPlaylistMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *UserPlaylistMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if utils.CheckUserInfo(model.user) == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		userPlaylists := service.UserPlaylistService{
			Uid:    m.uid,
			Limit: strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, enterMenu)
			return false
		}

		list := utils.GetPlaylists(response)
		for _, playlist := range list {
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(playlist.Name), ""})
		}

		model.menuData = list

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true
	}
}

func (m *UserPlaylistMenu) BottomOutHook() Hook {
	if !m.hasMore{
		return nil
	}
	return func(model *NeteaseModel) bool {
		m.offset = m.offset + len(m.menus)
		userPlaylists := service.UserPlaylistService{
			Uid:    m.uid,
			Limit: strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(model, nil)
			return false
		}
		list := utils.GetPlaylists(response)
		for _, playlist := range list {
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(playlist.Name), ""})
		}

		if menuData, ok := model.menuData.([]ds.Playlist); ok {
			model.menuData = append(menuData, list...)
		}

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true
	}
}

func (m *UserPlaylistMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}

