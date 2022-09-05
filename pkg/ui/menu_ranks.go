package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type RanksMenu struct {
    menus []MenuItem
    ranks []structs.Rank
}

func NewRanksMenu() *RanksMenu {
    return new(RanksMenu)
}

func (m *RanksMenu) MenuData() interface{} {
    return nil
}

func (m *RanksMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *RanksMenu) IsPlayable() bool {
    return false
}

func (m *RanksMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *RanksMenu) GetMenuKey() string {
    return "ranks"
}

func (m *RanksMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *RanksMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index >= len(m.ranks) {
        return nil
    }

    return NewPlaylistDetailMenu(m.ranks[index].Id)
}

func (m *RanksMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *RanksMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *RanksMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {
        if len(m.menus) > 0 && len(m.ranks) > 0 {
            return true
        }

        rankListService := service.ToplistService{}
        code, response := rankListService.Toplist()
        codeType := utils.CheckCode(code)
        if codeType != utils.Success {
            return false
        }

        var menus []MenuItem
        m.ranks = utils.GetRanks(response)
        for _, rank := range m.ranks {
            frequency := fmt.Sprintf("[%s]", rank.Frequency)
            menus = append(menus, MenuItem{utils.ReplaceSpecialStr(rank.Name), frequency})
        }
        m.menus = menus

        return true
    }
}

func (m *RanksMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *RanksMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
