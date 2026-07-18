//go:build darwin

package cocoa

import "testing"

func TestNSScreenFrameAndDisplayID(t *testing.T) {
	screen := NSScreen_MainScreen()
	if screen.ID == 0 {
		t.Fatal("NSScreen_MainScreen() returned nil")
	}
	frame := screen.Frame()
	if frame.Size.Width <= 0 || frame.Size.Height <= 0 {
		t.Fatalf("screen.Frame() = %#v", frame)
	}
	displayID := screen.DisplayID()
	if displayID == 0 {
		t.Fatal("screen.DisplayID() returned 0")
	}
	if _, found := NSScreen_WithDisplayID(displayID); !found {
		t.Fatalf("NSScreen_WithDisplayID(%d) did not find the main screen", displayID)
	}
}
