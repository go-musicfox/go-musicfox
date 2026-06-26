package model

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/util"
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

	focusedPrompt = lipgloss.NewStyle().Foreground(util.GetPrimaryColor()).Render("> ")

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
	return fmt.Sprintf("[ %s ]", lipgloss.NewStyle().Foreground(util.GetPrimaryColor()).Render(text))
}

func GetBlurredButton(text string) string {
	return fmt.Sprintf("[ %s ]", lipgloss.NewStyle().Foreground(lipgloss.BrightBlack).Render(text))
}

func GetFocusedSubmitButton() string {
	if focusedSubmitButton != "" {
		return focusedSubmitButton
	}
	focusedSubmitButton = GetFocusedButton(Submit)
	return focusedSubmitButton
}

func GetBlurredSubmitButton() string {
	if blurredSubmitButton != "" {
		return blurredSubmitButton
	}
	blurredSubmitButton = GetBlurredButton(Submit)
	return blurredSubmitButton
}
