//go:build darwin
// +build darwin

package entry

import (
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
	"go-musicfox/utils"
)

func AppEntry() {
	defer utils.Recover(false)

	go func() {
		defer utils.Recover(false)
		objc.Autorelease(func() {
			runCLI()
		})
	}()

	app := cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
