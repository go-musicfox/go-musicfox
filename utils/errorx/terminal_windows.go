//go:build windows

package errorx

// SaveTerminalState 在 Windows 上为 no-op：控制台模式由 TUI 框架管理，
// 无 termios 概念。
func SaveTerminalState() {}

// RestoreTerminalState 在 Windows 上为 no-op。
func RestoreTerminalState() {}

// restoreStderr 在 Windows 上为 no-op（Windows 不重定向 stderr）。
func restoreStderr() {}
