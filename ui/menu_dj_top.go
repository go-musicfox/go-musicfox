package ui

import (
    "github.com/anhoder/netease-music/service"
    "go-musicfox/ds"
    "go-musicfox/utils"
)

type DjHotType string

const (
    DjHot    DjHotType = "hot"
    DjNotHot           = "not_hot"
)

type DjHotMenu struct {
    menus   []MenuItem
    radios  []ds.DjRadio
    hotType DjHotType
}

func NewDjHotMenu(hotType DjHotType) *DjHotMenu {
    return &DjHotMenu{
        hotType: hotType,
    }
}

func (m *DjHotMenu) MenuData() interface{} {
    return nil
}

func (m *DjHotMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjHotMenu) IsPlayable() bool {
    return false
}

func (m *DjHotMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *DjHotMenu) GetMenuKey() string {
    return "dj_hot"
}

func (m *DjHotMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *DjHotMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index >= len(m.radios) {
        return nil
    }

    return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjHotMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjHotMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjHotMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {

        // 不重复请求
        if len(m.menus) > 0 && len(m.radios) > 0 {
            return true
        }

        djTopService := service.DjToplistService{
            Type:   string(m.hotType),
        }
        code, response := djTopService.DjToplist()
        codeType := utils.CheckCode(code)
        if codeType != utils.Success {
            return false
        }

        m.radios = utils.GetDjRadiosOfTopDj(response)
        m.menus = GetViewFromDjRadios(m.radios)

        return true
    }
}

func (m *DjHotMenu) BottomOutHook() Hook {
    return nil
}

func (m *DjHotMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
