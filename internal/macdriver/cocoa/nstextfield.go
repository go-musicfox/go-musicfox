//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSTextField = objc.GetClass("NSTextField")
}

var class_NSTextField objc.Class

var (
	sel_setStringValue         = objc.RegisterName("setStringValue:")
	sel_stringValue            = objc.RegisterName("stringValue")
	sel_setFont                = objc.RegisterName("setFont:")
	sel_setTextColor           = objc.RegisterName("setTextColor:")
	sel_setBezeled             = objc.RegisterName("setBezeled:")
	sel_setBordered            = objc.RegisterName("setBordered:")
	sel_setDrawsBackground     = objc.RegisterName("setDrawsBackground:")
	sel_setEditable            = objc.RegisterName("setEditable:")
	sel_setSelectable          = objc.RegisterName("setSelectable:")
	sel_setAlignment           = objc.RegisterName("setAlignment:")
	sel_sizeToFit              = objc.RegisterName("sizeToFit")
	sel_setMaximumNumberOfLines = objc.RegisterName("setMaximumNumberOfLines:")
	sel_tfSetBackgroundColor   = objc.RegisterName("setBackgroundColor:")
)

type NSTextField struct {
	NSView
}

func NSTextField_alloc() NSTextField {
	return NSTextField{
		NSView{
			core.NSObject{
				ID: objc.ID(class_NSTextField).Send(macdriver.SEL_alloc),
			},
		},
	}
}

// InitWithFrame initializes the text field with a frame.
// Note: NSTextField overrides NSView's initWithFrame, so we use the same selector.
func (t NSTextField) InitWithFrame(frameRect NSRect) NSTextField {
	t.ID = t.Send(sel_initWithFrame,
		frameRect.Origin.X, frameRect.Origin.Y,
		frameRect.Size.Width, frameRect.Size.Height,
	)
	return t
}

func (t NSTextField) SetStringValue(value string) {
	nsStr := core.NSString_alloc().InitWithUTF8String(value)
	defer nsStr.Release()
	t.Send(sel_setStringValue, nsStr.ID)
}

func (t NSTextField) StringValue() string {
	id := t.Send(sel_stringValue)
	if id == 0 {
		return ""
	}
	return core.NSString{NSObject: core.NSObject{ID: id}}.String()
}

func (t NSTextField) SetFont(font NSFont) {
	t.Send(sel_setFont, font.ID)
}

func (t NSTextField) SetTextColor(color NSColor) {
	t.Send(sel_setTextColor, color.ID)
}

func (t NSTextField) SetBezeled(bezeled bool) {
	t.Send(sel_setBezeled, bezeled)
}

func (t NSTextField) SetBordered(bordered bool) {
	t.Send(sel_setBordered, bordered)
}

func (t NSTextField) SetDrawsBackground(draws bool) {
	t.Send(sel_setDrawsBackground, draws)
}

func (t NSTextField) SetEditable(editable bool) {
	t.Send(sel_setEditable, editable)
}

func (t NSTextField) SetSelectable(selectable bool) {
	t.Send(sel_setSelectable, selectable)
}

func (t NSTextField) SetAlignment(alignment int) {
	t.Send(sel_setAlignment, alignment)
}

func (t NSTextField) SetBackgroundColor(color NSColor) {
	t.Send(sel_tfSetBackgroundColor, color.ID)
}

func (t NSTextField) SizeToFit() {
	t.Send(sel_sizeToFit)
}

func (t NSTextField) SetMaximumNumberOfLines(n int) {
	t.Send(sel_setMaximumNumberOfLines, n)
}
