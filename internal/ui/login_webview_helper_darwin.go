//go:build darwin

package ui

import (
	"sync"

	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/webkit"
)

var (
	webviewHelperClass objc.Class
	webviewHelperInst  core.NSObject

	// webviewLoginMu guards the package-level controller reference shared
	// between the main-thread ObjC handlers and the polling goroutine.
	webviewLoginMu   sync.Mutex
	webviewLoginCtrl *webviewLoginController
)

var (
	sel_webviewCreateWindow = objc.RegisterName("createWebviewLoginWindow")
	sel_webviewCloseWindow  = objc.RegisterName("closeWebviewLoginWindow")
	sel_webviewWillClose    = objc.RegisterName("webviewLoginWindowWillClose:")
	sel_webviewSetTitle     = objc.RegisterName("setTitle:")

	sel_performSelectorOnMainThread = objc.RegisterName("performSelectorOnMainThread:withObject:waitUntilDone:")
)

func init() {
	var err error
	// The class name must be globally unique (existing: DesktopLyricsHelper,
	// AVPlayerHandler, RemoteCommandHandler, NotificationDelegate,
	// DefaultAppDelegate).
	webviewHelperClass, err = objc.RegisterClass(
		"MusicfoxWebviewLoginHelper",
		objc.GetClass("NSObject"),
		nil,
		nil,
		[]objc.MethodDef{
			{Cmd: sel_webviewCreateWindow, Fn: handleWebviewCreateWindow},
			{Cmd: sel_webviewCloseWindow, Fn: handleWebviewCloseWindow},
			{Cmd: sel_webviewWillClose, Fn: handleWebviewWillClose},
		},
	)
	if err != nil {
		panic(err)
	}
	webviewHelperInst = core.NSObject{
		ID: objc.ID(webviewHelperClass).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
	}
}

func setWebviewLoginCtrl(c *webviewLoginController) {
	webviewLoginMu.Lock()
	defer webviewLoginMu.Unlock()
	webviewLoginCtrl = c
}

func getWebviewLoginCtrl() *webviewLoginController {
	webviewLoginMu.Lock()
	defer webviewLoginMu.Unlock()
	return webviewLoginCtrl
}

// webviewDispatchSync executes the given selector on the main thread, blocking
// until it completes. It must not be called from the main thread.
func webviewDispatchSync(sel objc.SEL) {
	webviewHelperInst.Send(sel_performSelectorOnMainThread, sel, objc.ID(0), true)
}

// handleWebviewCreateWindow runs on the main thread and builds the login
// window with a WKWebView, then switches the app activation policy so the
// window can become key and receive keyboard input.
func handleWebviewCreateWindow(id objc.ID, cmd objc.SEL) {
	core.Autorelease(func() {
		ctrl := getWebviewLoginCtrl()
		if ctrl == nil {
			return
		}

		// The app runs with NSApplicationActivationPolicyProhibited globally
		// (see internal/runtime/runtime_darwin.go). Under Prohibited a window
		// can never become key, so WebView keyboard input would be dead.
		// Switch to Accessory and activate for the duration of the login.
		app := cocoa.NSApp()
		app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyAccessory)
		app.ActivateIgnoringOtherApps(true)

		dataStore := webkit.WKWebsiteDataStore_NonPersistentDataStore()
		config := webkit.WKWebViewConfiguration_alloc().Init()
		config.SetWebsiteDataStore(dataStore)

		frame := cocoa.NSRect{
			Origin: cocoa.CGPoint{X: 0, Y: 0},
			Size:   cocoa.CGSize{Width: webviewLoginWindowWidth, Height: webviewLoginWindowHeight},
		}
		win := cocoa.NSWindow_alloc().InitWithContentRectStyleMaskBackingDefer(
			frame,
			cocoa.NSWindowStyleMaskTitled|cocoa.NSWindowStyleMaskClosable|cocoa.NSWindowStyleMaskMiniaturizable|cocoa.NSWindowStyleMaskResizable,
			cocoa.NSBackingStoreBuffered,
			false,
		)
		// Keep the window object alive after the user clicks the close button
		// so the controller can still release it deterministically.
		win.SetReleasedWhenClosed(false)

		title := core.String(loginWindowTitle)
		win.Send(sel_webviewSetTitle, title.ID)
		title.Release()

		wv := webkit.WKWebView_alloc().InitWithFrameConfiguration(frame, config)
		win.SetContentView(cocoa.NSView{NSObject: wv.NSObject})
		win.Center()

		urlStr := core.String(loginURL)
		url := core.NSURL_URLWithString(urlStr)
		urlStr.Release()
		request := webkit.NSURLRequest_RequestWithURL(url)
		wv.LoadRequest(request)

		// Detect the user closing the window manually. The name string is
		// retained (alloc+init) and released together with the other ObjC
		// objects in handleWebviewCloseWindow.
		observerName := core.String("NSWindowWillCloseNotification")
		cocoa.NSNotificationCenter_defaultCenter().AddObserverSelectorNameObject(
			webviewHelperInst.ID,
			sel_webviewWillClose,
			observerName,
			win.NSObject,
		)

		ctrl.mu.Lock()
		ctrl.window = win
		// httpCookieStore is a property of WKWebsiteDataStore, not of
		// WKWebView (sending httpCookieStore to the web view raises
		// "unrecognized selector" and crashes). The data store is retained by
		// the configuration, which keeps the cookie store alive.
		ctrl.cookieStore = dataStore.HttpCookieStore()
		ctrl.cookiesBlock = ctrl.newCookiesBlock()
		ctrl.config = config
		ctrl.webView = wv
		ctrl.observerName = observerName
		ctrl.mu.Unlock()

		win.MakeKeyAndOrderFront(0)
	})
}

