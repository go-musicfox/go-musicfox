package util

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

const SubmitText = "Submit"

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
