//go:build windows

package errorx

import (
	"io"
	"os"

	"golang.org/x/sys/windows"

	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// savedInputMode / savedOutputMode 为程序启动时保存的 Windows 控制台
// 输入/输出模式，崩溃时写回以恢复控制台（对应 Unix 的 termios）。
// 控制台模式是控制台 buffer 的属性，进程退出后不会自动复位，
// TUI 崩溃后若不恢复，控制台会残留 raw 输入模式（不回显）等改动。
var (
	savedInputMode  *uint32
	savedOutputMode *uint32
)

// SaveTerminalState 保存当前控制台输入/输出模式（GetConsoleMode），
// 需在程序最开头、TUI 修改模式之前调用。
// stdin/stdout 非控制台（如重定向到管道）时静默跳过。
func SaveTerminalState() {
	in, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err == nil {
		var mode uint32
		if err = windows.GetConsoleMode(in, &mode); err == nil {
			savedInputMode = &mode
		}
	}
	out, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err == nil {
		var mode uint32
		if err = windows.GetConsoleMode(out, &mode); err == nil {
			savedOutputMode = &mode
		}
	}
}

// RestoreTerminalState 将控制台模式恢复为 SaveTerminalState 时保存的值。
// 输出模式额外保留 ENABLE_VIRTUAL_TERMINAL_PROCESSING：crashReport 在
// 恢复之后仍会输出 ANSI 转义序列（退出备用屏幕/显示光标/重置样式），
// 该序列在传统 conhost 上需要 VT 处理才能渲染。
func RestoreTerminalState() {
	if savedInputMode != nil {
		if in, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE); err == nil {
			_ = windows.SetConsoleMode(in, *savedInputMode)
		}
	}
	if savedOutputMode != nil {
		if out, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE); err == nil {
			_ = windows.SetConsoleMode(out, *savedOutputMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
		}
	}
}

// restoreStderr 在 Windows 上为 no-op：redirectStderr 不会重定向 fd 2
// （Go 运行时输出走进程启动时绑定的控制台句柄，纯 Go 无法重定向），
// 因此 exec 子进程前无需恢复。
func restoreStderr() {}

// childStderr 返回 exec 子进程应继承的 stderr。
// Windows 上 Go runtime 的 stderr 句柄在进程启动（osinit）时固定，进程内
// 无法重定向到日志文件，故通过管道接管子进程 stderr，双写终端与日志文件：
// 子进程崩溃（runtime fatal / panic 堆栈）时详情仍能落盘，不丢失。
func childStderr() io.Writer {
	var w io.Writer = os.Stderr
	if f := slogx.LogFile(); f != nil {
		w = io.MultiWriter(w, f)
	}
	return w
}
