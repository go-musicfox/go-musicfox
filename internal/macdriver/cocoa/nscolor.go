//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSColor = objc.GetClass("NSColor")
}

var class_NSColor objc.Class

var (
	sel_clearColor                = objc.RegisterName("clearColor")
	sel_colorWithRedGreenBlueAlpha = objc.RegisterName("colorWithRed:green:blue:alpha:")
	sel_colorWithWhiteAlpha       = objc.RegisterName("colorWithWhite:alpha:")
	sel_whiteColor                = objc.RegisterName("whiteColor")
	sel_blackColor                = objc.RegisterName("blackColor")
)

type NSColor struct {
	core.NSObject
}

func NSColor_ClearColor() NSColor {
	return NSColor{
		core.NSObject{
			ID: objc.ID(class_NSColor).Send(sel_clearColor),
		},
	}
}

func NSColor_ColorWithRedGreenBlueAlpha(red, green, blue, alpha CGFloat) NSColor {
	return NSColor{
		core.NSObject{
			ID: objc.ID(class_NSColor).Send(sel_colorWithRedGreenBlueAlpha, red, green, blue, alpha),
		},
	}
}

func NSColor_ColorWithWhiteAlpha(white, alpha CGFloat) NSColor {
	return NSColor{
		core.NSObject{
			ID: objc.ID(class_NSColor).Send(sel_colorWithWhiteAlpha, white, alpha),
		},
	}
}

func NSColor_WhiteColor() NSColor {
	return NSColor{
		core.NSObject{
			ID: objc.ID(class_NSColor).Send(sel_whiteColor),
		},
	}
}

func NSColor_BlackColor() NSColor {
	return NSColor{
		core.NSObject{
			ID: objc.ID(class_NSColor).Send(sel_blackColor),
		},
	}
}
