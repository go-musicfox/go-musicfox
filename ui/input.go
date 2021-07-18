package ui

import (
    "fmt"
    "github.com/muesli/termenv"
)

var (
    focusedPrompt,
    blurredPrompt,
    focusedSubmitButton,
    blurredSubmitButton string
)

func GetFocusedPrompt() string {
    if focusedPrompt != "" {
        return focusedPrompt
    }

    focusedPrompt = termenv.String("> ").Foreground(GetPrimaryColor()).String()

    return focusedPrompt
}

func GetBlurredPrompt() string {
    if blurredPrompt != "" {
        return blurredPrompt
    }

    blurredPrompt = "> "

    return blurredPrompt
}

func GetFocusedSubmitButton() string {
    if focusedSubmitButton != "" {
        return focusedSubmitButton
    }

    focusedSubmitButton = fmt.Sprintf("[ %s ]", termenv.String("确认").Foreground(GetPrimaryColor()).String())

    return focusedSubmitButton
}

func GetBlurredSubmitButton() string {
    if blurredSubmitButton != "" {
        return blurredSubmitButton
    }

    blurredSubmitButton = fmt.Sprintf("[ %s ]", termenv.String("确认").Foreground(termenv.ANSIBrightBlack).String())

    return blurredSubmitButton
}