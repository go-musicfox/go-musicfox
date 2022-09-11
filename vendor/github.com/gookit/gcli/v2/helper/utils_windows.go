package helper

import (
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// GetTerminalSize for current console terminal.
func GetTerminalSize(refresh ...bool) (w int, h int) {
	if terminalWidth > 0 && len(refresh) > 0 && !refresh[0] {
		return terminalWidth, terminalHeight
	}

	var err error
	w, h, err = terminal.GetSize(int(syscall.Stdin))
	if err != nil {
		return
	}

	// cache result
	terminalWidth, terminalHeight = w, h
	return
}
