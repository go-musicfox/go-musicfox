//go:build darwin
// +build darwin

package entry

import (
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
	"go-musicfox/utils"
	"os"
)

func AppEntry() {
	defer utils.Recover(false)

	go func() {
		defer utils.Recover(false)
		objc.Autorelease(func() {
			runCLI()
		})
		os.Exit(0)
	}()

	app := cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
