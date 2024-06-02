package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type RanksMenu struct {
	baseMenu
	menus []model.MenuItem
	ranks []structs.Rank
}

func NewRanksMenu(base baseMenu) *RanksMenu {
	return &RanksMenu{
		baseMenu: base,
	}
}

func (m *RanksMenu) IsSearchable() bool {
	return true
}

func (m *RanksMenu) GetMenuKey() string {
	return "ranks"
}

func (m *RanksMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *RanksMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.ranks) {
		return nil
	}

	return NewPlaylistDetailMenu(m.baseMenu, m.ranks[index].Id)
}

func (m *RanksMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if len(m.menus) > 0 && len(m.ranks) > 0 {
			return true, nil
		}

		rankListService := service.ToplistService{}
		code, response := rankListService.Toplist()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		var menus []model.MenuItem
		m.ranks = _struct.GetRanks(response)
		for _, rank := range m.ranks {
			frequency := fmt.Sprintf("[%s]", rank.Frequency)
			menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(rank.Name), Subtitle: frequency})
		}
		m.menus = menus

		return true, nil
	}
}
