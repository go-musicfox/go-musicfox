//go:build windows

package slogx

import "os"

// redirectStderr 在 Windows 上为 no-op：Go 运行时输出走进程启动时绑定的
// 句柄，纯 Go 无法重定向；标准库 log/slog 输出已由 Init 接管。
func redirectStderr(f *os.File) {
	_ = f
	stderrWriter = os.Stderr
}
