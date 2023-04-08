//go:build darwin

package cocoa

import (
	"fmt"
	"testing"

	"github.com/ebitengine/purego/objc"
)

var (
	val int
	c   = make(chan struct{}, 1)
)

func TestMain(m *testing.M) {
	app := NSApp()
	if app.ID == 0 {
		panic("app init error")
	}

	app.SetActivationPolicy(NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)

	delegate := DefaultAppDelegate()
	delegate.RegisterDidFinishLaunchingCallback(func(_ objc.ID) {
		val = 111
		c <- struct{}{}
	})
	app.SetDelegate(delegate)

	go func() {
		m.Run()
		app.Terminate(0)
	}()

	app.Run()
}

func TestNSApplication(t *testing.T) {
	<-c
	fmt.Println(val)
	if val != 111 {
		panic("error")
	}
}
