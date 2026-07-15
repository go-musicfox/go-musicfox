//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSScreen = objc.GetClass("NSScreen")
}

var class_NSScreen objc.Class

var (
	sel_mainScreen = objc.RegisterName("mainScreen")
	sel_frame      = objc.RegisterName("frame")
	sel_visibleFrame = objc.RegisterName("visibleFrame")
)

type NSScreen struct {
	core.NSObject
}

func NSScreen_MainScreen() NSScreen {
	return NSScreen{
		core.NSObject{
			ID: objc.ID(class_NSScreen).Send(sel_mainScreen),
		},
	}
}

func (s NSScreen) Frame() NSRect {
	_ = s.Send(sel_frame)
	// NOTE: NSRect (struct) cannot be directly returned via purego/objc.Send.
	// Use CoreGraphics CGDisplayPixelsWide/CGDisplayPixelsHigh instead for screen dimensions.
	return NSRect{}
}

// VisibleFrame returns the visible frame (excluding Dock and menu bar).
func (s NSScreen) VisibleFrame() NSRect {
	s.Send(sel_visibleFrame)
	return NSRect{}
}
