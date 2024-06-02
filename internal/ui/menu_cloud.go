package ui

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type CloudMenu struct {
	baseMenu
	menus   []model.MenuItem
	songs   []structs.Song
	limit   int
	offset  int
	hasMore bool
}

func NewCloudMenu(base baseMenu) *CloudMenu {
	return &CloudMenu{
		baseMenu: base,
		limit:    100,
		offset:   0,
	}
}

func (m *CloudMenu) IsSearchable() bool {
	return true
}

func (m *CloudMenu) IsPlayable() bool {
	return true
}

func (m *CloudMenu) GetMenuKey() string {
	return "could"
}

func (m *CloudMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *CloudMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if _struct.CheckUserInfo(m.netease.user) == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		// 不重复请求
		if len(m.menus) > 0 && len(m.songs) > 0 {
			return true, nil
		}

		cloudService := service.UserCloudService{
			Offset: strconv.Itoa(m.offset),
			Limit:  strconv.Itoa(m.limit),
		}
		code, response := cloudService.UserCloud()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		m.songs = _struct.GetSongsOfCloud(response)
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *CloudMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset += m.limit

		cloudService := service.UserCloudService{
			Offset: strconv.Itoa(m.offset),
			Limit:  strconv.Itoa(m.limit),
		}
		code, response := cloudService.UserCloud()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(func() model.Page {
				main.RefreshMenuList()
				return nil
			})
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		songs := _struct.GetSongsOfCloud(response)
		menus := menux.GetViewFromSongs(songs)

		m.songs = append(m.songs, songs...)
		m.menus = append(m.menus, menus...)

		return true, nil
	}
}

func (m *CloudMenu) Songs() []structs.Song {
	return m.songs
}
