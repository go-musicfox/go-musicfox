//go:build windows

package player

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

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

// pipeAddr implements net.Addr for named pipe
var _ net.Addr = (*pipeAddr)(nil)

type pipeAddr string

func (a pipeAddr) Network() string { return "pipe" }
func (a pipeAddr) String() string  { return string(a) }

// pipeConn 实现 net.Conn 接口，基于 Windows 命名管道句柄
// 使用重叠 I/O (OVERLAPPED) 支持读写超时
var _ net.Conn = (*pipeConn)(nil)

type pipeConn struct {
	handle windows.Handle

	readOverlapped  windows.Overlapped
	writeOverlapped windows.Overlapped
	readEvent       windows.Handle
	writeEvent      windows.Handle

	readDeadline  time.Time
	writeDeadline time.Time

	closed bool
	mu     sync.Mutex
}

// newPipeConn 创建包装命名管道句柄的 pipeConn
func newPipeConn(handle windows.Handle) (*pipeConn, error) {
	readEvent, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("创建读取事件失败: %v", err)
	}
	writeEvent, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		windows.CloseHandle(readEvent)
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("创建写入事件失败: %v", err)
	}
	return &pipeConn{
		handle:          handle,
		readEvent:       readEvent,
		writeEvent:      writeEvent,
		readOverlapped:  windows.Overlapped{HEvent: readEvent},
		writeOverlapped: windows.Overlapped{HEvent: writeEvent},
	}, nil
}

func (c *pipeConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0, net.ErrClosed
	}
	c.mu.Unlock()

	if err := windows.ResetEvent(c.readEvent); err != nil {
		return 0, fmt.Errorf("重置读取事件失败: %v", err)
	}

	var n uint32
	err := windows.ReadFile(c.handle, b, &n, &c.readOverlapped)
	if err == windows.ERROR_IO_PENDING {
		timeout := c.deadlineToWait(c.readDeadline)
		if timeout == 0 {
			_ = windows.CancelIoEx(c.handle, &c.readOverlapped)
			return 0, os.ErrDeadlineExceeded
		}

		s, err := windows.WaitForSingleObject(c.readEvent, timeout)
		if err != nil {
			return 0, fmt.Errorf("等待读取失败: %v", err)
		}
		if s == uint32(windows.WAIT_TIMEOUT) {
			_ = windows.CancelIoEx(c.handle, &c.readOverlapped)
			return 0, os.ErrDeadlineExceeded
		}

		err = windows.GetOverlappedResult(c.handle, &c.readOverlapped, &n, false)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	if n == 0 {
		return 0, io.EOF
	}

	return int(n), nil
}

func (c *pipeConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0, net.ErrClosed
	}
	c.mu.Unlock()

	if err := windows.ResetEvent(c.writeEvent); err != nil {
		return 0, fmt.Errorf("重置写入事件失败: %v", err)
	}

	var n uint32
	err := windows.WriteFile(c.handle, b, &n, &c.writeOverlapped)
	if err == windows.ERROR_IO_PENDING {
		timeout := c.deadlineToWait(c.writeDeadline)
		if timeout == 0 {
			_ = windows.CancelIoEx(c.handle, &c.writeOverlapped)
			return 0, os.ErrDeadlineExceeded
		}

		s, err := windows.WaitForSingleObject(c.writeEvent, timeout)
		if err != nil {
			return 0, fmt.Errorf("等待写入失败: %v", err)
		}
		if s == uint32(windows.WAIT_TIMEOUT) {
			_ = windows.CancelIoEx(c.handle, &c.writeOverlapped)
			return 0, os.ErrDeadlineExceeded
		}

		err = windows.GetOverlappedResult(c.handle, &c.writeOverlapped, &n, false)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	return int(n), nil
}

func (c *pipeConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return net.ErrClosed
	}
	c.closed = true

	// 取消所有未完成的 I/O 操作
	_ = windows.CancelIoEx(c.handle, &c.readOverlapped)
	_ = windows.CancelIoEx(c.handle, &c.writeOverlapped)

	// 关闭事件句柄
	if c.readEvent != 0 {
		_ = windows.CloseHandle(c.readEvent)
		c.readEvent = 0
	}
	if c.writeEvent != 0 {
		_ = windows.CloseHandle(c.writeEvent)
		c.writeEvent = 0
	}

	// 关闭管道句柄
	return windows.CloseHandle(c.handle)
}

func (c *pipeConn) LocalAddr() net.Addr  { return pipeAddr(ipcServerPath()) }
func (c *pipeConn) RemoteAddr() net.Addr { return pipeAddr(ipcServerPath()) }

func (c *pipeConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDeadline = t
	c.writeDeadline = t
	c.mu.Unlock()
	return nil
}

func (c *pipeConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDeadline = t
	c.mu.Unlock()
	return nil
}

func (c *pipeConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	c.writeDeadline = t
	c.mu.Unlock()
	return nil
}

// deadlineToWait 将 deadline 转换为 WaitForSingleObject 的超时毫秒数
func (c *pipeConn) deadlineToWait(deadline time.Time) uint32 {
	if deadline.IsZero() {
		return windows.INFINITE
	}
	d := time.Until(deadline)
	if d <= 0 {
		return 0
	}
	ms := d.Milliseconds()
	if ms <= 0 {
		return 0
	}
	return uint32(ms)
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
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("打开命名管道失败: %v, path: %s", err, ipcServerPath())
	}

	return newPipeConn(handle)
}
