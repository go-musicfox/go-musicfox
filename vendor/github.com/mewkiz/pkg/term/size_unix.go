// +build !windows

package term

import (
	"syscall"
	"unsafe"
)

// winSize contains the dimentions of a terminal.
type winSize struct {
	Row, Col       uint16
	XPixel, YPixel uint16
}

// Size returns the col and row size for the active terminal.
func Size() (col, row int, err error) {
	var ws winSize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(0), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0, errno
	}
	col = int(ws.Col)
	row = int(ws.Row)
	return col, row, nil
}
