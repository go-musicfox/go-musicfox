//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSMutableAttributedString = objc.GetClass("NSMutableAttributedString")
}

var class_NSMutableAttributedString objc.Class

var (
	sel_attrStrInitWithString     = objc.RegisterName("initWithString:")
	sel_attrStrAddAttributeValueRange = objc.RegisterName("addAttribute:value:range:")
)

// NSForegroundColorAttributeName is the attribute key for text color.
const NSForegroundColorAttributeName = "NSColor"

type NSMutableAttributedString struct {
	core.NSObject
}

func NSMutableAttributedString_alloc() NSMutableAttributedString {
	return NSMutableAttributedString{
		core.NSObject{
			ID: objc.ID(class_NSMutableAttributedString).Send(macdriver.SEL_alloc),
		},
	}
}

func (s NSMutableAttributedString) InitWithString(str string) NSMutableAttributedString {
	nsStr := core.NSString_alloc().InitWithUTF8String(str)
	defer nsStr.Release()
	s.ID = s.Send(sel_attrStrInitWithString, nsStr.ID)
	return s
}

// AddAttribute adds a color attribute over the specified range.
// ObjC signature: addAttribute:(NSString *)name value:(id)value range:(NSRange)range
func (s NSMutableAttributedString) AddAttribute(attrName string, color NSColor, rng NSRange) {
	nsAttr := core.NSString_alloc().InitWithUTF8String(attrName)
	defer nsAttr.Release()
	s.Send(sel_attrStrAddAttributeValueRange, nsAttr.ID, color.ID, rng.Location, rng.Length)
}

// AddParagraphStyle sets the paragraph style for the entire string range.
func (s NSMutableAttributedString) AddParagraphStyle(style NSMutableParagraphStyle, rng NSRange) {
	nsAttr := core.NSString_alloc().InitWithUTF8String(NSParagraphStyleAttributeName)
	defer nsAttr.Release()
	s.Send(sel_attrStrAddAttributeValueRange, nsAttr.ID, style.ID, rng.Location, rng.Length)
}
