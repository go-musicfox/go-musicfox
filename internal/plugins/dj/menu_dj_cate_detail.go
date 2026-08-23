package dj

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type DjCategoryDetailMenu struct {
	ui.BaseMenu
	menus      []model.MenuItem
	radios     []structs.DjRadio
	categoryID int64
}

func NewDjCategoryDetailMenu(base ui.BaseMenu, categoryID int64) *DjCategoryDetailMenu {
	return &DjCategoryDetailMenu{
		BaseMenu:   base,
		categoryID: categoryID,
	}
}

func (m *DjCategoryDetailMenu) IsSearchable() bool {
	return true
}

func (m *DjCategoryDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("dj_category_detail_%d", m.categoryID)
}

func (m *DjCategoryDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjCategoryDetailMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return ui.BuildMenuOrToast("dj_radio_detail", m.BaseMenu, ui.DjRadioDetailOpts{DjRadioID: m.radios[index].Id})
}

func (m *DjCategoryDetailMenu) ItemToShare(index int) any {
	if index >= 0 && index < len(m.radios) {
		return m.radios[index]
	}
	return nil
}

func (m *DjCategoryDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true, nil
		}

		cateDetailService := service.DjRecommendTypeService{
			CateId: strconv.FormatInt(m.categoryID, 10),
		}
		code, response := cateDetailService.DjRecommendType()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.radios = _struct.GetDjRadios(response)
		m.menus = menux.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}
