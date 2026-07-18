//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSMutableParagraphStyle = objc.GetClass("NSMutableParagraphStyle")
}

var class_NSMutableParagraphStyle objc.Class

var (
	sel_para_setAlignment = objc.RegisterName("setAlignment:")
)

// NSParagraphStyleAttributeName is the key for paragraph style in attributed strings.
const NSParagraphStyleAttributeName = "NSParagraphStyle"

type NSMutableParagraphStyle struct {
	core.NSObject
}

// NewParagraphStyle creates a default NSMutableParagraphStyle.
func NewParagraphStyle() NSMutableParagraphStyle {
	return NSMutableParagraphStyle{
		core.NSObject{
			ID: objc.ID(class_NSMutableParagraphStyle).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
		},
	}
}

func (p NSMutableParagraphStyle) SetAlignment(alignment int) {
	p.Send(sel_para_setAlignment, alignment)
}
