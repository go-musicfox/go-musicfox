//go:build linux

package ui

import (
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/go-musicfox/go-musicfox/internal/webkitgtk"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

// webviewLoginAvailable reports whether the native WebKitGTK login window is
// available on the current platform (Linux only). A supported WebKitGTK
// stack must be loadable; WebKitGTKVersion() returns 0 when none is found.
func webviewLoginAvailable() bool {
	return webkitgtk.WebKitGTKVersion() != 0
}

func usesLegacyWebKitAPI(version int) bool {
	return version == webkitgtk.Version41 || version == webkitgtk.Version40
}

// webviewLoginController opens a native WebKitGTK login window and polls the
// WebKitCookieManager until the MUSIC_U cookie appears.
//
// Threading model: GTK/WebKit objects may only be touched from the thread that
// runs the GTK main loop. Open() starts a dedicated goroutine (runGTK) locked
// to an OS thread that initializes GTK, builds the window and blocks in
// gtk_main. Cookie polling is a g_timeout_add source running on that same
// thread; the asynchronous CookieManager completion callback also runs there.
// The controller only forwards events through the buffered channel, which is
// safe from any goroutine. Close() only calls gtk_main_quit() (safe from any
// thread); the GTK thread then destroys the window during its cleanup.
type webviewLoginController struct {
	mu sync.Mutex

	events chan WebviewLoginEvent

	session       uintptr // WebKitNetworkSession (ephemeral)
	cookieManager uintptr // WebKitCookieManager
	window        uintptr // GtkWindow
	webView       uintptr // WebKitWebView

	pollTimerID uint // g_timeout_add 返回值
	inFlight    bool
	opened      bool
	closed      bool
	stopPolling chan struct{}

	// windowDestroyed records that the user's close-request already destroyed
	// the window (returning FALSE), so runGTK must not close it again after
	// gtk_main returns.
	windowDestroyed bool

	// Callback function pointers, kept alive for the whole controller
	// lifetime: GTK holds them while the main loop is running.
	timeoutCB      uintptr
	cookiesReadyCB uintptr
	closeCB        uintptr
}

func newWebviewLoginController() *webviewLoginController {
	return &webviewLoginController{events: make(chan WebviewLoginEvent, 4)}
}

// sendEvent forwards an event to the page. Never blocks; the channel is never
// closed.
func (c *webviewLoginController) sendEvent(ev WebviewLoginEvent) {
	select {
	case c.events <- ev:
	default:
	}
}

// PopEvent returns the next pending event, if any. Non-blocking.
func (c *webviewLoginController) PopEvent() (WebviewLoginEvent, bool) {
	select {
	case ev := <-c.events:
		return ev, true
	default:
		return WebviewLoginEvent{}, false
	}
}

// Open starts the GTK login window on a dedicated goroutine. Without a
// DISPLAY/WAYLAND_DISPLAY there is no way to show a GTK window, so it reports
// WindowClosed immediately and the page falls back to the form login.
func (c *webviewLoginController) Open() {
	c.mu.Lock()
	if c.closed || c.opened {
		c.mu.Unlock()
		return
	}
	c.opened = true
	c.stopPolling = make(chan struct{})
	c.mu.Unlock()

	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		slog.Error("WebView 登录不可用: 未检测到 DISPLAY/WAYLAND_DISPLAY 环境变量")
		c.mu.Lock()
		if c.stopPolling != nil {
			close(c.stopPolling)
			c.stopPolling = nil
		}
		c.mu.Unlock()
		c.sendEvent(WebviewLoginEvent{WindowClosed: true, ErrMsg: "未检测到 DISPLAY/WAYLAND_DISPLAY 环境变量，无法打开登录窗口"})
		return
	}

	errorx.Go(c.runGTK)
}

// Close stops the GTK main loop so the GTK goroutine tears the window down on
// its own thread. Safe to call multiple times; every exit path (success, user
// close, cancel) goes through here.
func (c *webviewLoginController) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	if c.stopPolling != nil {
		close(c.stopPolling)
		c.stopPolling = nil
	}
	c.mu.Unlock()

	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return
	}
	webkitgtk.GtkMainQuit()
}

