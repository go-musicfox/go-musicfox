package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
)

type DjCategoryDetailMenu struct {
	DefaultMenu
	menus      []MenuItem
	radios     []structs.DjRadio
	categoryId int64
}

func NewDjCategoryDetailMenu(categoryId int64) *DjCategoryDetailMenu {
	return &DjCategoryDetailMenu{
		categoryId: categoryId,
	}
}

func (m *DjCategoryDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("dj_category_detail_%d", m.categoryId)
}

func (m *DjCategoryDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjCategoryDetailMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjCategoryDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true
		}

		cateDetailService := service.DjRecommendTypeService{
			CateId: strconv.FormatInt(m.categoryId, 10),
		}
		code, response := cateDetailService.DjRecommendType()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		m.radios = utils.GetDjRadios(response)
		m.menus = GetViewFromDjRadios(m.radios)

		return true
	}
}
