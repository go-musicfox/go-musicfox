//go:build windows
// +build windows

package sysutil

import (
	"errors"
	"os/exec"
	"syscall"

	"github.com/gookit/goutil/sysutil/process"
)

// IsWin system. linux windows darwin
func IsWin() bool {
	return true
}

// IsWindows system. linux windows darwin
func IsWindows() bool {
	return true
}

// IsMac system
func IsMac() bool {
	return false
}

// IsDarwin system
func IsDarwin() bool {
	return false
}

// IsLinux system
func IsLinux() bool {
	return false
}

// Kill a process by pid
func Kill(pid int, signal syscall.Signal) error {
	return errors.New("not support")
}

// ProcessExists check process exists by pid
func ProcessExists(pid int) bool {
	return process.Exists(pid)
}

// OpenBrowser Open browser URL
//
// Mac：
//
//	open 'https://github.com/inhere'
//
// Linux:
//
//	xdg-open URL
//	x-www-browser 'https://github.com/inhere'
//
// Windows:
//
//	cmd /c start https://github.com/inhere
func OpenBrowser(URL string) error {
	return exec.Command("cmd", "/c", "start", URL).Run()
}
