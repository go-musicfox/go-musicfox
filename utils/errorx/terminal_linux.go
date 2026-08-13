//go:build linux

package errorx

import (
	"io"
	"os"

	"golang.org/x/sys/unix"

	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// savedTermios 为程序启动时保存的原始终端状态，崩溃时写回以恢复终端。
var savedTermios *unix.Termios

// SaveTerminalState 保存当前终端状态（termios），需在程序最开头、
// 任何终端模式修改之前调用。stdin 非终端时静默跳过。
func SaveTerminalState() {
	t, err := unix.IoctlGetTermios(0, unix.TCGETS)
	if err != nil {
		return
	}
	savedTermios = t
}

// RestoreTerminalState 将终端状态恢复为 SaveTerminalState 时保存的值。
func RestoreTerminalState() {
	if savedTermios != nil {
		_ = unix.IoctlSetTermios(0, unix.TCSETS, savedTermios)
	}
}

// restoreStderr 将 fd 2 恢复为 slogx 保存的原始 stderr。
// 包装进程因 import 副作用初始化了日志（重定向了自身 stderr），
// 需在 exec 子进程前恢复，使子进程继承真正的 stderr。
func restoreStderr() {
	if w, ok := slogx.StderrWriter().(*os.File); ok {
		_ = unix.Dup2(int(w.Fd()), 2)
	}
}

// childStderr 返回 exec 子进程应继承的 stderr。
// Unix 上子进程自身的 slogx.Init 会重定向 fd 2 到日志文件，
// 因此子进程直接继承终端 stderr 即可。
func childStderr() io.Writer {
	return os.Stderr
}
