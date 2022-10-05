package ui

import (
	"strings"
)

type searchRes struct {
	item  MenuItem
	index int
}

type LocalSearchMenu struct {
	IMenu
	resItems []searchRes
}

func NewSearchMenu(originMenu IMenu, search string) *LocalSearchMenu {
	menu := &LocalSearchMenu{
		IMenu: originMenu,
	}

	for i, item := range originMenu.MenuViews() {
		if strings.Contains(item.Title, search) || strings.Contains(item.Subtitle, search) {
			menu.resItems = append(menu.resItems, searchRes{
				item:  item,
				index: i,
			})
		}
	}
	return menu
}

func (m *LocalSearchMenu) IsLocatable() bool {
	return false
}

func (m *LocalSearchMenu) MenuViews() []MenuItem {
	var items []MenuItem
	for _, item := range m.resItems {
		items = append(items, item.item)
	}
	return items
}

func (m *LocalSearchMenu) SubMenu(model *NeteaseModel, index int) IMenu {
	if index > len(m.resItems)-1 {
		return nil
	}

	return m.IMenu.SubMenu(model, m.resItems[index].index)
}

func (m *LocalSearchMenu) RealDataIndex(index int) int {
	if index > len(m.resItems)-1 {
		return 0
	}

	return m.resItems[index].index
}

func (m *LocalSearchMenu) BottomOutHook() Hook {
	return nil
}

func (m *LocalSearchMenu) TopOutHook() Hook {
	return nil
}
