package util

import (
	"fmt"

	"charm.land/lipgloss/v2"
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

	focusedPrompt = lipgloss.NewStyle().Foreground(GetPrimaryColor()).Render("> ")

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
	return fmt.Sprintf("[ %s ]", lipgloss.NewStyle().Foreground(GetPrimaryColor()).Render(text))
}

func GetBlurredButton(text string) string {
	return fmt.Sprintf("[ %s ]", lipgloss.NewStyle().Foreground(lipgloss.BrightBlack).Render(text))
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
