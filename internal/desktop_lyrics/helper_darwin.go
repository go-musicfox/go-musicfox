//go:build darwin

package desktop_lyrics

import (
	"sync"

	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var (
	helperClass    objc.Class
	helperInst     core.NSObject
	dragViewClass  objc.Class

	sel_createWindow          = objc.RegisterName("createWindow")
	sel_showWindow            = objc.RegisterName("showWindow")
	sel_hideWindow            = objc.RegisterName("hideWindow")
	sel_closeWindow           = objc.RegisterName("closeWindow")
	sel_updateText            = objc.RegisterName("updateText")
	sel_scrollTick            = objc.RegisterName("scrollTick")
	sel_windowWillMove        = objc.RegisterName("windowWillMove:")
	sel_windowDidMove         = objc.RegisterName("windowDidMove:")
	sel_persistWindowPosition = objc.RegisterName("persistWindowPosition")

	// LyricsDragView mouse event selectors.
	sel_mouseDown        = objc.RegisterName("mouseDown:")
	sel_mouseDragged     = objc.RegisterName("mouseDragged:")
	sel_mouseUp          = objc.RegisterName("mouseUp:")
	sel_locationInWindow = objc.RegisterName("locationInWindow")
	sel_convertPoint     = objc.RegisterName("convertPoint:fromView:")
	sel_hitTest          = objc.RegisterName("hitTest:")

	sel_performSelectorOnMainThread = objc.RegisterName("performSelectorOnMainThread:withObject:waitUntilDone:")
	sel_performAfterDelay           = objc.RegisterName("performSelector:withObject:afterDelay:")
	sel_cancelPreviousPerform       = objc.RegisterName("cancelPreviousPerformRequestsWithTarget:selector:object:")

	// Shared state for the callback to access the controller
	dispatchMu   sync.Mutex
	dispatchCtrl *darwinController
)

func init() {
	var err error

	// ---- Helper (NSObject subclass for main-thread dispatch) ----
	helperClass, err = objc.RegisterClass(
		"DesktopLyricsHelper",
		objc.GetClass("NSObject"),
		nil, nil,
		[]objc.MethodDef{
			{Cmd: sel_createWindow, Fn: handleCreateWindow},
			{Cmd: sel_showWindow, Fn: handleShowWindow},
			{Cmd: sel_hideWindow, Fn: handleHideWindow},
			{Cmd: sel_closeWindow, Fn: handleCloseWindow},
			{Cmd: sel_updateText, Fn: handleUpdateText},
			{Cmd: sel_scrollTick, Fn: handleScrollTick},
			{Cmd: sel_windowWillMove, Fn: handleWindowWillMove},
			{Cmd: sel_windowDidMove, Fn: handleWindowDidMove},
			{Cmd: sel_persistWindowPosition, Fn: handlePersistWindowPosition},
		},
	)
	if err != nil {
		panic(err)
	}
	helperInst = core.NSObject{
		ID: objc.ID(helperClass).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
	}

	// ---- LyricsDragView (NSView subclass for mouse-based dragging) ----
	// Implements LyricsX-style drag: mouseDown records offset, mouseDragged
	// updates the screen-relative position factors, mouseUp persists to DB.
	dragViewClass, err = objc.RegisterClass(
		"LyricsDragView",
		objc.GetClass("NSView"),
		nil, nil,
		[]objc.MethodDef{
			{Cmd: sel_mouseDown, Fn: handleDragViewMouseDown},
			{Cmd: sel_mouseDragged, Fn: handleDragViewMouseDragged},
			{Cmd: sel_mouseUp, Fn: handleDragViewMouseUp},
		},
	)
	if err != nil {
		panic(err)
	}
}

func getDispatchCtrl() *darwinController {
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	return dispatchCtrl
}

func setDispatchCtrl(ctrl *darwinController) {
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	dispatchCtrl = ctrl
}

// ---- ObjC method handlers (called on main thread) ----

func handleCreateWindow(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.createWindow()
	}
}

func handleShowWindow(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.doShow()
	}
}

func handleHideWindow(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.doHide()
	}
}

func handleCloseWindow(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.doClose()
	}
}

func handleUpdateText(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.doUpdateText()
	}
}

func handleScrollTick(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.doTick()
	}
}

func handleWindowWillMove(id objc.ID, cmd objc.SEL, notification objc.ID) {
	// No-op: window is fullscreen and never moves; drag is handled by LyricsDragView.
}

func handleWindowDidMove(id objc.ID, cmd objc.SEL, notification objc.ID) {
	// No-op: window is fullscreen and never moves; drag is handled by LyricsDragView.
}

func handlePersistWindowPosition(id objc.ID, cmd objc.SEL) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.persistPositionFactors()
	}
}

// ---- LyricsDragView mouse handlers ----

func handleDragViewMouseDown(id objc.ID, cmd objc.SEL, event objc.ID) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		pt := objc.Send[cocoa.CGPoint](event, sel_locationInWindow)
		ctrl.handleDragStart(pt.X, pt.Y)
	}
}

func handleDragViewMouseDragged(id objc.ID, cmd objc.SEL, event objc.ID) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		pt := objc.Send[cocoa.CGPoint](event, sel_locationInWindow)
		ctrl.handleDragMove(pt.X, pt.Y)
	}
}

func handleDragViewMouseUp(id objc.ID, cmd objc.SEL, event objc.ID) {
	if ctrl := getDispatchCtrl(); ctrl != nil {
		ctrl.handleDragEnd()
	}
}

// ---- Dispatch functions ----

// dispatchSync dispatches a selector synchronously on the main thread.
func dispatchSync(sel objc.SEL) {
	helperInst.Send(sel_performSelectorOnMainThread, sel, objc.ID(0), true)
}

// dispatchAsync dispatches a selector asynchronously on the main thread.
func dispatchAsync(sel objc.SEL) {
	helperInst.Send(sel_performSelectorOnMainThread, sel, objc.ID(0), false)
}

// scheduleAfter schedules a one-shot selector call after the given delay in seconds.
func scheduleAfter(sel objc.SEL, delay float64) {
	helperInst.Send(sel_performAfterDelay, sel, objc.ID(0), delay)
}

// cancelScheduled cancels a previously scheduled performSelector:withObject:afterDelay:.
// This is a class method (+cancelPreviousPerformRequestsWithTarget:selector:object:)
// so we send it to the class, not the instance.
func cancelScheduled(sel objc.SEL) {
	objc.ID(helperClass).Send(sel_cancelPreviousPerform, helperInst.ID, sel, objc.ID(0))
}
