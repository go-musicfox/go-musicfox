//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSFont = objc.GetClass("NSFont")
}

var class_NSFont objc.Class

var (
	sel_fontWithNameSize   = objc.RegisterName("fontWithName:size:")
	sel_systemFontOfSize   = objc.RegisterName("systemFontOfSize:")
	sel_boldSystemFontOfSize = objc.RegisterName("boldSystemFontOfSize:")
)

type NSFont struct {
	core.NSObject
}

func NSFont_FontWithNameSize(fontName string, fontSize CGFloat) NSFont {
	nsName := core.NSString_alloc().InitWithUTF8String(fontName)
	defer nsName.Release()

	return NSFont{
		core.NSObject{
			ID: objc.ID(class_NSFont).Send(sel_fontWithNameSize, nsName.ID, fontSize),
		},
	}
}

func NSFont_SystemFontOfSize(fontSize CGFloat) NSFont {
	return NSFont{
		core.NSObject{
			ID: objc.ID(class_NSFont).Send(sel_systemFontOfSize, fontSize),
		},
	}
}

func NSFont_BoldSystemFontOfSize(fontSize CGFloat) NSFont {
	return NSFont{
		core.NSObject{
			ID: objc.ID(class_NSFont).Send(sel_boldSystemFontOfSize, fontSize),
		},
	}
}
