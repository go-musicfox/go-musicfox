//go:build !windows
// +build !windows

package sysutil

import (
	"os/exec"
	"runtime"
	"syscall"
)

// IsWin system. linux windows darwin
func IsWin() bool {
	return false
}

// IsWindows system. linux windows darwin
func IsWindows() bool {
	return false
}

// IsMac system
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// IsDarwin system
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux system
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// Kill a process by pid
func Kill(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}

// ProcessExists check process exists by pid
func ProcessExists(pid int) bool {
	return nil == syscall.Kill(pid, 0)
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
	bin := "x-www-browser"
	if IsDarwin() {
		bin = "open"
	}

	return exec.Command(bin, URL).Run()
}
