package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type DjRecommendMenu struct {
	DefaultMenu
	menus  []MenuItem
	radios []structs.DjRadio
}

func NewDjRecommendMenu() *DjRecommendMenu {
	return &DjRecommendMenu{}
}

func (m *DjRecommendMenu) IsSearchable() bool {
	return true
}

func (m *DjRecommendMenu) GetMenuKey() string {
	return "dj_recommend"
}

func (m *DjRecommendMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjRecommendMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjRecommendMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true
		}

		djRecommendService := service.DjRecommendService{}
		code, response := djRecommendService.DjRecommend()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		m.radios = utils.GetDjRadios(response)
		m.menus = GetViewFromDjRadios(m.radios)

		return true
	}
}
