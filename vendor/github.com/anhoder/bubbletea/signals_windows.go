//go:build windows
// +build windows

package tea

import (
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// listenForResize is not available on windows because windows does not
// implement syscall.SIGWINCH.
func listenForResize(output *os.File, msgs chan Msg, errs chan error) {
	ticker := time.NewTicker(time.Second)
	var width, height int
	for range ticker.C {
		w, h, err := terminal.GetSize(int(output.Fd()))
		if err != nil {
			errs <- err
		}
		if w != width || h != height {
			width, height = w, h
			msgs <- WindowSizeMsg{w, h}
		}
	}
}
