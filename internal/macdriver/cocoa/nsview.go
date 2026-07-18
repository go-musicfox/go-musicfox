//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSView = objc.GetClass("NSView")
}

var class_NSView objc.Class

var (
	sel_initWithFrame          = objc.RegisterName("initWithFrame:")
	sel_addSubview             = objc.RegisterName("addSubview:")
	sel_setWantsLayer          = objc.RegisterName("setWantsLayer:")
	sel_removeFromSuperview    = objc.RegisterName("removeFromSuperview")
	sel_setFrameOrigin         = objc.RegisterName("setFrameOrigin:")
	sel_setFrameSize           = objc.RegisterName("setFrameSize:")
	sel_setViewFrame             = objc.RegisterName("setFrame:")
	sel_addSubviewPositioned     = objc.RegisterName("addSubview:positioned:relativeTo:")
)

type NSView struct {
	core.NSObject
}

func NSView_alloc() NSView {
	return NSView{
		core.NSObject{
			ID: objc.ID(class_NSView).Send(macdriver.SEL_alloc),
		},
	}
}

func (v NSView) InitWithFrame(frameRect NSRect) NSView {
	v.ID = v.Send(sel_initWithFrame,
		frameRect.Origin.X, frameRect.Origin.Y,
		frameRect.Size.Width, frameRect.Size.Height,
	)
	return v
}

func (v NSView) AddSubview(subview NSView) {
	v.Send(sel_addSubview, subview.ID)
}

func (v NSView) SetWantsLayer(wantsLayer bool) {
	v.Send(sel_setWantsLayer, wantsLayer)
}

func (v NSView) RemoveFromSuperview() {
	v.Send(sel_removeFromSuperview)
}

// SetFrameOrigin sets the view's frame origin (x, y) in its superview.
func (v NSView) SetFrameOrigin(x, y CGFloat) {
	v.Send(sel_setFrameOrigin, x, y)
}

// SetFrameSize sets the view's frame size (width, height).
func (v NSView) SetFrameSize(w, h CGFloat) {
	v.Send(sel_setFrameSize, w, h)
}

// NSWindowOrderingMode constants
const (
	NSWindowAbove = 1
	NSWindowBelow = -1
)

// SetFrame sets the view's frame (x, y, width, height).
func (v NSView) SetFrame(x, y, w, h CGFloat) {
	v.Send(sel_setViewFrame, x, y, w, h)
}

// AddSubviewPositioned adds a subview at the specified position relative to another view.
// place: NSWindowAbove or NSWindowBelow.
// otherView: the view to place relative to (pass NSView{} with ID 0 for nil).
func (v NSView) AddSubviewPositioned(subview NSView, place int, otherView NSView) {
	var otherID objc.ID
	if otherView.ID != 0 {
		otherID = otherView.ID
	}
	v.Send(sel_addSubviewPositioned, subview.ID, place, otherID)
}
