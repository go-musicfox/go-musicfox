package model

import (
	"fmt"

	"github.com/muesli/termenv"
)

type Loading struct {
	main *Main
}

func NewLoading(m *Main) *Loading {
	return &Loading{
		main: m,
	}
}

func (loading *Loading) start() {
	termenv.DefaultOutput().MoveCursor(loading.main.menuTitleStartRow, 0)

	var subTitle string
	if loading.main.menuTitle.Subtitle != "" {
		subTitle = loading.main.menuTitle.Subtitle + " " + loading.main.options.LoadingText
	} else {
		subTitle = loading.main.options.LoadingText
	}
	fmt.Print(loading.main.MenuTitleView(loading.main.app, nil, &MenuItem{
		Title:    loading.main.menuTitle.Title,
		Subtitle: subTitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}

func (loading *Loading) complete() {
	termenv.DefaultOutput().MoveCursor(loading.main.menuTitleStartRow, 0)

	fmt.Print(loading.main.MenuTitleView(loading.main.app, nil, &MenuItem{
		Title:    loading.main.menuTitle.Title,
		Subtitle: loading.main.menuTitle.Subtitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}
