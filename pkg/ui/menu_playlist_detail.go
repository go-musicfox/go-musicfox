package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
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

func getSongsInPlaylist(playlistId int64, getAll bool) (codeType utils.ResCode, songs []structs.Song) {
	var (
		code     float64
		response []byte
	)
	if !getAll {
		playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
		code, response = playlistDetail.PlaylistDetail()
	} else {
		allTrack := service.PlaylistTrackAllService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
		code, response = allTrack.AllTracks()
	}
	codeType = utils.CheckCode(code)
	if codeType != utils.Success {
		return
	}
	songs = utils.GetSongsOfPlaylist(response)

	return
}

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		codeType, songs := getSongsInPlaylist(m.playlistId, configs.ConfigRegistry.MainShowAllSongsOfPlaylist)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(main.EnterMenu)
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}
		m.songs = songs
		m.menus = utils.GetViewFromSongs(songs)

		return true, nil
	}
}

func (m *PlaylistDetailMenu) Songs() []structs.Song {
	return m.songs
}
