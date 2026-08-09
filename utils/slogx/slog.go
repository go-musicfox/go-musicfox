// Package slogx 统一管理应用日志：将 slog/log 输出与 stderr 写入日志文件，
// 并提供崩溃提示所需的原始 stderr writer。
package slogx

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

var levelVar slog.LevelVar

// LogFileName 日志文件名
const LogFileName = "musicfox.log"

// maxLogFileSize 日志文件超过该大小后，启动时轮转为 musicfox.log.old
const maxLogFileSize = 10 << 20 // 10 MiB

var (
	initOnce sync.Once
	logPath  string
	// logFile 为日志文件句柄（O_APPEND），供包装进程双写子进程 stderr 使用
	logFile *os.File
	// stderrWriter 为重定向前保存的原始 stderr，供崩溃提示等必须直达终端的输出使用
	stderrWriter io.Writer
)

func init() {
	Init()
}

// Init 初始化日志系统：轮转日志文件、接管 log/slog 输出、将 stderr 双写至日志与终端。
// 幂等，可在 main 入口显式调用以尽早生效。
func Init() {
	initOnce.Do(func() {
		logPath = ResolveLogFilePath()

		rotateLogFile(logPath)

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic(fmt.Sprintf("failed to open log file, err: %v", err))
		}
		logFile = f

		logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{AddSource: true, Level: &levelVar}))
		log.SetOutput(f)
		slog.SetDefault(logger)

		redirectStderr(f)

		writeBanner(f)
	})
}

// LogFilePath 返回当前日志文件路径；未初始化时返回空串。
func LogFilePath() string {
	return logPath
}

// LogFile 返回日志文件句柄（O_APPEND），供包装进程双写子进程 stderr 使用。
// 未初始化时返回 nil。
func LogFile() *os.File {
	return logFile
}

// ResolveLogFilePath 返回日志文件路径（不依赖 Init，供包装进程等场景使用）。
func ResolveLogFilePath() string {
	return filepath.Join(app.LogDir(), LogFileName)
}

// StderrWriter 返回重定向前保存的原始 stderr，用于崩溃提示等必须在进程退出前
// 直达终端的输出（经重定向管道写会被 os.Exit 截断）。
func StderrWriter() io.Writer {
	if stderrWriter != nil {
		return stderrWriter
	}
	return os.Stderr
}

// Error 构造携带错误堆栈的 slog 属性。
func Error(err any) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	return slog.String("error", fmt.Sprintf("%+v", err))
}

// Bytes 构造字节数组内容的 slog 属性。
func Bytes(k string, b []byte) slog.Attr {
	return slog.String(k, string(b))
}

// LevelVar 返回全局日志级别，用于动态调整。
func LevelVar() *slog.LevelVar {
	return &levelVar
}

// rotateLogFile 日志文件超过大小上限时直接清空，避免无限增长。
// 崩溃日志会因截断丢失，但 wrapper 进程的提示在退出时已从日志尾部
// 提取过摘要，此处优先保证磁盘占用可控。
func rotateLogFile(path string) {
	if fi, err := os.Stat(path); err == nil && fi.Size() > maxLogFileSize {
		_ = os.Truncate(path, 0)
	}
}

// writeBanner 写入启动横幅，便于 issue 排查环境信息。
func writeBanner(f *os.File) {
	version := types.AppVersion
	if types.BuildTags != "" {
		version += " [" + types.BuildTags + "]"
	}
	_, _ = fmt.Fprintf(f, "==== musicfox %s %s/%s started at %s, log: %s ====\n",
		version, runtime.GOOS, runtime.GOARCH, time.Now().Format(time.RFC3339), logPath)
}
