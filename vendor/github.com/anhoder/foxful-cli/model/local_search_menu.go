package model

import (
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
	)
	for _, v := range m.resItems {
		// matchedMap := lo.Associate(v.MatchedIndexes, func(i int) (int, bool) { return i, true })
		// titleRune := []rune(menu.Title)
		// for i := 0; i < len(titleRune); i++ {
		// 	if matchedMap[i] {
		// 		fmt.Print(fmt.Sprintf(bold, string(match.Str[i])))
		// 	} else {
		// 		fmt.Print(string(match.Str[i]))
		// 	}
		// }
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

func (m *LocalSearchMenuImpl) BottomOutHook() Hook {
	return nil
}

func (m *LocalSearchMenuImpl) TopOutHook() Hook {
	return nil
}