// handleWebviewCloseWindow runs on the main thread, closes the login window
// and restores the global Prohibited activation policy.
func handleWebviewCloseWindow(id objc.ID, cmd objc.SEL) {
	core.Autorelease(func() {
		ctrl := getWebviewLoginCtrl()
		if ctrl == nil {
			// Still restore the global Prohibited activation policy.
			cocoa.NSApp().SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
			return
		}

		ctrl.mu.Lock()
		win := ctrl.window
		config := ctrl.config
		webView := ctrl.webView
		observerName := ctrl.observerName
		blk := ctrl.cookiesBlock
		ctrl.window = cocoa.NSWindow{}
		ctrl.config = webkit.WKWebViewConfiguration{}
		ctrl.webView = webkit.WKWebView{}
		ctrl.observerName = core.NSString{}
		ctrl.cookiesBlock = 0
		ctrl.mu.Unlock()

		// Release the cookie block (+1 from NewBlock). The WebKit framework
		// keeps its own reference for any in-flight
		// getAllCookiesWithCompletionHandler: call, so this is safe even while
		// a poll is still pending; disposing the block also drops the captured
		// closure from the blockFunctionCache.
		if blk != 0 {
			blk.Release()
		}

		// Close first so the willClose notification is delivered and consumed
		// before the notification name string is released.
		if win.ID != 0 {
			win.Close()
			// Remove the notification observer while the window and the name
			// string are still valid; otherwise the notification center keeps a
			// dangling pointer to the released window and a later session's
			// window reusing the same address could fire a spurious
			// webviewLoginWindowWillClose:.
			cocoa.NSNotificationCenter_defaultCenter().RemoveObserverNameObject(
				webviewHelperInst.ID, observerName, win.NSObject,
			)
			win.Release()
		}
		if config.ID != 0 {
			config.Release()
		}
		if webView.ID != 0 {
			webView.Release()
		}
		if observerName.ID != 0 {
			observerName.Release()
		}
		// Restore the global Prohibited activation policy on every exit path.
		cocoa.NSApp().SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	})
}

// handleWebviewWillClose runs on the main thread when the user closes the
// login window. It only forwards an event to the controller; the page owns the
// cleanup.
func handleWebviewWillClose(id objc.ID, cmd objc.SEL, notification objc.ID) {
	ctrl := getWebviewLoginCtrl()
	if ctrl == nil {
		return
	}
	ctrl.sendEvent(WebviewLoginEvent{WindowClosed: true})
}
