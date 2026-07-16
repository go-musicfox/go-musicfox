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
	sel_sizeWithAttributes   = objc.RegisterName("sizeWithAttributes:")
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

// NSFontAttributeName is the attributed-string key for a font.
// Its underlying Foundation string value is "NSFont".
const NSFontAttributeName = "NSFont"

// MeasureTextSize returns the rendered size of text drawn with font, using
// -[NSString sizeWithAttributes:]. This reflects the font's real glyph advances
// (proportional Latin widths, full-width CJK and punctuation) rather than a
// per-character heuristic.
func MeasureTextSize(text string, font NSFont) CGSize {
	nsStr := core.NSString_alloc().InitWithUTF8String(text)
	defer nsStr.Release()

	key := core.NSString_alloc().InitWithUTF8String(NSFontAttributeName)
	defer key.Release()

	attrs := core.NSMutableDictionary_init()
	defer attrs.Release()
	attrs.SetValueForKey(key, font.NSObject)

	return objc.Send[CGSize](nsStr.ID, sel_sizeWithAttributes, attrs.ID)
}
