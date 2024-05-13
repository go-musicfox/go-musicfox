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

type AlbumSubscribeListMenu struct {
	baseMenu
	menus   []model.MenuItem
	albums  []structs.Album
	offset  int
	limit   int
	hasMore bool
}

func NewAlbumSubscribeListMenu(base baseMenu) *AlbumSubscribeListMenu {
	return &AlbumSubscribeListMenu{
		baseMenu: base,
		offset:   0,
		limit:    50,
	}
}

func (m *AlbumSubscribeListMenu) IsSearchable() bool {
	return true
}

func (m *AlbumSubscribeListMenu) GetMenuKey() string {
	return "album_sub_list"
}

func (m *AlbumSubscribeListMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumSubscribeListMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *AlbumSubscribeListMenu) BeforeEnterMenuHook() model.Hook {
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
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		// 是否有更多数据
		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
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

func (m *AlbumSubscribeListMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
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
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		// 是否有更多数据
		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
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

func (m *AlbumSubscribeListMenu) Albums() []structs.Album {
	return m.albums
}
