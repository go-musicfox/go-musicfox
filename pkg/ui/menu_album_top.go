package ui

import (
	"fmt"
	"strconv"
	"strings"

	"go-musicfox/pkg/structs"
	"go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
)

type AlbumTopMenu struct {
	DefaultMenu
	menus   []MenuItem
	albums  []structs.Album
	area    string
	offset  int
	limit   int
	hasMore bool
}

func NewAlbumTopMenu(area string) *AlbumTopMenu {
	return &AlbumTopMenu{
		area:   area,
		offset: 0,
		limit:  50,
	}
}

func (m *AlbumTopMenu) IsSearchable() bool {
	return true
}

func (m *AlbumTopMenu) MenuData() interface{} {
	return m.albums
}

func (m *AlbumTopMenu) GetMenuKey() string {
	return fmt.Sprintf("album_top_%s", m.area)
}

func (m *AlbumTopMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *AlbumTopMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.albums[index].Id)
}

func (m *AlbumTopMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true
		}

		topAlbumService := service.TopAlbumService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := topAlbumService.TopAlbum()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		// 是否有更多数据
		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		m.albums = utils.GetTopAlbums(response)

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

func (m *AlbumTopMenu) BottomOutHook() Hook {
	if !m.hasMore {
		return nil
	}
	return func(model *NeteaseModel) bool {
		m.offset = m.offset + len(m.menus)
		topAlbumService := service.TopAlbumService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := topAlbumService.TopAlbum()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		// 是否有更多数据
		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		albums := utils.GetTopAlbums(response)

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
