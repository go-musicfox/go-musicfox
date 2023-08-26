package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DjCategoryMenu struct {
	baseMenu
	menus      []model.MenuItem
	categories []structs.DjCategory
}

func NewDjCategoryMenu(base baseMenu) *DjCategoryMenu {
	return &DjCategoryMenu{
		baseMenu: base,
	}
}

func (m *DjCategoryMenu) IsSearchable() bool {
	return true
}

func (m *DjCategoryMenu) GetMenuKey() string {
	return "dj_category"
}

func (m *DjCategoryMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjCategoryMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.categories) {
		return nil
	}

	return NewDjCategoryDetailMenu(m.baseMenu, m.categories[index].Id)
}

func (m *DjCategoryMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		// 不重复请求
		if len(m.menus) > 0 && len(m.categories) > 0 {
			return true, nil
		}

		djCateService := service.DjCatelistService{}
		code, response := djCateService.DjCatelist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		m.categories = utils.GetDjCategory(response)
		m.menus = utils.GetViewFromDjCate(m.categories)

		return true, nil
	}
}
