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

func GetFocusedButton(text string) string {
	return fmt.Sprintf("[ %s ]", termenv.String(text).Foreground(GetPrimaryColor()).String())
}

func GetBlurredButton(text string) string {
	return fmt.Sprintf("[ %s ]", termenv.String(text).Foreground(termenv.ANSIBrightBlack).String())
}

func GetFocusedSubmitButton() string {
	if focusedSubmitButton != "" {
		return focusedSubmitButton
	}
	focusedSubmitButton = GetFocusedButton("确认")
	return focusedSubmitButton
}

func GetBlurredSubmitButton() string {
	if blurredSubmitButton != "" {
		return blurredSubmitButton
	}
	blurredSubmitButton = GetBlurredButton("确认")
	return blurredSubmitButton
}
