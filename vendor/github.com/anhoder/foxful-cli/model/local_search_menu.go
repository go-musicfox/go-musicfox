package model

import (
	tea "charm.land/bubbletea/v2"
	"github.com/sahilm/fuzzy"
)

type searchableMenus []MenuItem

func (m searchableMenus) String(i int) string {
	return m[i].OriginString()
}

func (m searchableMenus) Len() int {
	return len(m)
}

type LocalSearchMenuImpl struct {
	Menu
	resItems fuzzy.Matches
}

func DefaultSearchMenu() *LocalSearchMenuImpl {
	return &LocalSearchMenuImpl{}
}

func (m *LocalSearchMenuImpl) Search(originMenu Menu, search string) {
	m.Menu = originMenu
	m.resItems = fuzzy.FindFrom(search, searchableMenus(originMenu.MenuViews()))
}

func (m *LocalSearchMenuImpl) MenuViews() []MenuItem {
	var (
		items []MenuItem
		menus = m.Menu.MenuViews()
		seen  = make(map[int]bool, len(m.resItems))
	)
	for _, v := range m.resItems {
		if seen[v.Index] {
			continue
		}
		seen[v.Index] = true
		items = append(items, menus[v.Index])
	}
	return items
}

func (m *LocalSearchMenuImpl) SubMenu(a *App, index int) Menu {
	if index > len(m.resItems)-1 {
		return nil
	}

	return m.Menu.SubMenu(a, m.resItems[index].Index)
}

func (m *LocalSearchMenuImpl) RealDataIndex(index int) int {
	if index > len(m.resItems)-1 {
		return 0
	}

	return m.resItems[index].Index
}

// Action remaps the filtered-result index to the origin menu's real data index
// before delegating, so callbacks receive the correct item. Without this
// override the embedded Menu.Action would be invoked with the filtered index.
func (m *LocalSearchMenuImpl) Action(a *App, index int) (Page, tea.Cmd) {
	if index > len(m.resItems)-1 {
		return nil, nil
	}
	return m.Menu.Action(a, m.resItems[index].Index)
}

// ContextMenuItems remaps the filtered-result index to the origin menu's real
// data index before delegating.
func (m *LocalSearchMenuImpl) ContextMenuItems(a *App, index int) []ContextMenuItem {
	if index > len(m.resItems)-1 {
		return nil
	}
	return m.Menu.ContextMenuItems(a, m.resItems[index].Index)
}

// ContextMenuAction remaps the filtered-result index to the origin menu's real
// data index before delegating.
func (m *LocalSearchMenuImpl) ContextMenuAction(a *App, index int, item ContextMenuItem) (Page, tea.Cmd) {
	if index > len(m.resItems)-1 {
		return nil, nil
	}
	return m.Menu.ContextMenuAction(a, m.resItems[index].Index, item)
}

func (m *LocalSearchMenuImpl) BottomOutHook() Hook {
	return nil
}

func (m *LocalSearchMenuImpl) TopOutHook() Hook {
	return nil
}
