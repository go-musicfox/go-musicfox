//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSWindow = objc.GetClass("NSWindow")
}

var class_NSWindow objc.Class

var (
	sel_initWithContentRectStyleMaskBackingDefer = objc.RegisterName("initWithContentRect:styleMask:backing:defer:")
	sel_setLevel                                = objc.RegisterName("setLevel:")
	sel_setBackgroundColor                       = objc.RegisterName("setBackgroundColor:")
	sel_setAlphaValue                           = objc.RegisterName("setAlphaValue:")
	sel_setIsVisible                            = objc.RegisterName("setIsVisible:")
	sel_makeKeyAndOrderFront                    = objc.RegisterName("makeKeyAndOrderFront:")
	sel_orderOut                                = objc.RegisterName("orderOut:")
	sel_orderFront                              = objc.RegisterName("orderFront:")
	sel_setCollectionBehavior                   = objc.RegisterName("setCollectionBehavior:")
	sel_setMovableByWindowBackground            = objc.RegisterName("setMovableByWindowBackground:")
	sel_setMovable                              = objc.RegisterName("setMovable:")
	sel_setHasShadow                            = objc.RegisterName("setHasShadow:")
	sel_setOpaque                               = objc.RegisterName("setOpaque:")
	sel_setIgnoresMouseEvents                   = objc.RegisterName("setIgnoresMouseEvents:")
	sel_setContentView                          = objc.RegisterName("setContentView:")
	sel_contentView                             = objc.RegisterName("contentView")
	sel_setFrameDisplay                         = objc.RegisterName("setFrame:display:")
	sel_center                                  = objc.RegisterName("center")
	sel_close                                   = objc.RegisterName("close")
	sel_setTitlebarAppearsTransparent           = objc.RegisterName("setTitlebarAppearsTransparent:")
	sel_setReleasedWhenClosed                   = objc.RegisterName("setReleasedWhenClosed:")
	sel_setStyleMask                            = objc.RegisterName("setStyleMask:")
	sel_orderBack                               = objc.RegisterName("orderBack:")
)

// NSWindowStyleMask values
const (
	NSWindowStyleMaskBorderless     uint = 0
	NSWindowStyleMaskTitled         uint = 1 << 0
	NSWindowStyleMaskClosable       uint = 1 << 1
	NSWindowStyleMaskMiniaturizable uint = 1 << 2
	NSWindowStyleMaskResizable      uint = 1 << 3
	NSWindowStyleMaskFullSizeContentView uint = 1 << 15
)

// NSBackingStoreType
const (
	NSBackingStoreBuffered core.NSUInteger = 2
)

// NSWindowLevel
const (
	NSNormalWindowLevel     int = 0
	NSFloatingWindowLevel   int = 3
	NSScreenSaverWindowLevel int = 1000
)

// NSWindowCollectionBehavior
const (
	NSWindowCollectionBehaviorDefault                    uint = 0
	NSWindowCollectionBehaviorCanJoinAllSpaces           uint = 1 << 0
	NSWindowCollectionBehaviorStationary                 uint = 1 << 4
	NSWindowCollectionBehaviorFullScreenAuxiliary         uint = 1 << 17
	NSWindowCollectionBehaviorFullScreenNone              uint = 1 << 9
	NSWindowCollectionBehaviorFullScreenAllowsTiling      uint = 1 << 11
	NSWindowCollectionBehaviorFullScreenDisallowsTiling   uint = 1 << 12
)

type NSWindow struct {
	core.NSObject
}

func NSWindow_alloc() NSWindow {
	return NSWindow{
		core.NSObject{
			ID: objc.ID(class_NSWindow).Send(macdriver.SEL_alloc),
		},
	}
}

func (w NSWindow) InitWithContentRectStyleMaskBackingDefer(
	contentRect NSRect,
	styleMask uint,
	backing core.NSUInteger,
	deferFlag bool,
) NSWindow {
	w.ID = w.Send(sel_initWithContentRectStyleMaskBackingDefer,
		contentRect.Origin.X, contentRect.Origin.Y,
		contentRect.Size.Width, contentRect.Size.Height,
		styleMask, backing, deferFlag,
	)
	return w
}

func (w NSWindow) SetLevel(level int) {
	w.Send(sel_setLevel, level)
}

func (w NSWindow) SetBackgroundColor(color NSColor) {
	w.Send(sel_setBackgroundColor, color.ID)
}

func (w NSWindow) SetAlphaValue(alpha CGFloat) {
	w.Send(sel_setAlphaValue, alpha)
}

func (w NSWindow) SetIsVisible(visible bool) {
	w.Send(sel_setIsVisible, visible)
}

func (w NSWindow) MakeKeyAndOrderFront(sender objc.ID) {
	w.Send(sel_makeKeyAndOrderFront, sender)
}

func (w NSWindow) OrderOut(sender objc.ID) {
	w.Send(sel_orderOut, sender)
}

func (w NSWindow) OrderFront(sender objc.ID) {
	w.Send(sel_orderFront, sender)
}

func (w NSWindow) OrderBack(sender objc.ID) {
	w.Send(sel_orderBack, sender)
}

func (w NSWindow) SetCollectionBehavior(behavior uint) {
	w.Send(sel_setCollectionBehavior, behavior)
}

func (w NSWindow) SetMovableByWindowBackground(movable bool) {
	w.Send(sel_setMovableByWindowBackground, movable)
}

func (w NSWindow) SetMovable(movable bool) {
	w.Send(sel_setMovable, movable)
}

func (w NSWindow) SetHasShadow(hasShadow bool) {
	w.Send(sel_setHasShadow, hasShadow)
}

func (w NSWindow) SetOpaque(opaque bool) {
	w.Send(sel_setOpaque, opaque)
}

func (w NSWindow) SetIgnoresMouseEvents(ignores bool) {
	w.Send(sel_setIgnoresMouseEvents, ignores)
}

func (w NSWindow) SetContentView(view NSView) {
	w.Send(sel_setContentView, view.ID)
}

func (w NSWindow) ContentView() NSView {
	return NSView{
		core.NSObject{
			ID: w.Send(sel_contentView),
		},
	}
}

func (w NSWindow) Frame() NSRect {
	return objc.Send[NSRect](w.ID, sel_frame)
}

func (w NSWindow) SetFrameDisplayTopLeft(frameRect NSRect, display bool) {
	w.Send(sel_setFrameDisplay,
		frameRect.Origin.X, frameRect.Origin.Y,
		frameRect.Size.Width, frameRect.Size.Height,
		display,
	)
}

func (w NSWindow) Center() {
	w.Send(sel_center)
}

func (w NSWindow) Close() {
	w.Send(sel_close)
}

func (w NSWindow) SetTitlebarAppearsTransparent(transparent bool) {
	w.Send(sel_setTitlebarAppearsTransparent, transparent)
}

func (w NSWindow) SetReleasedWhenClosed(released bool) {
	w.Send(sel_setReleasedWhenClosed, released)
}

func (w NSWindow) SetStyleMask(mask uint) {
	w.Send(sel_setStyleMask, mask)
}
