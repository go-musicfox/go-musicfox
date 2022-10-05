package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type DjHotType string

const (
	DjHot    DjHotType = "hot"
	DjNotHot           = "not_hot"
)

type DjHotMenu struct {
	DefaultMenu
	menus   []MenuItem
	radios  []structs.DjRadio
	hotType DjHotType
}

func NewDjHotMenu(hotType DjHotType) *DjHotMenu {
	return &DjHotMenu{
		hotType: hotType,
	}
}

func (m *DjHotMenu) IsSearchable() bool {
	return true
}

func (m *DjHotMenu) GetMenuKey() string {
	return "dj_hot"
}

func (m *DjHotMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjHotMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjHotMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true
		}

		djTopService := service.DjToplistService{
			Type: string(m.hotType),
		}
		code, response := djTopService.DjToplist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		m.radios = utils.GetDjRadiosOfTopDj(response)
		m.menus = GetViewFromDjRadios(m.radios)

		return true
	}
}
