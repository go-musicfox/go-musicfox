//go:build darwin

package entry

import (
	"os"

	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/progrium/macdriver/cocoa"
)

func AppEntry() {
	defer utils.Recover(false)

	go func() {
		defer utils.Recover(false)
		runCLI()
		os.Exit(0)
	}()

	app := cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
