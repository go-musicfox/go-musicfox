//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_CALayer = objc.GetClass("CALayer")
}

var class_CALayer objc.Class

var (
	sel_layer                = objc.RegisterName("layer")
	sel_setCornerRadius      = objc.RegisterName("setCornerRadius:")
	sel_setMasksToBounds     = objc.RegisterName("setMasksToBounds:")
	sel_layerSetBackgroundColor = objc.RegisterName("setBackgroundColor:")
)

type CALayer struct {
	core.NSObject
}

// Layer returns the CALayer backing the NSView.
func (v NSView) Layer() CALayer {
	return CALayer{
		core.NSObject{
			ID: v.Send(sel_layer),
		},
	}
}

func (l CALayer) SetCornerRadius(radius CGFloat) {
	l.Send(sel_setCornerRadius, radius)
}

func (l CALayer) SetMasksToBounds(masks bool) {
	l.Send(sel_setMasksToBounds, masks)
}

// SetBackgroundCGColor sets the background color using a CGColorRef.
func (l CALayer) SetBackgroundCGColor(cgColor uintptr) {
	l.Send(sel_layerSetBackgroundColor, cgColor)
}
