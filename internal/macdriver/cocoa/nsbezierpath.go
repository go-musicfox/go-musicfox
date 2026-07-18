//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSBezierPath = objc.GetClass("NSBezierPath")
}

var class_NSBezierPath objc.Class

var (
	sel_bezierPath            = objc.RegisterName("bezierPath")
	sel_moveToPoint           = objc.RegisterName("moveToPoint:")
	sel_lineToPoint           = objc.RegisterName("lineToPoint:")
	sel_curveToPoint          = objc.RegisterName("curveToPoint:controlPoint1:controlPoint2:")
	sel_CGPath                = objc.RegisterName("CGPath")
	sel_removeAllPoints       = objc.RegisterName("removeAllPoints")
	sel_appendBezierPathWithOvalInRect = objc.RegisterName("appendBezierPathWithOvalInRect:")
)

// NSBezierPath represents a Bezier path for drawing vector-based shapes.
type NSBezierPath struct {
	core.NSObject
}

// NSBezierPath_New creates a new empty NSBezierPath (retained, not autoreleased).
// Uses alloc/init instead of the bezierPath convenience method to avoid
// the object being deallocated when the autorelease pool drains.
func NSBezierPath_New() NSBezierPath {
	id := objc.ID(class_NSBezierPath).Send(macdriver.SEL_alloc)
	id = id.Send(macdriver.SEL_init)
	return NSBezierPath{
		NSObject: core.NSObject{
			ID: id,
		},
	}
}

// MoveToPoint starts a new subpath at the given point.
func (p NSBezierPath) MoveToPoint(x, y CGFloat) {
	p.Send(sel_moveToPoint, x, y)
}

// LineToPoint adds a line segment from the current point to the given point.
func (p NSBezierPath) LineToPoint(x, y CGFloat) {
	p.Send(sel_lineToPoint, x, y)
}

// CurveToPoint adds a cubic Bezier curve segment.
// endPoint: the curve end point
// controlPoint1: first control point
// controlPoint2: second control point
func (p NSBezierPath) CurveToPoint(endX, endY, cp1x, cp1y, cp2x, cp2y CGFloat) {
	p.Send(sel_curveToPoint, endX, endY, cp1x, cp1y, cp2x, cp2y)
}

// CGPath returns the Core Graphics path representation.
func (p NSBezierPath) CGPath() uintptr {
	return uintptr(p.Send(sel_CGPath))
}

// RemoveAllPoints removes all path elements, resetting to an empty path.
func (p NSBezierPath) RemoveAllPoints() {
	p.Send(sel_removeAllPoints)
}

// AppendBezierPathWithOvalInRect appends a closed oval path that fits within rect.
func (p NSBezierPath) AppendBezierPathWithOvalInRect(rect NSRect) {
	p.Send(sel_appendBezierPathWithOvalInRect,
		rect.Origin.X, rect.Origin.Y, rect.Size.Width, rect.Size.Height,
	)
}
