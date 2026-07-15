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
	helperClass objc.Class
	helperInst  core.NSObject

	sel_createWindow = objc.RegisterName("createWindow")
	sel_showWindow   = objc.RegisterName("showWindow")
	sel_hideWindow   = objc.RegisterName("hideWindow")
	sel_closeWindow  = objc.RegisterName("closeWindow")
	sel_updateText   = objc.RegisterName("updateText")

	sel_performSelectorOnMainThread = objc.RegisterName("performSelectorOnMainThread:withObject:waitUntilDone:")

	// Shared state for the callback to access the controller
	dispatchMu   sync.Mutex
	dispatchCtrl *darwinController
)

func init() {
	cocoa.ImportCoreGraphics()

	var err error
	helperClass, err = objc.RegisterClass(
		"DesktopLyricsHelper",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{Cmd: sel_createWindow, Fn: handleCreateWindow},
			{Cmd: sel_showWindow, Fn: handleShowWindow},
			{Cmd: sel_hideWindow, Fn: handleHideWindow},
			{Cmd: sel_closeWindow, Fn: handleCloseWindow},
			{Cmd: sel_updateText, Fn: handleUpdateText},
		},
	)
	if err != nil {
		panic(err)
	}

	helperInst = core.NSObject{
		ID: objc.ID(helperClass).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
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

// ---- Dispatch functions ----

// dispatchSync dispatches a selector synchronously on the main thread.
func dispatchSync(sel objc.SEL) {
	helperInst.Send(sel_performSelectorOnMainThread, sel, objc.ID(0), true)
}

// dispatchAsync dispatches a selector asynchronously on the main thread.
func dispatchAsync(sel objc.SEL) {
	helperInst.Send(sel_performSelectorOnMainThread, sel, objc.ID(0), false)
}
