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
	sel_mainScreen        = objc.RegisterName("mainScreen")
	sel_screens           = objc.RegisterName("screens")
	sel_frame             = objc.RegisterName("frame")
	sel_visibleFrame      = objc.RegisterName("visibleFrame")
	sel_deviceDescription = objc.RegisterName("deviceDescription")
	sel_objectForKey      = objc.RegisterName("objectForKey:")
	sel_count             = objc.RegisterName("count")
	sel_objectAtIndex     = objc.RegisterName("objectAtIndex:")
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
	return objc.Send[NSRect](s.ID, sel_frame)
}

func (s NSScreen) DisplayID() uint32 {
	key := core.String("NSScreenNumber")
	defer key.Release()
	description := s.Send(sel_deviceDescription)
	var number core.NSNumber
	number.SetObjcID(description.Send(sel_objectForKey, key.ID))
	return uint32(number.IntValue())
}

func NSScreen_Screens() []NSScreen {
	screensID := objc.ID(class_NSScreen).Send(sel_screens)
	count := objc.Send[uint](screensID, sel_count)
	screens := make([]NSScreen, 0, count)
	for i := uint(0); i < count; i++ {
		screens = append(screens, NSScreen{NSObject: core.NSObject{ID: screensID.Send(sel_objectAtIndex, i)}})
	}
	return screens
}

func NSScreen_WithDisplayID(displayID uint32) (NSScreen, bool) {
	for _, screen := range NSScreen_Screens() {
		if screen.DisplayID() == displayID {
			return screen, true
		}
	}
	return NSScreen{}, false
}

// VisibleFrame returns the visible frame (excluding Dock and menu bar).
func (s NSScreen) VisibleFrame() NSRect {
	s.Send(sel_visibleFrame)
	return NSRect{}
}
