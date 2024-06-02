package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type DjHotType string

const (
	DjHot    DjHotType = "hot"
	DjNotHot DjHotType = "not_hot"
)

type DjHotMenu struct {
	baseMenu
	menus   []model.MenuItem
	radios  []structs.DjRadio
	hotType DjHotType
}

func NewDjHotMenu(base baseMenu, hotType DjHotType) *DjHotMenu {
	return &DjHotMenu{
		baseMenu: base,
		hotType:  hotType,
	}
}

func (m *DjHotMenu) IsSearchable() bool {
	return true
}

func (m *DjHotMenu) GetMenuKey() string {
	return "dj_hot"
}

func (m *DjHotMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjHotMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.baseMenu, m.radios[index].Id)
}

func (m *DjHotMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true, nil
		}

		djTopService := service.DjToplistService{
			Type: string(m.hotType),
		}
		code, response := djTopService.DjToplist()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.radios = _struct.GetDjRadiosOfTopDj(response)
		m.menus = menux.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}
