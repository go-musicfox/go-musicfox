//go:build darwin
// +build darwin

package entry

import (
	"os"

	"go-musicfox/utils"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
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
