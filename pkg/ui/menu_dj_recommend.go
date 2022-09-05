package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type DjRecommendMenu struct {
    menus  []MenuItem
    radios []structs.DjRadio
}

func NewDjRecommendMenu() *DjRecommendMenu {
    return &DjRecommendMenu{}
}

func (m *DjRecommendMenu) MenuData() interface{} {
    return nil
}

func (m *DjRecommendMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjRecommendMenu) IsPlayable() bool {
    return false
}

func (m *DjRecommendMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *DjRecommendMenu) GetMenuKey() string {
    return "dj_recommend"
}

func (m *DjRecommendMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *DjRecommendMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index >= len(m.radios) {
        return nil
    }

    return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjRecommendMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjRecommendMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
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

func (m *DjRecommendMenu) BottomOutHook() Hook {
    return nil
}

func (m *DjRecommendMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