// runGTK owns the GTK main loop thread. It initializes GTK, builds the login
// window, registers the polling timer and blocks in gtk_main until Close() or
// the user's close-request quits the loop, then destroys the window.
func (c *webviewLoginController) runGTK() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Neither WebKitGTK 6.0, 4.1 nor 4.0 is available: degrade to the form login.
	if webkitgtk.WebKitGTKVersion() == 0 {
		slog.Error("WebView 登录不可用: 未找到 WebKitGTK 6.0/4.1/4.0 运行库")
		c.sendEvent(WebviewLoginEvent{WindowClosed: true, ErrMsg: "未找到 WebKitGTK 运行库（请安装 webkit2gtk）"})
		return
	}

	// gtk_init aborts the whole process on a stale DISPLAY/WAYLAND_DISPLAY;
	// gtk_init_check reports the failure instead. At this point no GTK object
	// has been created yet, so there is nothing to clean up.
	if !webkitgtk.GtkInitCheck() {
		slog.Error("WebView 登录不可用: GTK 初始化失败")
		c.sendEvent(WebviewLoginEvent{WindowClosed: true, ErrMsg: "GTK 初始化失败，无法打开登录窗口"})
		return
	}

	var (
		session       uintptr
		cookieManager uintptr
		webView       uintptr
	)
	// webkit_web_view_new() 返回 floating reference（WebKitWebView 派生自
	// GInitiallyUnowned）。必须立即 g_object_ref_sink 接管所有权：GtkWindowAddChild
	// 对非 floating 对象只增计数，于是引用计数变为「我们 1 + 窗口 1」；defer unref
	// 归还我们的 1，窗口销毁时再归零。若不对 floating 引用做 sink，AddChild 会把它
	// 吞掉（窗口成为唯一持有者），窗口销毁时 webview 即被释放，随后的 defer unref
	// 将作用在已释放内存上（use-after-free）。
	if usesLegacyWebKitAPI(webkitgtk.WebKitGTKVersion()) {
		// WebKitGTK 4.1 and 4.0 use the WebContext/CookieManager API; 4.1
		// differs from 4.0 only in its libsoup ABI.
		ctx := webkitgtk.WebContextGetDefault()
		cookieManager = webkitgtk.WebContextGetCookieManager(ctx)
		webView = webkitgtk.WebViewNew()
		webkitgtk.GObjectRefSink(webView)
	} else {
		session = webkitgtk.NetworkSessionNewEphemeral()
		// Release our application-side references (transfer-full from the
		// constructors) on every exit path. Deferred cleanup runs LIFO, so the
		// window is unref'd before the web view, which is unref'd before the
		// session: the window holds a reference to the web view until it is
		// destroyed. The cookie manager is a borrowed reference of the session
		// and must not be unref'd separately.
		defer webkitgtk.GObjectUnref(session)
		cookieManager = webkitgtk.NetworkSessionGetCookieManager(session)
		webView = webkitgtk.WebViewNewWithNetworkSession(session)
		webkitgtk.GObjectRefSink(webView)
	}
	defer webkitgtk.GObjectUnref(webView)
	window := webkitgtk.GtkWindowNew()
	defer webkitgtk.GObjectUnref(window)

	// Publish the GTK objects and the callback pointers (the closures capture
	// c) before entering the main loop.
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		// Close() raced with startup: never show the window.
		webkitgtk.GtkWindowClose(window)
		return
	}
	c.session = session
	c.cookieManager = cookieManager
	c.webView = webView
	c.window = window
	c.timeoutCB = purego.NewCallback(c.pollTick)
	c.cookiesReadyCB = purego.NewCallback(c.cookiesReady)
	c.closeCB = purego.NewCallback(c.handleCloseRequest)
	c.mu.Unlock()

	webkitgtk.GtkWindowSetTitle(window, loginWindowTitle)
	webkitgtk.GtkWindowSetDefaultSize(window, int32(webviewLoginWindowWidth), int32(webviewLoginWindowHeight))
	webkitgtk.GtkWindowAddChild(window, webView)
	webkitgtk.GtkWindowPresent(window)
	// GTK3 widgets are not shown automatically (unlike GTK4); show the web
	// view explicitly. This is a no-op on GTK4.
	webkitgtk.GtkShowWidget(webView)
	webkitgtk.WebViewLoadURI(webView, loginURL)

	// User closes the window manually: "close-request" (GTK4) or
	// "delete-event" (GTK3); returning FALSE lets GTK destroy the window.
	webkitgtk.GSignalConnectCloseRequest(window, c.closeCB)

	// Poll the cookie manager once per second from the GTK main loop thread.
	c.pollTimerID = webkitgtk.GTimeoutAdd(uint32(webviewLoginPollInterval/time.Millisecond), c.timeoutCB, 0)

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		// Close() raced with startup while the window was being set up.
		webkitgtk.GtkWindowClose(window)
		return
	}

	webkitgtk.GtkMain()

	c.mu.Lock()
	destroyed := c.windowDestroyed
	c.mu.Unlock()
	if !destroyed {
		// Programmatic close path: gtk_main_quit() only stopped the loop, the
		// window would linger on screen. Destroy it on the GTK thread.
		webkitgtk.GtkWindowClose(window)
	}
}

