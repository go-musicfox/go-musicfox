package errorx

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

func Recover(ignore bool) (hasCaught bool) {
	err := recover()
	if err != nil {
		slog.Error("catch panic", slog.Any("error", err), slog.Any("stack", debug.Stack()))
		if ignore {
			hasCaught = true
			return
		}
		crashReport(err)
		os.Exit(1)
	}
	return
}

// crashReport 在 panic 已记录到日志后，恢复终端并打印崩溃提示，引导用户
// 创建 GitHub issue 并附上日志文件。输出经 StderrWriter 直达原始终端
// （不经过 stderr 重定向管道，避免被 os.Exit 截断）。
func crashReport(err any) {
	RestoreTerminalState()
	_, _ = fmt.Fprint(slogx.StderrWriter(),
		"\x1b[?1049l\x1b[?25h\x1b[0m", formatCrashReport(err))
}

// formatCrashReport 构造崩溃提示文案，含日志路径、issue 链接与版本环境信息。
func formatCrashReport(err any) string {
	version := types.AppVersion
	if types.BuildTags != "" {
		version += " [" + types.BuildTags + "]"
	}
	return fmt.Sprintf(`
musicfox 发生未预期的错误，已退出。

错误: %v

错误详情与堆栈已写入日志: %s

请前往 GitHub 创建 Issue 反馈此问题，并附上上述日志文件:
https://github.com/go-musicfox/go-musicfox/issues/new?template=1_bug_report.yml

版本: %s
平台: %s/%s
`, err, slogx.LogFilePath(), version, runtime.GOOS, runtime.GOARCH)
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
