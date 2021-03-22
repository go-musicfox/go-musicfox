package ui

import (
    "fmt"
    "github.com/anhoder/go-musicfox/constants"
    "github.com/muesli/termenv"
    "strings"
    "unicode/utf8"
)

type Loading struct {
    model  *neteaseModel
    curLen int
}

func NewLoading(m *neteaseModel) *Loading {
    return &Loading{
        model: m,
    }
}

// 开始
func (loading *Loading)start() {
    termenv.MoveCursor(loading.model.menuTitleStartRow, 0)

    loading.curLen = utf8.RuneCountInString(loading.model.menuTitle) + utf8.RuneCountInString(constants.MainLoadingText)

    fmt.Printf("%s%s%s",
        strings.Repeat(" ", loading.model.menuTitleStartColumn),
        SetFgStyle(loading.model.menuTitle, termenv.ANSIGreen),
        SetFgStyle(constants.MainLoadingText, termProfile.Color("#5f5f5f")))

    termenv.MoveCursor(0, 0)
}

// 完成
func (loading *Loading)complete() {
    termenv.MoveCursor(loading.model.menuTitleStartRow, 0)

    spaceLen := loading.curLen - utf8.RuneCountInString(loading.model.menuTitle)
    if spaceLen < 0 {
        spaceLen = 0
    }
    fmt.Printf("%s%s%s",
        strings.Repeat(" ", loading.model.menuTitleStartColumn),
        SetFgStyle(loading.model.menuTitle, termenv.ANSIGreen),
        strings.Repeat("　", spaceLen))

    termenv.MoveCursor(0, 0)
}