// pollTick is the GSourceFunc registered with g_timeout_add. It runs on the
// GTK main loop thread and starts one asynchronous cookie fetch at a time.
// Returning 1 keeps the timer running.
func (c *webviewLoginController) pollTick(_ uintptr) int32 {
	c.mu.Lock()
	if c.closed || c.inFlight {
		c.mu.Unlock()
		return 1
	}
	manager := c.cookieManager
	cb := c.cookiesReadyCB
	if manager == 0 || cb == 0 {
		c.mu.Unlock()
		return 1
	}
	c.inFlight = true
	c.mu.Unlock()

	if usesLegacyWebKitAPI(webkitgtk.WebKitGTKVersion()) {
		webkitgtk.CookieManagerGetCookies(manager, loginURL, 0, cb, 0)
	} else {
		webkitgtk.CookieManagerGetAllCookies(manager, 0, cb, 0)
	}
	return 1
}

// cookiesReady is the GAsyncReadyCallback of the active WebKitCookieManager
// API. It runs on the GTK main loop thread, marshals the cookies into a cookie
// string and forwards an event once MUSIC_U is found.
func (c *webviewLoginController) cookiesReady(_ uintptr, res uintptr, _ uintptr) {
	c.mu.Lock()
	manager := c.cookieManager
	c.inFlight = false
	c.mu.Unlock()

	if manager == 0 || res == 0 {
		return
	}

	// err is a GError** slot; the caller ignores the error, but the GError
	// itself must be freed.
	var errPtr uintptr
	var list uintptr
	if usesLegacyWebKitAPI(webkitgtk.WebKitGTKVersion()) {
		list = webkitgtk.CookieManagerGetCookiesFinish(manager, res, uintptr(unsafe.Pointer(&errPtr)))
	} else {
		list = webkitgtk.CookieManagerGetAllCookiesFinish(manager, res, uintptr(unsafe.Pointer(&errPtr)))
	}
	if errPtr != 0 {
		webkitgtk.GErrorFree(errPtr)
	}
	var (
		pairs []string
		found bool
	)
	length := webkitgtk.GListLength(list)
	for i := uint(0); i < length; i++ {
		cookie := webkitgtk.GListNthData(list, i)
		if cookie == 0 {
			continue
		}
		name := webkitgtk.SoupCookieGetName(cookie)
		value := webkitgtk.SoupCookieGetValue(cookie)
		pairs = append(pairs, name+"="+value)
		if name == musicUCookieName {
			found = true
		}
	}

	// The GList and its SoupCookie elements are owned by the caller
	// (equivalent to g_list_free_full(list, soup_cookie_free)); release them
	// after all reads, once the name/value strings have been copied into Go
	// strings.
	for i := uint(0); i < length; i++ {
		if cookie := webkitgtk.GListNthData(list, i); cookie != 0 {
			webkitgtk.SoupCookieFree(cookie)
		}
	}
	webkitgtk.GListFree(list)

	if !found {
		return
	}
	c.sendEvent(WebviewLoginEvent{CookieString: strings.Join(pairs, "; ")})
}

// handleCloseRequest is the "close-request" (GTK4) / "delete-event" (GTK3)
// signal handler. It runs on the GTK main loop thread. The second callback
// argument is the user_data on GTK4 and the GdkEvent pointer on GTK3; it is
// ignored here. Returning FALSE lets GTK destroy the window.
func (c *webviewLoginController) handleCloseRequest(_ uintptr, _ uintptr) int32 {
	c.mu.Lock()
	alreadyClosed := c.closed
	c.windowDestroyed = true
	c.mu.Unlock()

	if !alreadyClosed {
		c.sendEvent(WebviewLoginEvent{WindowClosed: true})
	}
	webkitgtk.GtkMainQuit()
	return 0
}
