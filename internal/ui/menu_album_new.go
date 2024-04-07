package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
)

type AlbumNewMenu struct {
	baseMenu
	menus  []model.MenuItem
	albums []structs.Album
	area   string
	offset int
	limit  int
	total  int
}

func NewAlbumNewMenu(base baseMenu, area string) *AlbumNewMenu {
	return &AlbumNewMenu{
		baseMenu: base,
		area:     area,
		offset:   0,
		limit:    50,
		total:    -1,
	}
}

func (m *AlbumNewMenu) IsSearchable() bool {
	return true
}

func (m *AlbumNewMenu) GetMenuKey() string {
	return fmt.Sprintf("album_new_%s", m.area)
}

func (m *AlbumNewMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumNewMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *AlbumNewMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true, nil
		}

		newAlbumService := service.AlbumNewService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := newAlbumService.AlbumNew()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
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
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		return true, nil
	}
}

func (m *AlbumNewMenu) BottomOutHook() model.Hook {
	if m.total != -1 && m.offset < m.total {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset = m.offset + len(m.menus)
		newAlbumService := service.AlbumNewService{
			Area:   m.area,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := newAlbumService.AlbumNew()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
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
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
		}

		m.albums = append(m.albums, albums...)

		return true, nil
	}
}

func (m *AlbumNewMenu) Albums() []structs.Album {
	return m.albums
}
