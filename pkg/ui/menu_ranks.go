package ui

import (
	"fmt"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type RanksMenu struct {
	DefaultMenu
	menus []MenuItem
	ranks []structs.Rank
}

func NewRanksMenu() *RanksMenu {
	return new(RanksMenu)
}

func (m *RanksMenu) IsSearchable() bool {
	return true
}

func (m *RanksMenu) GetMenuKey() string {
	return "ranks"
}

func (m *RanksMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *RanksMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.ranks) {
		return nil
	}

	return NewPlaylistDetailMenu(m.ranks[index].Id)
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
			menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(rank.Name), Subtitle: frequency})
		}
		m.menus = menus

		return true
	}
}
