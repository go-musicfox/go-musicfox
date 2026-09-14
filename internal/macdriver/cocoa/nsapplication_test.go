//go:build darwin

package cocoa

import (
	"os"
	"runtime"
	"testing"
)

// AppKit must run on the startup thread, including in the test executable.
func init() {
	runtime.LockOSThread()
}

func TestMain(m *testing.M) {
	app := NSApp()
	if app.ID == 0 {
		panic("app init error")
	}

	app.SetActivationPolicy(NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)

	delegate := DefaultAppDelegate()
	app.SetDelegate(delegate)

	go func() {
		code := m.Run()
		os.Exit(code)
	}()

	app.Run()
}
