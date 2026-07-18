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
	sel_layer                   = objc.RegisterName("layer")
	sel_setCornerRadius         = objc.RegisterName("setCornerRadius:")
	sel_setMasksToBounds        = objc.RegisterName("setMasksToBounds:")
	sel_layerSetBackgroundColor = objc.RegisterName("setBackgroundColor:")
	sel_setFrame                = objc.RegisterName("setFrame:")
	sel_setOpacity              = objc.RegisterName("setOpacity:")
	sel_setContentsScale        = objc.RegisterName("setContentsScale:")
	sel_layerNew                = objc.RegisterName("layer")
	sel_layerShadowColor          = objc.RegisterName("setShadowColor:")
	sel_layerShadowRadius         = objc.RegisterName("setShadowRadius:")
	sel_layerShadowOpacity        = objc.RegisterName("setShadowOpacity:")
	sel_layerShadowOffset         = objc.RegisterName("setShadowOffset:")
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

// SetFrame sets the layer's frame (x, y, width, height).
func (l CALayer) SetFrame(x, y, w, h CGFloat) {
	l.Send(sel_setFrame,
		x, y, w, h,
	)
}

// SetOpacity sets the layer's opacity (0.0-1.0).
func (l CALayer) SetOpacity(opacity CGFloat) {
	l.Send(sel_setOpacity, opacity)
}

// SetContentsScale sets the layer's contents scale for Retina displays.
func (l CALayer) SetContentsScale(scale CGFloat) {
	l.Send(sel_setContentsScale, scale)
}

// CALayer_New creates and returns a new autoreleased CALayer instance.
func CALayer_New() CALayer {
	return CALayer{
		core.NSObject{
			ID: objc.ID(class_CALayer).Send(sel_layerNew),
		},
	}
}

// SetShadowColor sets the shadow color using a CGColorRef.
func (l CALayer) SetShadowColor(cgColor uintptr) {
	l.Send(sel_layerShadowColor, cgColor)
}

// SetShadowRadius sets the shadow blur radius.
func (l CALayer) SetShadowRadius(radius CGFloat) {
	l.Send(sel_layerShadowRadius, radius)
}

// SetShadowOpacity sets the shadow opacity (0.0-1.0).
func (l CALayer) SetShadowOpacity(opacity CGFloat) {
	l.Send(sel_layerShadowOpacity, opacity)
}

// SetShadowOffset sets the shadow offset (width, height).
func (l CALayer) SetShadowOffset(w, h CGFloat) {
	l.Send(sel_layerShadowOffset, w, h)
}
