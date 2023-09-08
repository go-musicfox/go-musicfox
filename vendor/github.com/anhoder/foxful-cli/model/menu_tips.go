package model

import (
	"fmt"

	"github.com/muesli/termenv"
)

type MenuTips struct {
	main       *Main
	originMenu *MenuItem
}

func NewMenuTips(m *Main, originMenu *MenuItem) *MenuTips {
	return &MenuTips{
		main:       m,
		originMenu: originMenu,
	}
}

func (t *MenuTips) DisplayTips(tips string) {
	var (
		subTitle  string
		menuTitle *MenuItem
	)
	termenv.DefaultOutput().MoveCursor(t.main.MenuTitleStartRow(), 0)

	if t.originMenu != nil {
		menuTitle = t.originMenu
	} else {
		menuTitle = t.main.MenuTitle()
	}

	if menuTitle.Subtitle != "" {
		subTitle = menuTitle.Subtitle + " " + tips
	} else {
		subTitle = tips
	}
	fmt.Print(t.main.MenuTitleView(t.main.app, nil, &MenuItem{
		Title:    menuTitle.Title,
		Subtitle: subTitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}

func (t *MenuTips) Recover() {
	termenv.DefaultOutput().MoveCursor(t.main.menuTitleStartRow, 0)

	fmt.Print(t.main.MenuTitleView(t.main.app, nil, &MenuItem{
		Title:    t.main.menuTitle.Title,
		Subtitle: t.main.menuTitle.Subtitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}
