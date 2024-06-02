//go:build darwin

package runtime

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

func Run(f func()) {
	defer errorx.Recover(false)

	app := cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)

	delegate := cocoa.DefaultAppDelegate()
	delegate.RegisterDidFinishLaunchingCallback(func(_ objc.ID) {
		errorx.Go(func() {
			core.Autorelease(func() {
				f()
				app.Terminate(0)
			})
		})
	})
	app.SetDelegate(delegate)
	app.Run()
}
