//go:build windows
// +build windows

package termenv

import (
	"golang.org/x/sys/windows"
)

func colorProfile() Profile {
	return TrueColor
}

func foregroundColor() Color {
	// default gray
	return ANSIColor(7)
}

func backgroundColor() Color {
	// default black
	return ANSIColor(0)
}

// EnableWindowsANSI enables virtual terminal processing on Windows platforms.
// This allows the use of ANSI escape sequences in Windows console applications.
// Ensure this gets called before anything gets rendered with termenv.
// Returns the original console mode and an error if one occurred.
func EnableWindowsANSIConsole() (uint32, error) {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return 0, err
	}

	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return 0, err
	}

	// See https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING {
		vtpmode := mode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		if err := windows.SetConsoleMode(handle, vtpmode); err != nil {
			return 0, err
		}
	}

	return mode, nil
}

// RestoreWindowsConsole restores the console mode to a previous state.
func RestoreWindowsConsole(mode uint32) error {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return err
	}

	return windows.SetConsoleMode(handle, mode)
}
