package ui

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
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
		if utils.CheckUserInfo(m.netease.user) == utils.NeedLogin {
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
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		m.songs = utils.GetSongsOfCloud(response)
		m.menus = utils.GetViewFromSongs(m.songs)

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
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(func() model.Page {
				main.RefreshMenuList()
				return nil
			})
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		songs := utils.GetSongsOfCloud(response)
		menus := utils.GetViewFromSongs(songs)

		m.songs = append(m.songs, songs...)
		m.menus = append(m.menus, menus...)

		return true, nil
	}
}

func (m *CloudMenu) Songs() []structs.Song {
	return m.songs
}
