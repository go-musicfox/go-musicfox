//go:build darwin
// +build darwin

package entry

import (
	"github.com/progrium/macdriver/cocoa"
	"go-musicfox/utils"
)

func AppEntry() {
	defer utils.Recover(false)

	go func() {
		defer utils.Recover(false)
		runCLI()
	}()

	app := cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
