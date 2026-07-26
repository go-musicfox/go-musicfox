package util

import (
	"fmt"

	"github.com/anhoder/foxful-cli/style"
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

	focusedPrompt = style.DefaultStyleSet().Prompt.Render("> ")

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
	return fmt.Sprintf("[ %s ]", style.DefaultStyleSet().Button.Render(text))
}

func GetBlurredButton(text string) string {
	return fmt.Sprintf("[ %s ]", style.DefaultStyleSet().ButtonBlurred.Render(text))
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
