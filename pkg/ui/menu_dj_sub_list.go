package ui

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type DjSubListMenu struct {
	baseMenu
	menus  []model.MenuItem
	radios []structs.DjRadio
	limit  int
	offset int
	total  int
}

func NewDjSubListMenu(base baseMenu) *DjSubListMenu {
	return &DjSubListMenu{
		baseMenu: base,
		limit:    50,
		offset:   0,
		total:    -1,
	}
}

func (m *DjSubListMenu) IsSearchable() bool {
	return true
}

func (m *DjSubListMenu) GetMenuKey() string {
	return "dj_sub"
}

func (m *DjSubListMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjSubListMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.baseMenu, m.radios[index].Id)
}

func (m *DjSubListMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if utils.CheckUserInfo(m.netease.user) == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuLoginCallback(main))
			return false, page
		}

		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true, nil
		}

		djSublistService := service.DjSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := djSublistService.DjSublist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuLoginCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		if total, err := jsonparser.GetInt(response, "count"); err != nil {
			m.total = int(total)
		}

		m.radios = utils.GetDjRadios(response)
		m.menus = utils.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}

func (m *DjSubListMenu) BottomOutHook() model.Hook {
	if len(m.radios) >= m.total {
		return nil
	}

	return func(main *model.Main) (bool, model.Page) {
		m.offset += m.limit

		if utils.CheckUserInfo(m.netease.user) == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuLoginCallback(main))
			return false, page
		}

		djSublistService := service.DjSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := djSublistService.DjSublist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuLoginCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		if total, err := jsonparser.GetInt(response, "count"); err != nil {
			m.total = int(total)
		}

		radios := utils.GetDjRadios(response)
		menus := utils.GetViewFromDjRadios(radios)

		m.radios = append(m.radios, radios...)
		m.menus = append(m.menus, menus...)

		return true, nil
	}
}
