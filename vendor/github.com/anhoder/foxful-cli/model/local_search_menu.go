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

type LocalSearchMenu struct {
	Menu
	resItems fuzzy.Matches
}

func NewSearchMenu(originMenu Menu, search string) *LocalSearchMenu {
	menu := &LocalSearchMenu{
		Menu:     originMenu,
		resItems: fuzzy.FindFrom(search, searchableMenus(originMenu.MenuViews())),
	}

	return menu
}

func (m *LocalSearchMenu) IsLocatable() bool {
	return false
}

func (m *LocalSearchMenu) MenuViews() []MenuItem {
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

func (m *LocalSearchMenu) SubMenu(a *App, index int) Menu {
	if index > len(m.resItems)-1 {
		return nil
	}

	return m.Menu.SubMenu(a, m.resItems[index].Index)
}

func (m *LocalSearchMenu) RealDataIndex(index int) int {
	if index > len(m.resItems)-1 {
		return 0
	}

	return m.resItems[index].Index
}

func (m *LocalSearchMenu) BottomOutHook() Hook {
	return nil
}

func (m *LocalSearchMenu) TopOutHook() Hook {
	return nil
}
