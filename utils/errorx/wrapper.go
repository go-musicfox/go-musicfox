package errorx

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// childEnv 标记包装进程的子进程。
const childEnv = "MUSICFOX_INTERNAL_CHILD"

// RunWrapped 以包装进程方式运行应用本体：
//   - 子进程（app）退出码为 2（Go 运行时 fatal，如 cgo/purego 崩溃）时，
//     恢复终端并打印友好提示，引导用户提交 issue 并附上日志；
//   - 退出码为 1 时子进程已自行打印提示，不重复输出；
//   - 正常退出或信号终止（用户中断）时静默。
func RunWrapped(app func()) {
	// 无论子进程还是包装进程，都在任何终端模式修改前保存原始状态
	SaveTerminalState()

	if os.Getenv(childEnv) == "1" {
		app()
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	// slogx 的 import 副作用已在本进程初始化日志（重定向了 fd 2），
	// 恢复 stderr 后再 exec，保证子进程继承到真正的 stderr（终端）
	restoreStderr()
	cmd.Env = append(os.Environ(), childEnv+"=1")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	err := cmd.Run()
	if err == nil {
		return
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return
	}

	// 信号终止（如 Ctrl+C 中断）不提示
	if exitErr.ExitCode() < 0 {
		return
	}

	// 退出码 2: Go 运行时 fatal error（未捕获 panic / cgo/purego 崩溃），
	// 子进程无法在进程内提示，由包装进程恢复终端并输出
	if exitErr.ExitCode() == 2 {
		RestoreTerminalState()
		printWrapperCrashHint(exitErr.ExitCode())
	}

	os.Exit(exitErr.ExitCode())
}

// printWrapperCrashHint 打印崩溃提示：日志路径、日志尾部摘要与 issue 链接。
// 输出走 stdout/stderr（包装进程未被重定向），终端直接可见。
func printWrapperCrashHint(code int) {
	var b strings.Builder
	fmt.Fprintf(&b, "\nmusicfox 异常退出（退出码 %d），可能发生了未预期的错误。\n\n", code)
	fmt.Fprintf(&b, "错误详情已写入日志: %s\n", slogx.ResolveLogFilePath())
	if tail := logTail(slogx.ResolveLogFilePath(), 8); len(tail) > 0 {
		b.WriteString("\n日志尾部:\n")
		for _, line := range tail {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}
	b.WriteString("\n请前往 GitHub 创建 Issue 反馈此问题，并附上上述日志文件:\n")
	b.WriteString("https://github.com/go-musicfox/go-musicfox/issues/new?template=1_bug_report.yml\n")
	fmt.Fprintf(&b, "\n版本: %s\n平台: %s/%s\n", appVersion(), runtime.GOOS, runtime.GOARCH)

	_, _ = os.Stdout.WriteString(b.String())
}

// logTail 返回日志文件末尾最多 n 个非空行，用于提示中展示崩溃摘要。
func logTail(path string, n int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	const bufSize = 16 << 10 // 16 KiB
	fi, err := f.Stat()
	if err != nil {
		return nil
	}
	buf := make([]byte, bufSize)
	start := fi.Size() - int64(len(buf))
	if start < 0 {
		start = 0
	}
	read, err := f.ReadAt(buf, start)
	if err != nil && read == 0 {
		return nil
	}
	buf = buf[:read]

	lines := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	var tail []string
	for i := len(lines) - 1; i >= 0 && len(tail) < n; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			tail = append(tail, line)
		}
	}
	// 逆序还原
	for i, j := 0, len(tail)-1; i < j; i, j = i+1, j-1 {
		tail[i], tail[j] = tail[j], tail[i]
	}
	return tail
}

// appVersion 读取版本信息（-ldflags 注入），供提示展示。
func appVersion() string {
	version := types.AppVersion
	if types.BuildTags != "" {
		version += " [" + types.BuildTags + "]"
	}
	return version
}
