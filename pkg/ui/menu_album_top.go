package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type AlbumTopMenu struct {
	baseMenu
	menus   []model.MenuItem
	albums  []structs.Album
	area    string
	offset  int
	limit   int
	hasMore bool
}

func NewAlbumTopMenu(base baseMenu, area string) *AlbumTopMenu {
	return &AlbumTopMenu{
		baseMenu: base,
		area:     area,
		offset:   0,
		limit:    50,
	}
}

func (m *AlbumTopMenu) IsSearchable() bool {
	return true
}

func (m *AlbumTopMenu) GetMenuKey() string {
	return fmt.Sprintf("album_top_%s", m.area)
}

func (m *AlbumTopMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumTopMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *AlbumTopMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true, nil
		}

		topAlbumService := service.TopAlbumService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := topAlbumService.TopAlbum()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
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
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		return true, nil
	}
}

func (m *AlbumTopMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset = m.offset + len(m.menus)
		topAlbumService := service.TopAlbumService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := topAlbumService.TopAlbum()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
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
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		m.albums = append(m.albums, albums...)

		return true, nil
	}
}

func (m *AlbumTopMenu) Albums() []structs.Album {
	return m.albums
}
