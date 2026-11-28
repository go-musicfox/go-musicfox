package errorx

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// maxCrashHeadLines 限制终端崩溃提示中展示的堆栈头行数，避免刷屏，
// 同时保证能覆盖首个 goroutine 的关键帧。
const maxCrashHeadLines = 24

func Recover(ignore bool) (hasCaught bool) {
	err := recover()
	if err != nil {
		stack := debug.Stack()
		slog.Error("catch panic", slog.Any("error", err), slog.Any("stack", stack))
		if ignore {
			hasCaught = true
			return
		}
		crashReport(err, stack)
		os.Exit(1)
	}
	return
}

// crashReport 在 panic 已记录到日志后，恢复终端并打印崩溃提示，引导用户
// 创建 GitHub issue 并附上日志文件。输出经 StderrWriter 直达原始终端
// （不经过 stderr 重定向管道，避免被 os.Exit 截断）。
func crashReport(err any, stack []byte) {
	RestoreTerminalState()
	_, _ = fmt.Fprint(slogx.StderrWriter(),
		"\x1b[?1049l\x1b[?25h\x1b[0m", formatCrashReport(err, stack))
}

// formatCrashReport 构造崩溃提示文案，含错误、堆栈头、日志路径、issue 链接
// 与版本环境信息。除完整错误写入日志外，终端额外展示首个 goroutine 前若干行，
// 避免用户反馈时只贴出无关的日志尾部而丢失根因。
func formatCrashReport(err any, stack []byte) string {
	version := types.AppVersion
	if types.BuildTags != "" {
		version += " [" + types.BuildTags + "]"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\nmusicfox 发生未预期的错误，已退出。\n\n")
	fmt.Fprintf(&b, "错误: %v\n", err)
	if head := stackHead(stack, maxCrashHeadLines); len(head) > 0 {
		b.WriteString("\n崩溃堆栈（首个 goroutine）:\n")
		for _, line := range head {
			fmt.Fprintf(&b, "  %s\n", line)
		}
		if len(head) >= maxCrashHeadLines {
			b.WriteString("  ...\n")
		}
	}
	fmt.Fprintf(&b, "\n错误详情与完整堆栈已写入日志: %s\n", slogx.LogFilePath())
	b.WriteString("\n请前往 GitHub 创建 Issue 反馈此问题，并附上上述日志文件:\n")
	b.WriteString("https://github.com/go-musicfox/go-musicfox/issues/new?template=1_bug_report.yml\n")
	fmt.Fprintf(&b, "\n版本: %s\n平台: %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	return b.String()
}

// stackHead 返回堆栈文本的前 n 行，用于在终端展示首个 goroutine。
func stackHead(stack []byte, n int) []string {
	if len(stack) == 0 || n <= 0 {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(stack), "\n"), "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return lines
}

func PanicRecoverWrapper(ignorePanic bool, f func()) {
	defer Recover(ignorePanic)
	f()
}

func Go(f func(), ignorePanic ...bool) {
	var ignore bool
	if len(ignorePanic) > 0 {
		ignore = ignorePanic[0]
	}
	go PanicRecoverWrapper(ignore, f)
}

func WaitGoStart(f func(), ignorePanic ...bool) {
	var wait = make(chan struct{})
	Go(func() {
		Go(f, ignorePanic...)
		wait <- struct{}{}
	}, ignorePanic...)
	<-wait
}
