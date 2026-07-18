//go:build !windows

package player

import (
	"net"
	"os"
	"path/filepath"
	"strings"
)

// ipcServerPath 返回 Unix socket IPC 服务器路径
func ipcServerPath() string {
	if IsTermux() {
		return "/data/data/com.termux/files/usr/tmp/mpvsocket"
	}
	return "/tmp/mpvsocket"
}

// ipcLogPath 返回 mpv 日志文件路径
func ipcLogPath() string {
	if IsTermux() {
		return "/data/data/com.termux/files/usr/tmp/mpvipc.log"
	}
	return "/tmp/mpvipc.log"
}

// dialIPC 建立 Unix socket 连接到 mpv IPC 服务器
func (p *mpvPlayer) dialIPC() (net.Conn, error) {
	return net.DialUnix("unix", nil, &net.UnixAddr{
		Name: ipcServerPath(),
		Net:  "unix",
	})
}

// IsTermux 检查当前是否在 Termux 环境中运行
func IsTermux() bool {
	// 方法1：检查特定环境变量
	if path, ok := os.LookupEnv("PREFIX"); ok {
		if strings.Contains(path, "com.termux") {
			return true
		}
	}

	// 方法2：检查特定目录是否存在
	termuxPaths := []string{
		"/data/data/com.termux/files/usr",
		"/data/data/com.termux/files/home",
	}

	for _, path := range termuxPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// 方法3：检查可执行文件路径
	if exe, err := os.Executable(); err == nil {
		if strings.Contains(filepath.Dir(exe), "com.termux") {
			return true
		}
	}

	return false
}
