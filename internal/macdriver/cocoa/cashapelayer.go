//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_CAShapeLayer = objc.GetClass("CAShapeLayer")
}

var class_CAShapeLayer objc.Class

var (
	sel_shapeLayerNew    = objc.RegisterName("layer")
	sel_setPath          = objc.RegisterName("setPath:")
	sel_setStrokeColor   = objc.RegisterName("setStrokeColor:")
	sel_setFillColor     = objc.RegisterName("setFillColor:")
	sel_setLineWidth     = objc.RegisterName("setLineWidth:")
	sel_removeFromSuperlayer = objc.RegisterName("removeFromSuperlayer")
)

// CAShapeLayer is a Core Animation layer that draws a cubic Bezier spline.
type CAShapeLayer struct {
	CALayer
}

// CAShapeLayer_New creates a new autoreleased CAShapeLayer.
func CAShapeLayer_New() CAShapeLayer {
	return CAShapeLayer{
		CALayer: CALayer{
			NSObject: core.NSObject{
				ID: objc.ID(class_CAShapeLayer).Send(sel_shapeLayerNew),
			},
		},
	}
}

// SetPath sets the CGPathRef that defines the shape.
func (l CAShapeLayer) SetPath(cgPath uintptr) {
	l.Send(sel_setPath, cgPath)
}

// SetStrokeCGColor sets the stroke color using a CGColorRef.
func (l CAShapeLayer) SetStrokeCGColor(cgColor uintptr) {
	l.Send(sel_setStrokeColor, cgColor)
}

// SetFillCGColor sets the fill color using a CGColorRef.
func (l CAShapeLayer) SetFillCGColor(cgColor uintptr) {
	l.Send(sel_setFillColor, cgColor)
}

// SetLineWidth sets the stroked line width.
func (l CAShapeLayer) SetLineWidth(width CGFloat) {
	l.Send(sel_setLineWidth, width)
}

// RemoveFromSuperlayer removes this layer from its parent layer.
func (l CAShapeLayer) RemoveFromSuperlayer() {
	l.Send(sel_removeFromSuperlayer)
}
