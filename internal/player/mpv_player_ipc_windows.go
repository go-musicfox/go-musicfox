//go:build windows

package player

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/sys/windows"
)

// ipcServerPath 返回 Windows named pipe IPC 服务器路径
func ipcServerPath() string {
	return `\\.\pipe\mpvsocket`
}

// ipcLogPath 返回 mpv 日志文件路径（Windows 临时目录）
func ipcLogPath() string {
	return os.TempDir() + `\mpvipc.log`
}

// dialIPC 建立 Windows named pipe 连接到 mpv IPC 服务器
func (p *mpvPlayer) dialIPC() (net.Conn, error) {
	pathPtr, err := windows.UTF16PtrFromString(ipcServerPath())
	if err != nil {
		return nil, fmt.Errorf("编码管道路径失败: %v", err)
	}

	handle, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("打开命名管道失败: %v", err)
	}

	file := os.NewFile(uintptr(handle), "mpvpipe")
	conn, err := net.FileConn(file)
	file.Close()
	if err != nil {
		return nil, fmt.Errorf("包装命名管道连接失败: %v", err)
	}
	return conn, nil
}
