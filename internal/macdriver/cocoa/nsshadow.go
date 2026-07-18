//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSShadow = objc.GetClass("NSShadow")
}

var class_NSShadow objc.Class

var (
	sel_setShadowBlurRadius = objc.RegisterName("setShadowBlurRadius:")
	sel_setShadowColor      = objc.RegisterName("setShadowColor:")
	sel_setShadowOffset     = objc.RegisterName("setShadowOffset:")
	sel_setShadow           = objc.RegisterName("setShadow:")
	sel_CGColor             = objc.RegisterName("CGColor")
)

type NSShadow struct {
	core.NSObject
}

func NSShadow_alloc() NSShadow {
	return NSShadow{
		core.NSObject{
			ID: objc.ID(class_NSShadow).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
		},
	}
}

func (s NSShadow) SetShadowBlurRadius(radius CGFloat) {
	s.Send(sel_setShadowBlurRadius, radius)
}

func (s NSShadow) SetShadowColor(color NSColor) {
	s.Send(sel_setShadowColor, color.ID)
}

// SetShadowOffset sets the shadow offset. w=horizontal, h=vertical.
func (s NSShadow) SetShadowOffset(w, h CGFloat) {
	// NSSize is passed as two CGFloat values
	s.Send(sel_setShadowOffset, w, h)
}

// SetShadow sets the shadow on an NSView.
func SetViewShadow(view NSView, shadow NSShadow) {
	view.Send(sel_setShadow, shadow.ID)
}

// CGColor returns the CGColorRef as uintptr for use with CALayer.
func (c NSColor) CGColorRef() uintptr {
	return uintptr(c.Send(sel_CGColor))
}
