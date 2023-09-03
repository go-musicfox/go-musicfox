package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type DjCategoryDetailMenu struct {
	baseMenu
	menus      []model.MenuItem
	radios     []structs.DjRadio
	categoryId int64
}

func NewDjCategoryDetailMenu(base baseMenu, categoryId int64) *DjCategoryDetailMenu {
	return &DjCategoryDetailMenu{
		baseMenu:   base,
		categoryId: categoryId,
	}
}

func (m *DjCategoryDetailMenu) IsSearchable() bool {
	return true
}

func (m *DjCategoryDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("dj_category_detail_%d", m.categoryId)
}

func (m *DjCategoryDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjCategoryDetailMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.baseMenu, m.radios[index].Id)
}

func (m *DjCategoryDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true, nil
		}

		cateDetailService := service.DjRecommendTypeService{
			CateId: strconv.FormatInt(m.categoryId, 10),
		}
		code, response := cateDetailService.DjRecommendType()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		m.radios = utils.GetDjRadios(response)
		m.menus = utils.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}
