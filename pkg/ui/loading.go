package ui

import (
	"fmt"
	"github.com/muesli/termenv"
	"go-musicfox/pkg/configs"
	"strings"
	"unicode/utf8"
)

type Loading struct {
	model  *NeteaseModel
	curLen int
}

func NewLoading(m *NeteaseModel) *Loading {
	return &Loading{
		model: m,
	}
}

// 开始
func (loading *Loading) start() {
	termenv.MoveCursor(loading.model.menuTitleStartRow, 0)

	loading.curLen = utf8.RuneCountInString(loading.model.menuTitle.OriginString()) + utf8.RuneCountInString(" "+configs.ConfigRegistry.MainLoadingText)

	var repeatSpace string
	if loading.model.menuTitleStartColumn > 0 {
		repeatSpace = strings.Repeat(" ", loading.model.menuTitleStartColumn)
	}
	fmt.Printf("%s%s%s",
		repeatSpace,
		SetFgStyle(loading.model.menuTitle.String(), termenv.ANSIBrightGreen),
		SetFgStyle(" "+configs.ConfigRegistry.MainLoadingText, termenv.ANSIBrightBlack))

	termenv.MoveCursor(0, 0)
}

// 完成
func (loading *Loading) complete() {
	termenv.MoveCursor(loading.model.menuTitleStartRow, 0)

	spaceLen := loading.curLen - utf8.RuneCountInString(loading.model.menuTitle.OriginString())
	if spaceLen < 0 {
		spaceLen = 0
	}

	var repeatSpace string
	if loading.model.menuTitleStartColumn > 0 {
		repeatSpace = strings.Repeat(" ", loading.model.menuTitleStartColumn)
	}

	fmt.Printf("%s%s%s",
		repeatSpace,
		SetFgStyle(loading.model.menuTitle.String(), termenv.ANSIBrightGreen),
		strings.Repeat("　", spaceLen))

	termenv.MoveCursor(0, 0)
}
