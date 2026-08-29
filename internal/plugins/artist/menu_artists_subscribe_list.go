package artist

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// ArtistsSubscribeListMenu 展示我收藏的歌手列表（需要登录；支持滚动加载）。
// 登录门控经 BaseMenu 转发方法访问（m.ToLoginPage），与提取前行为一致。
type ArtistsSubscribeListMenu struct {
	ui.BaseMenu
	menus   []model.MenuItem
	artists []structs.Artist
	offset  int
	limit   int
	hasMore bool
}

func NewArtistsSubscribeListMenu(base ui.BaseMenu) *ArtistsSubscribeListMenu {
	return &ArtistsSubscribeListMenu{
		BaseMenu: base,
		offset:   0,
		limit:    50,
	}
}

func (m *ArtistsSubscribeListMenu) IsSearchable() bool {
	return true
}

func (m *ArtistsSubscribeListMenu) GetMenuKey() string {
	return "artists_sub_list"
}

func (m *ArtistsSubscribeListMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistsSubscribeListMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.artists) {
		return nil
	}
	artistMenu, err := ui.BuildMenu("artist_detail", m.BaseMenu, ui.ArtistDetailOpts{ArtistID: m.artists[index].Id, Name: m.artists[index].Name})
	if err != nil {
		return nil
	}
	return artistMenu
}

func (m *ArtistsSubscribeListMenu) BeforeEnterMenuHook() model.Hook {
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
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		// 是否有更多数据
		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		m.artists = _struct.GetArtistsSublist(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(artist.Name)})
		}

		return true, nil
	}
}

func (m *ArtistsSubscribeListMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset = m.offset + len(m.menus)
		artistService := service.ArtistSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := artistService.ArtistSublist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		// 是否有更多数据
		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		m.artists = _struct.GetArtistsSublist(response)
		for _, artist := range m.artists {
			m.menus = append(m.menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(artist.Name)})
		}

		return true, nil
	}
}

func (m *ArtistsSubscribeListMenu) Artists() []structs.Artist {
	return m.artists
}
