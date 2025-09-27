package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
)

// Version 应用版本
var Version = "2.0.0-mpv"

// BuildTime 构建时间
var BuildTime = "unknown"

// GitCommit Git提交哈希
var GitCommit = "unknown"

func main() {
	// 解析命令行参数
	var (
		configPath = pflag.StringP("config", "c", "", "配置文件路径")
		logLevel   = pflag.StringP("log-level", "l", "info", "日志级别 (debug, info, warn, error)")
		version    = pflag.BoolP("version", "v", false, "显示版本信息")
		help       = pflag.BoolP("help", "h", false, "显示帮助信息")
	)
	pflag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("MusicFox MPV v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 显示帮助信息
	if *help {
		printUsage()
		os.Exit(0)
	}

	// 创建应用实例
	app, err := NewMPVApp(*configPath, *logLevel)
	if err != nil {
		log.Fatalf("Failed to create MPV app: %v", err)
	}

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// 启动应用
	if err := app.Run(ctx); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	fmt.Println("MusicFox MPV shutdown complete")
}

func printUsage() {
	fmt.Printf(`MusicFox MPV v%s - 专用MPV播放器版本

`, Version)
	fmt.Println("Usage:")
	fmt.Println("  musicfox-mpv [flags] [command]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  play <song_url>     播放指定歌曲")
	fmt.Println("  pause               暂停播放")
	fmt.Println("  resume              恢复播放")
	fmt.Println("  stop                停止播放")
	fmt.Println("  next                下一首")
	fmt.Println("  prev                上一首")
	fmt.Println("  volume <level>      设置音量 (0-100)")
	fmt.Println("  status              显示播放状态")
	fmt.Println("  playlist            播放列表管理")
	fmt.Println("  interactive         进入交互模式")
	fmt.Println()
	fmt.Println("Flags:")
	pflag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  musicfox-mpv play /path/to/song.mp3")
	fmt.Println("  musicfox-mpv --config /path/to/config.yaml interactive")
	fmt.Println("  musicfox-mpv playlist create \"My Playlist\"")
	fmt.Println()
	fmt.Println("Note: This version is specifically optimized for MPV player backend.")
	fmt.Println("Make sure MPV is installed and available in your PATH.")
}