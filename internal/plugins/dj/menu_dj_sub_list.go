package dj

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// DjSubListMenu 展示我订阅的电台列表（需要登录；支持滚动加载）。登录门控经
// BaseMenu 转发方法访问（m.User()/m.ToLoginPage），与提取前行为一致。
type DjSubListMenu struct {
	ui.BaseMenu
	menus  []model.MenuItem
	radios []structs.DjRadio
	limit  int
	offset int
	total  int
}

func NewDjSubListMenu(base ui.BaseMenu) *DjSubListMenu {
	return &DjSubListMenu{
		BaseMenu: base,
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

	return ui.BuildMenuOrToast("dj_radio_detail", m.BaseMenu, ui.DjRadioDetailOpts{DjRadioID: m.radios[index].Id})
}

func (m *DjSubListMenu) ItemToShare(index int) any {
	if index >= 0 && index < len(m.radios) {
		return m.radios[index]
	}
	return nil
}

func (m *DjSubListMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		if _struct.CheckUserInfo(m.User()) == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
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
			page, _ := m.ToLoginPage(enterMenuCallback(main))
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

		if _struct.CheckUserInfo(m.User()) == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		}

		djSublistService := service.DjSublistService{
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := djSublistService.DjSublist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
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

// enterMenuCallback mirrors ui.EnterMenuCallback (unexported there): the login
// callback re-enters the requesting menu once login succeeds.
func enterMenuCallback(main *model.Main) ui.LoginCallback {
	return func() model.Page {
		return main.EnterMenu(nil, nil)
	}
}
