package model

import (
	"fmt"

	"github.com/anhoder/foxful-cli/util"
	"github.com/muesli/termenv"
)

var (
	focusedPrompt,
	blurredPrompt,
	focusedSubmitButton,
	blurredSubmitButton string
)

const SubmitText = "确认"

func GetFocusedPrompt() string {
	if focusedPrompt != "" {
		return focusedPrompt
	}

	focusedPrompt = termenv.String("> ").Foreground(util.GetPrimaryColor()).String()

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
	return fmt.Sprintf("[ %s ]", termenv.String(text).Foreground(util.GetPrimaryColor()).String())
}

func GetBlurredButton(text string) string {
	return fmt.Sprintf("[ %s ]", termenv.String(text).Foreground(termenv.ANSIBrightBlack).String())
}

func GetFocusedSubmitButton() string {
	if focusedSubmitButton != "" {
		return focusedSubmitButton
	}
	focusedSubmitButton = GetFocusedButton(SubmitText)
	return focusedSubmitButton
}

func GetBlurredSubmitButton() string {
	if blurredSubmitButton != "" {
		return blurredSubmitButton
	}
	blurredSubmitButton = GetBlurredButton(SubmitText)
	return blurredSubmitButton
}
