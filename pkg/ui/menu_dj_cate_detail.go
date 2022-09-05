package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
)

type DjCategoryDetailMenu struct {
    menus      []MenuItem
    radios     []structs.DjRadio
    categoryId int64
}

func NewDjCategoryDetailMenu(categoryId int64) *DjCategoryDetailMenu {
    return &DjCategoryDetailMenu{
        categoryId: categoryId,
    }
}

func (m *DjCategoryDetailMenu) MenuData() interface{} {
    return nil
}

func (m *DjCategoryDetailMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjCategoryDetailMenu) IsPlayable() bool {
    return false
}

func (m *DjCategoryDetailMenu) ResetPlaylistWhenPlay() bool {
    return false
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

func (m *DjCategoryDetailMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjCategoryDetailMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
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

func (m *DjCategoryDetailMenu) BottomOutHook() Hook {
    return nil
}

func (m *DjCategoryDetailMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
