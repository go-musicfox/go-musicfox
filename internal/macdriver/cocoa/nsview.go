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
	sel_initWithFrame = objc.RegisterName("initWithFrame:")
	sel_addSubview    = objc.RegisterName("addSubview:")
	sel_setWantsLayer = objc.RegisterName("setWantsLayer:")
	sel_removeFromSuperview = objc.RegisterName("removeFromSuperview")
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
