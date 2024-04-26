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

type AlbumSubListMenu struct {
	baseMenu
	menus  []model.MenuItem
	albums []structs.Album
	offset int
	limit  int
	total  int
}

func NewAlbumSubListMenu(base baseMenu) *AlbumSubListMenu {
	return &AlbumSubListMenu{
		baseMenu: base,
		offset:   0,
		limit:    50,
		total:    -1,
	}
}

func (m *AlbumSubListMenu) IsSearchable() bool {
	return true
}

func (m *AlbumSubListMenu) GetMenuKey() string {
	return "album_sub_list"
}

func (m *AlbumSubListMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumSubListMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *AlbumSubListMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if len(m.menus) > 0 && len(m.albums) > 0 {
			return true, nil
		}

		albumService := service.AlbumSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := albumService.AlbumSublist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		m.albums = utils.GetAlbumsSublist(response)

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

func (m *AlbumSubListMenu) BottomOutHook() model.Hook {
	if m.total != -1 && m.offset < m.total {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset = m.offset + len(m.menus)
		newAlbumService := service.AlbumSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := newAlbumService.AlbumSublist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		// 总数量
		if total, err := jsonparser.GetInt(response, "total"); err == nil {
			m.total = int(total)
		}

		albums := utils.GetAlbumsSublist(response)

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

func (m *AlbumSubListMenu) Albums() []structs.Album {
	return m.albums
}
