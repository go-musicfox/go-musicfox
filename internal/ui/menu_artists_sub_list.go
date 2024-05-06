package ui

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/netease-music/service"
)

type ArtistsSubListMenu struct {
	baseMenu
	menus   []model.MenuItem
	artists []structs.Artist
	offset  int
	limit   int
	total   int
}

func NewArtistsSubListMenu(base baseMenu) *ArtistsSubListMenu {
	return &ArtistsSubListMenu{
		baseMenu: base,
		offset:   0,
		limit:    50,
		total:    -1,
	}
}

func (m *ArtistsSubListMenu) IsSearchable() bool {
	return true
}

func (m *ArtistsSubListMenu) GetMenuKey() string {
	return "artists_sub_list"
}

func (m *ArtistsSubListMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistsSubListMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.artists) {
		return nil
	}
	return NewArtistDetailMenu(m.baseMenu, m.artists[index].Id, m.artists[index].Name)
}

func (m *ArtistsSubListMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 不重复请求
		if len(m.menus) > 0 && len(m.artists) > 0 {
			return true, nil
		}

		artistService := service.ArtistSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := artistService.ArtistSublist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}
		m.artists = utils.GetArtistsSublist(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(artist.Name)})
		}

		return true, nil
	}
}

func (m *ArtistsSubListMenu) BottomOutHook() model.Hook {
	if m.total != -1 && m.offset < m.total {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset = m.offset + len(m.menus)
		artistService := service.ArtistSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := artistService.ArtistSublist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		// 总数量
		if total, err := jsonparser.GetInt(response, "total"); err == nil {
			m.total = int(total)
		}

		m.artists = utils.GetArtistsSublist(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(artist.Name)})
		}

		return true, nil
	}
}

func (m *ArtistsSubListMenu) Artists() []structs.Artist {
	return m.artists
}
