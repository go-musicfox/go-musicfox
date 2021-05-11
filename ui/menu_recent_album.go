package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"go-musicfox/ds"
	"go-musicfox/utils"
	"strconv"
	"strings"
)

type RecentAlbumMenu struct {
	menus  []MenuItem
	albums []ds.Album
	area   string
	offset int
	limit  int
	total  int
}

func NewRecentAlbumMenu(area string) *RecentAlbumMenu {
	return &RecentAlbumMenu{
		area:   area,
		offset: 0,
		limit:  50,
		total:  -1,
	}
}

func (m *RecentAlbumMenu) MenuData() interface{} {
	return m.albums
}

func (m *RecentAlbumMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *RecentAlbumMenu) IsPlayable() bool {
	return false
}

func (m *RecentAlbumMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *RecentAlbumMenu) GetMenuKey() string {
	return fmt.Sprintf("recent_album_%s", m.area)
}

func (m *RecentAlbumMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *RecentAlbumMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.albums[index].Id)
}

func (m *RecentAlbumMenu) ExtraView() string {
	return ""
}

func (m *RecentAlbumMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *RecentAlbumMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *RecentAlbumMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true
		}

		newAlbumService := service.AlbumNewService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := newAlbumService.AlbumNew()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		// 总数量
		if total, err := jsonparser.GetInt(response, "total"); err == nil {
			m.total = int(total)
		}

		m.albums = utils.GetRecentAlbums(response)

		for _, album := range m.albums {
			var artists []string
			for _, artist := range album.Artists {
				artists = append(artists, artist.Name)
			}
			artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(album.Name), utils.ReplaceSpecialStr(artistsStr)})
		}

		return true
	}
}

func (m *RecentAlbumMenu) BottomOutHook() Hook {
	if m.total != -1 && m.offset < m.total  {
		return nil
	}
	return func(model *NeteaseModel) bool {
		m.offset = m.offset + len(m.menus)
		newAlbumService := service.AlbumNewService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := newAlbumService.AlbumNew()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		// 总数量
		if total, err := jsonparser.GetInt(response, "total"); err == nil {
			m.total = int(total)
		}

		albums := utils.GetRecentAlbums(response)

		for _, album := range albums {
			var artists []string
			for _, artist := range album.Artists {
				artists = append(artists, artist.Name)
			}
			artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
			m.menus = append(m.menus, MenuItem{utils.ReplaceSpecialStr(album.Name), utils.ReplaceSpecialStr(artistsStr)})
		}

		m.albums = append(m.albums, albums...)

		return true
	}
}

func (m *RecentAlbumMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}
