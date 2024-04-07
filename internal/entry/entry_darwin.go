//go:build darwin

package entry

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/utils"
)

func AppEntry() {
	defer utils.Recover(false)

	var app = cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)

	delegate := cocoa.DefaultAppDelegate()
	delegate.RegisterDidFinishLaunchingCallback(func(_ objc.ID) {
		go utils.PanicRecoverWrapper(false, func() {
			core.Autorelease(func() {
				runCLI()
				app.Terminate(0)
			})
		})
	})
	app.SetDelegate(delegate)
	app.Run()
}
