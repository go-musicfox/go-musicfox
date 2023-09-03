package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DjRecommendMenu struct {
	baseMenu
	menus  []model.MenuItem
	radios []structs.DjRadio
}

func NewDjRecommendMenu(base baseMenu) *DjRecommendMenu {
	return &DjRecommendMenu{
		baseMenu: base,
	}
}

func (m *DjRecommendMenu) IsSearchable() bool {
	return true
}

func (m *DjRecommendMenu) GetMenuKey() string {
	return "dj_recommend"
}

func (m *DjRecommendMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjRecommendMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.baseMenu, m.radios[index].Id)
}

func (m *DjRecommendMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true, nil
		}

		djRecommendService := service.DjRecommendService{}
		code, response := djRecommendService.DjRecommend()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		m.radios = utils.GetDjRadios(response)
		m.menus = utils.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}
