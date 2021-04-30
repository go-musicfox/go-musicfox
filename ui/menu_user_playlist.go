package ui

import (
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"go-musicfox/ds"
	"go-musicfox/utils"
	"strconv"
)

type UserPlaylistMenu struct {
	menus     []MenuItem
	playlists []ds.Playlist
	offset    int
	limit     int
	hasMore   bool
}

func NewUserPlaylistMenu() *UserPlaylistMenu {
	return &UserPlaylistMenu{
		offset: 0,
		limit: 100,
	}
}

func (m *UserPlaylistMenu) MenuData() interface{} {
	return m.playlists
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
	if len(m.playlists) < index {
		return nil
	}
	return &PlaylistDetailMenu{PlaylistId: m.playlists[index].Id}
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

		if len(m.menus) > 0 && len(m.playlists) > 0 {
			return true
		}

		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(model.user.UserId, 10),
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
			menus = append(menus, MenuItem{utils.ReplaceSpecialStr(playlist.Name), ""})
		}
		m.menus = menus

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
			Uid:    strconv.FormatInt(model.user.UserId, 10),
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
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(playlist.Name), ""})
		}

		m.playlists = append(m.playlists, list...)

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

