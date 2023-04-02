//go:build darwin

package cocoa

import (
	"testing"
)

func TestNSApplication(t *testing.T) {
	app := NSApp()
	app.SetActivationPolicy(NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	app.SetDelegate(DefaultAppDelegate.ID)
	app.Run()
}
