package ui

import (
	"fmt"

	"github.com/go-musicfox/go-musicfox/pkg/configs"

	"github.com/muesli/termenv"
)

type Loading struct {
	model *NeteaseModel
}

func NewLoading(m *NeteaseModel) *Loading {
	return &Loading{
		model: m,
	}
}

// 开始
func (loading *Loading) start() {
	termenv.DefaultOutput().MoveCursor(loading.model.menuTitleStartRow, 0)

	var subTitle string
	if loading.model.menuTitle.Subtitle != "" {
		subTitle = loading.model.menuTitle.Subtitle + " " + configs.ConfigRegistry.MainLoadingText
	} else {
		subTitle = configs.ConfigRegistry.MainLoadingText
	}
	fmt.Print(loading.model.menuTitleView(loading.model, nil, &MenuItem{
		Title:    loading.model.menuTitle.Title,
		Subtitle: subTitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}

// 完成
func (loading *Loading) complete() {
	termenv.DefaultOutput().MoveCursor(loading.model.menuTitleStartRow, 0)

	fmt.Print(loading.model.menuTitleView(loading.model, nil, &MenuItem{
		Title:    loading.model.menuTitle.Title,
		Subtitle: loading.model.menuTitle.Subtitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}
