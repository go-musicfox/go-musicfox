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

func (m *DjSubListMenu) ItemToShare(index int) any {
		if index >= 0 && index < len(m.radios) {
			return m.radios[index]
		}
		return  nil
}

func (m *DjSubListMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if _struct.CheckUserInfo(m.netease.user) == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
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
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		if total, err := jsonparser.GetInt(response, "count"); err != nil {
			m.total = int(total)
		}

		m.radios = _struct.GetDjRadios(response)
		m.menus = menux.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}

func (m *DjSubListMenu) BottomOutHook() model.Hook {
	if len(m.radios) >= m.total {
		return nil
	}

	return func(main *model.Main) (bool, model.Page) {
		m.offset += m.limit

		if _struct.CheckUserInfo(m.netease.user) == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		djSublistService := service.DjSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := djSublistService.DjSublist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}

		if total, err := jsonparser.GetInt(response, "count"); err != nil {
			m.total = int(total)
		}

		radios := _struct.GetDjRadios(response)
		menus := menux.GetViewFromDjRadios(radios)

		m.radios = append(m.radios, radios...)
		m.menus = append(m.menus, menus...)

		return true, nil
	}
}
