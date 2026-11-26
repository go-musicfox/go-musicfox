//go:build !windows

package slogx

import (
	"os"

	"golang.org/x/sys/unix"
)

// redirectStderr 将 fd 2 直接重定向到日志文件，使 Go 运行时（panic 堆栈、
// fatal error）与 cgo 的输出同步写入日志，进程异常退出时也不丢失。
// 原始 stderr 保存为 stderrWriter，供崩溃提示等必须直达终端的输出使用。
func redirectStderr(f *os.File) {
	origFd, err := unix.Dup(2)
	if err != nil {
		stderrWriter = os.Stderr
		return
	}
	if err := unix.Dup2(int(f.Fd()), 2); err != nil {
		_ = unix.Close(origFd)
		stderrWriter = os.Stderr
		return
	}
	stderrWriter = os.NewFile(uintptr(origFd), "stderr")
}
