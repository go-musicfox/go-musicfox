package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DjHotType string

const (
	DjHot    DjHotType = "hot"
	DjNotHot DjHotType = "not_hot"
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

func (m *DjHotMenu) SubMenu(_ *NeteaseModel, index int) Menu {
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
