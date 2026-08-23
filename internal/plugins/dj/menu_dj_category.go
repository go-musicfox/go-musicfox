package dj

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type DjCategoryMenu struct {
	ui.BaseMenu
	menus      []model.MenuItem
	categories []structs.DjCategory
}

func NewDjCategoryMenu(base ui.BaseMenu) *DjCategoryMenu {
	return &DjCategoryMenu{
		BaseMenu: base,
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

	return ui.BuildMenuOrToast("dj_category_detail", m.BaseMenu, DjCategoryDetailOpts{CategoryID: m.categories[index].Id})
}

func (m *DjCategoryMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		// 不重复请求
		if len(m.menus) > 0 && len(m.categories) > 0 {
			return true, nil
		}

		djCateService := service.DjCatelistService{}
		code, response := djCateService.DjCatelist()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.categories = _struct.GetDjCategory(response)
		m.menus = menux.GetViewFromDjCate(m.categories)

		return true, nil
	}
}
