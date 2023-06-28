package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DjCategoryMenu struct {
	DefaultMenu
	menus      []MenuItem
	categories []structs.DjCategory
}

func NewDjCategoryMenu() *DjCategoryMenu {
	return &DjCategoryMenu{}
}

func (m *DjCategoryMenu) IsSearchable() bool {
	return true
}

func (m *DjCategoryMenu) GetMenuKey() string {
	return "dj_category"
}

func (m *DjCategoryMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjCategoryMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.categories) {
		return nil
	}

	return NewDjCategoryDetailMenu(m.categories[index].Id)
}

func (m *DjCategoryMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		// 不重复请求
		if len(m.menus) > 0 && len(m.categories) > 0 {
			return true
		}

		djCateService := service.DjCatelistService{}
		code, response := djCateService.DjCatelist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		m.categories = utils.GetDjCategory(response)
		m.menus = GetViewFromDjCate(m.categories)

		return true
	}
}
