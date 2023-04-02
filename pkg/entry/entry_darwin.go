//go:build darwin

package entry

import (
	"github.com/go-musicfox/go-musicfox/pkg/cocoa"
	"github.com/go-musicfox/go-musicfox/utils"
)

func AppEntry() {
	defer utils.Recover(false)

	var app = cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.SetDelegate(cocoa.DefaultAppDelegate())
	app.ActivateIgnoringOtherApps(true)

	go func() {
		defer utils.Recover(false)
		runCLI()

		app.Terminate(0)
	}()

	app.Run()
}
