package ui

import "github.com/muesli/termenv"

var (
    focusedPrompt       = termenv.String("> ").Foreground(GetPrimaryColor()).String()
    blurredPrompt       = "> "
    focusedSubmitButton = "[ " + termenv.String("确认").Foreground(GetPrimaryColor()).String() + " ]"
    blurredSubmitButton = "[ " + termenv.String("确认").Foreground(termenv.ANSIBrightBlack).String() + " ]"
)
