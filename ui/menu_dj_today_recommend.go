package ui

import (
    "github.com/anhoder/netease-music/service"
    "go-musicfox/ds"
    "go-musicfox/utils"
)

type DjTodayRecommendMenu struct {
    menus  []MenuItem
    radios []ds.DjRadio
}

func NewDjTodayRecommendMenu() *DjTodayRecommendMenu {
    return &DjTodayRecommendMenu{}
}

func (m *DjTodayRecommendMenu) MenuData() interface{} {
    return nil
}

func (m *DjTodayRecommendMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjTodayRecommendMenu) IsPlayable() bool {
    return false
}

func (m *DjTodayRecommendMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *DjTodayRecommendMenu) GetMenuKey() string {
    return "dj_today_recommend"
}

func (m *DjTodayRecommendMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *DjTodayRecommendMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index >= len(m.radios) {
        return nil
    }

    return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjTodayRecommendMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjTodayRecommendMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjTodayRecommendMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {

        // 不重复请求
        if len(m.menus) > 0 && len(m.radios) > 0 {
            return true
        }

        djTodayService := service.DjTodayPerferedService{}
        code, response := djTodayService.DjTodayPerfered()
        codeType := utils.CheckCode(code)
        if codeType != utils.Success {
            return false
        }

        m.radios = utils.GetDjRadiosOfToday(response)
        m.menus = GetViewFromDjRadios(m.radios)

        return true
    }
}

func (m *DjTodayRecommendMenu) BottomOutHook() Hook {
    return func(model *NeteaseModel) bool {

        djTodayService := service.DjTodayPerferedService{}
        code, response := djTodayService.DjTodayPerfered()
        codeType := utils.CheckCode(code)
        if codeType != utils.Success {
            return false
        }

        radios := utils.GetDjRadiosOfToday(response)
        menus := GetViewFromDjRadios(radios)

        m.radios = append(m.radios, radios...)
        m.menus = append(m.menus, menus...)

        return true
    }
}

func (m *DjTodayRecommendMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
