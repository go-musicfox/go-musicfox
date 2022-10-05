package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
	"strings"
)

type AlbumNewMenu struct {
	DefaultMenu
	menus  []MenuItem
	albums []structs.Album
	area   string
	offset int
	limit  int
	total  int
}

func NewAlbumNewMenu(area string) *AlbumNewMenu {
	return &AlbumNewMenu{
		area:   area,
		offset: 0,
		limit:  50,
		total:  -1,
	}
}

func (m *AlbumNewMenu) IsSearchable() bool {
	return true
}

func (m *AlbumNewMenu) MenuData() interface{} {
	return m.albums
}

func (m *AlbumNewMenu) GetMenuKey() string {
	return fmt.Sprintf("album_new_%s", m.area)
}

func (m *AlbumNewMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumNewMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.albums[index].Id)
}

func (m *AlbumNewMenu) BeforeEnterMenuHook() Hook {
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

		m.albums = utils.GetNewAlbums(response)

		for _, album := range m.albums {
			var artists []string
			for _, artist := range album.Artists {
				artists = append(artists, artist.Name)
			}
			artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		return true
	}
}

func (m *AlbumNewMenu) BottomOutHook() Hook {
	if m.total != -1 && m.offset < m.total {
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

		albums := utils.GetNewAlbums(response)

		for _, album := range albums {
			var artists []string
			for _, artist := range album.Artists {
				artists = append(artists, artist.Name)
			}
			artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
			m.menus = append(m.menus, MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		m.albums = append(m.albums, albums...)

		return true
	}
}
