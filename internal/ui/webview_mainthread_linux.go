//go:build linux

package ui

import (
	"sync/atomic"

	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

// gtkMainQueue hands GTK/WebKit work to the process main thread. WebKitGTK
// 2.54 requires WTF::initializeMainThread() to run on the process main thread
// (tid == pid); otherwise it aborts in JSC. Bubbletea owns the Go main
// goroutine, so main() starts this service on the locked process main thread.
var gtkMainQueue = make(chan func(), 1)

// gtkMainStarted is set once ServeWebviewMainThread is running, so Open() can
// fall back to a dedicated goroutine in tests (where main() is not involved).
var gtkMainStarted atomic.Bool

// ServeWebviewMainThread runs GTK/WebKit work on the process main thread.
// It blocks forever and is started by main() on Linux.
func ServeWebviewMainThread() {
	gtkMainStarted.Store(true)
	for f := range gtkMainQueue {
		errorx.PanicRecoverWrapper(false, f)
	}
}

// runOnWebviewMainThread schedules f on the process main thread. It returns
// false when the service is not running (e.g. in tests) or its single queue
// slot is already occupied, so the caller can fall back to a goroutine.
func runOnWebviewMainThread(f func()) bool {
	if !gtkMainStarted.Load() {
		return false
	}
	select {
	case gtkMainQueue <- f:
		return true
	default:
		return false
	}
}
