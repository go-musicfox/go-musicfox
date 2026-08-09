//go:build windows

package ui

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/13thgoutham/go-webview2/pkg/webview2"
	"github.com/13thgoutham/go-webview2/webviewloader"

	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// webviewLoginAvailable reports whether the native WebView2 login window is
// available on the current platform (Windows only).
const webviewLoginAvailable = true

// webviewLoginController opens a native WebView2 login window and polls the
// WebView2 cookie manager until the MUSIC_U cookie appears.
//
// Threading model: everything runs on a single UI thread. The messageLoop
// goroutine (pinned with runtime.LockOSThread) creates the window, starts the
// WebView2 environment, pumps the window messages and drives the cookie
// polling through a WM_TIMER. This single-thread layout is mandatory: the
// WebView2 completion callbacks (environment, controller and cookie list) are
// posted to a hidden message window owned by the thread that initiated the
// corresponding call, and are only dispatched while that thread pumps
// messages. Spawning the WebView2 init or the cookie polling on separate
// goroutines without a message pump would leave the callbacks undelivered.
type webviewLoginController struct {
	mu sync.Mutex

	events chan WebviewLoginEvent

	hwnd uintptr

	// WebView2 COM objects obtained during initialization. They are created on
	// the UI thread and released in cleanup() after the message loop exits.
	controller    *webview2.ICoreWebView2Controller
	core          *webview2.ICoreWebView2
	cookieManager *webview2.ICoreWebView2CookieManager

	// Handler objects handed to native WebView2 calls. They are stored here
	// for the whole controller lifetime so the Go GC cannot collect them
	// before the asynchronous COM callbacks fire.
	envHandler        *webview2EnvHandler
	controllerHandler *webview2.ICoreWebView2CreateCoreWebView2ControllerCompletedHandler
	cookiesHandler    *webview2.ICoreWebView2GetCookiesCompletedHandler

	// inFlight serializes cookie polls. It is read and written only on the UI
	// thread (WM_TIMER dispatch and the GetCookies completion callback), so it
	// needs no mutex of its own; c.mu still guards closed/controller and the
	// other fields that Close() may read from an arbitrary goroutine.
	inFlight bool

	opened bool
	closed bool
}

// Win32 window class and message constants.
const (
	webviewLoginWndClass = "GoMusicfoxWebviewLoginWnd"

	wmTimer         = 0x0113
	wmClose         = 0x0010
	wmDestroy       = 0x0002
	wmGetMinMaxInfo = 0x0024

	// cookiePollTimerID identifies the window timer that drives the cookie
	// polling on the UI thread. It must match the wParam of WM_TIMER messages.
	cookiePollTimerID = uintptr(1)

	wsOverlappedWindow = 0x00CF0000 // WS_OVERLAPPEDWINDOW

	cwUsedefault = 0x80000000 // CW_USEDEFAULT
	swShow       = 5          // SW_SHOW

	// LoadImageW arguments for the shared arrow cursor (a NULL class cursor
	// makes the mouse pointer invisible over the window).
	idcArrow    = 32512      // IDC_ARROW
	imageCursor = 2          // IMAGE_CURSOR
	lrShared    = 0x00008000 // LR_SHARED

	// Minimum login window size (WM_GETMINMAXINFO).
	minWebviewLoginWindowWidth  = 400
	minWebviewLoginWindowHeight = 300

	// GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT: query the module handle
	// without incrementing its reference count.
	getModuleHandleExFlagUnchangedRefcount = 0x00000002
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	ole32  = windows.NewLazySystemDLL("ole32.dll")

	user32RegisterClassExW = user32.NewProc("RegisterClassExW")
	user32CreateWindowExW  = user32.NewProc("CreateWindowExW")
	user32DestroyWindow    = user32.NewProc("DestroyWindow")
	user32ShowWindow       = user32.NewProc("ShowWindow")
	user32UpdateWindow     = user32.NewProc("UpdateWindow")
	user32SetFocus         = user32.NewProc("SetFocus")
	user32GetMessageW      = user32.NewProc("GetMessageW")
	user32TranslateMessage = user32.NewProc("TranslateMessage")
	user32DispatchMessageW = user32.NewProc("DispatchMessageW")
	user32DefWindowProcW   = user32.NewProc("DefWindowProcW")
	user32GetClientRect    = user32.NewProc("GetClientRect")
	user32PostQuitMessage  = user32.NewProc("PostQuitMessage")
	user32PostMessageW     = user32.NewProc("PostMessageW")
	user32SetTimer         = user32.NewProc("SetTimer")
	user32KillTimer        = user32.NewProc("KillTimer")
	user32LoadImageW       = user32.NewProc("LoadImageW")
	ole32CoInitializeEx    = ole32.NewProc("CoInitializeEx")
	ole32CoUninitialize    = ole32.NewProc("CoUninitialize")
)

// webviewLoginWndProcCallback is the window procedure for the login window. It
// is a package-level callback so the window class can be re-registered across
// login sessions.
var webviewLoginWndProcCallback = windows.NewCallback(webviewLoginWndProc)

// webviewLoginWindowContext maps an HWND to its owning controller so the
// window procedure can forward close events without per-window storage.
var (
	webviewLoginWindowContext     = map[uintptr]*webviewLoginController{}
	webviewLoginWindowContextSync sync.RWMutex
)

func setWebviewLoginWindowContext(hwnd uintptr, c *webviewLoginController) {
	webviewLoginWindowContextSync.Lock()
	webviewLoginWindowContext[hwnd] = c
	webviewLoginWindowContextSync.Unlock()
}

func getWebviewLoginWindowContext(hwnd uintptr) *webviewLoginController {
	webviewLoginWindowContextSync.RLock()
	defer webviewLoginWindowContextSync.RUnlock()
	return webviewLoginWindowContext[hwnd]
}

func deleteWebviewLoginWindowContext(hwnd uintptr) {
	webviewLoginWindowContextSync.Lock()
	delete(webviewLoginWindowContext, hwnd)
	webviewLoginWindowContextSync.Unlock()
}

// Win32 struct mirrors (the golang.org/x/sys/windows package does not expose
// these window structs).

type webviewLoginW32Point struct{ x, y int32 }

type webviewLoginW32Rect struct{ left, top, right, bottom int32 }

// webviewLoginW32Msg mirrors the Win32 MSG struct (48 bytes on 64-bit).
type webviewLoginW32Msg struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       webviewLoginW32Point
	lPrivate uint32
}

// webviewLoginW32MinMaxInfo mirrors the Win32 MINMAXINFO struct.
type webviewLoginW32MinMaxInfo struct {
	ptReserved     webviewLoginW32Point
	ptMaxSize      webviewLoginW32Point
	ptMaxPosition  webviewLoginW32Point
	ptMinTrackSize webviewLoginW32Point
	ptMaxTrackSize webviewLoginW32Point
}

// webviewLoginW32WndClassEx mirrors the Win32 WNDCLASSEXW struct.
type webviewLoginW32WndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       windows.Handle
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

// Open creates the login window and starts the single messageLoop goroutine,
// which owns the window, the WebView2 runtime and the cookie polling.
func (c *webviewLoginController) Open() {
	c.mu.Lock()
	if c.closed || c.opened {
		c.mu.Unlock()
		return
	}
	c.opened = true
	c.mu.Unlock()

	errorx.Go(c.messageLoop)
}

// Close closes the login window. Safe to call multiple times and from any
// goroutine; every exit path (success, user close, cancel) goes through here.
// It posts WM_CLOSE to the login window; the window procedure tears the window
// down and the message loop performs the COM cleanup afterwards.
func (c *webviewLoginController) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	hwnd := c.hwnd
	c.mu.Unlock()

	if hwnd != 0 {
		// user32 PostMessage is thread-safe and may be called from any thread.
		user32PostMessageW.Call(hwnd, wmClose, 0, 0)
	}
}

// messageLoop is the single UI thread of the login window. It is pinned to one
// OS thread with LockOSThread so the window and the WebView2 hidden message
// windows keep receiving their messages: GetMessageW only drains the queue of
// the current OS thread, and a goroutine that migrates between threads would
// strand the window messages on the old thread.
func (c *webviewLoginController) messageLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// COM must be initialized on the thread that creates the WebView2
	// environment (the same thread that pumps its callbacks).
	if r, _, _ := ole32CoInitializeEx.Call(0, 2); int32(r) < 0 {
		slog.Warn("CoInitializeEx 调用失败", slog.Any("hr", r))
	}
	defer ole32CoUninitialize.Call()

	hwnd := c.createWindow()

	c.mu.Lock()
	c.hwnd = hwnd
	closed := c.closed
	c.mu.Unlock()

	if hwnd == 0 {
		// Window creation failed; nothing to pump or destroy.
		if !closed {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		return
	}
	if closed {
		// Close() ran while the window was being created.
		user32DestroyWindow.Call(hwnd)
		return
	}

	user32ShowWindow.Call(hwnd, swShow)
	user32UpdateWindow.Call(hwnd)
	user32SetFocus.Call(hwnd)

	if err := c.setupWebView2(hwnd); err != nil {
		c.mu.Lock()
		closed = c.closed
		c.mu.Unlock()
		if !closed {
			slog.Error("初始化 WebView2 登录窗口失败", slogx.Error(err))
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		user32DestroyWindow.Call(hwnd)
		return
	}

	// Message pump. The WebView2 environment/controller/cookie completion
	// callbacks are all delivered here through the hidden message windows of
	// this thread.
	var msg webviewLoginW32Msg
	for {
		ret, _, _ := user32GetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		// 0 means WM_QUIT; 0xFFFFFFFF means GetMessage failed.
		if ret == 0 || ret == 0xFFFFFFFF {
			break
		}
		user32TranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		user32DispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	c.cleanup()
}

// createWindow registers the window class and creates the (initially hidden)
// login window. It returns the HWND, or 0 on failure.
func (c *webviewLoginController) createWindow() uintptr {
	var hinstance windows.Handle
	if err := windows.GetModuleHandleEx(getModuleHandleExFlagUnchangedRefcount, nil, &hinstance); err != nil {
		slog.Error("获取模块句柄失败", slogx.Error(err))
		return 0
	}

	className, err := windows.UTF16PtrFromString(webviewLoginWndClass)
	if err != nil {
		slog.Error("转换窗口类名失败", slogx.Error(err))
		return 0
	}
	windowName, err := windows.UTF16PtrFromString(loginWindowTitle)
	if err != nil {
		slog.Error("转换窗口标题失败", slogx.Error(err))
		return 0
	}

	// Load the shared arrow cursor; a NULL class cursor makes the mouse
	// pointer invisible over the window.
	cursor, _, _ := user32LoadImageW.Call(0, idcArrow, imageCursor, 0, 0, lrShared)

	wc := webviewLoginW32WndClassEx{
		cbSize:        uint32(unsafe.Sizeof(webviewLoginW32WndClassEx{})),
		lpfnWndProc:   webviewLoginWndProcCallback,
		hInstance:     hinstance,
		hCursor:       windows.Handle(cursor),
		lpszClassName: className,
	}
	// The class survives in the process after the window is destroyed, so a
	// repeated RegisterClassExW during a later login session fails with
	// ERROR_CLASS_ALREADY_EXISTS — expected and safe to ignore.
	user32RegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := user32CreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		wsOverlappedWindow,
		cwUsedefault, cwUsedefault,
		uintptr(webviewLoginWindowWidth), uintptr(webviewLoginWindowHeight),
		0, 0,
		uintptr(hinstance),
		0,
	)
	if hwnd == 0 {
		slog.Error("创建 WebView2 登录窗口失败")
		return 0
	}
	setWebviewLoginWindowContext(hwnd, c)
	return hwnd
}

// setupWebView2 prepares the WebView2 completion handlers and starts the
// environment creation. CreateCoreWebView2EnvironmentWithOptions is
// non-blocking: it returns immediately and the environment completion callback
// is posted to a hidden message window of this thread, to be dispatched by the
// message pump in messageLoop. The user data folder is kept at
// os.UserCacheDir()/go-musicfox/webview2 (shared with the previous
// implementation; a dedicated per-user subfolder was considered but the
// difference is accepted to avoid churn).
func (c *webviewLoginController) setupWebView2(hwnd uintptr) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errWebviewLoginClosed
	}
	c.envHandler = &webview2EnvHandler{ctrl: c, hwnd: hwnd}
	c.controllerHandler = webview2.NewICoreWebView2CreateCoreWebView2ControllerCompletedHandler(&webview2ControllerHandler{ctrl: c, hwnd: hwnd})
	c.cookiesHandler = webview2.NewICoreWebView2GetCookiesCompletedHandler(&webview2CookiesHandler{ctrl: c})
	c.mu.Unlock()

	userDataDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	userDataDir = filepath.Join(userDataDir, "go-musicfox", "webview2")
	if err := os.MkdirAll(userDataDir, 0o755); err != nil {
		return err
	}

	return webviewloader.CreateCoreWebView2EnvironmentWithOptions(
		c.envHandler,
		webviewloader.WithUserDataFolder(userDataDir),
	)
}

// pollOnce starts one cookie poll if no other poll is in flight. It runs on
// the UI thread (WM_TIMER dispatch); the GetCookies completion handler is
// delivered back to the same thread through the message pump.
func (c *webviewLoginController) pollOnce() {
	c.mu.Lock()
	if c.closed || c.cookieManager == nil {
		c.mu.Unlock()
		return
	}
	cookieManager := c.cookieManager
	c.mu.Unlock()

	if c.inFlight {
		return
	}
	// Mark the poll as in flight before the async call so a (contract-violating)
	// synchronous completion cannot strand the flag; GetCookiesCompleted resets
	// it. On a synchronous error keep it clear so the next tick retries.
	c.inFlight = true
	if err := cookieManager.GetCookies("", c.cookiesHandler); err != nil {
		c.inFlight = false
		return
	}
}

// handleCookies runs on the UI thread after each GetCookies call. It reads the
// cookie list, releases the transient cookie objects and the list itself, and
// forwards a successful MUSIC_U detection through the event channel.
func (c *webviewLoginController) handleCookies(result *webview2.ICoreWebView2CookieList) {
	if result == nil {
		return
	}
	count, err := result.GetCount()
	if err != nil {
		releaseWebview2Com(unsafe.Pointer(result), &result.Vtbl.IUnknownVtbl)
		return
	}
	var (
		pairs []string
		found bool
	)
	for i := uint32(0); i < count; i++ {
		cookie, err := result.GetValueAtIndex(i)
		if err != nil || cookie == nil {
			continue
		}
		name, err := cookie.GetName()
		if err != nil {
			releaseWebview2Com(unsafe.Pointer(cookie), &cookie.Vtbl.IUnknownVtbl)
			continue
		}
		value, err := cookie.GetValue()
		if err != nil {
			releaseWebview2Com(unsafe.Pointer(cookie), &cookie.Vtbl.IUnknownVtbl)
			continue
		}
		pairs = append(pairs, name+"="+value)
		if name == musicUCookieName {
			found = true
		}
		// Each cookie returned by GetValueAtIndex carries its own reference.
		releaseWebview2Com(unsafe.Pointer(cookie), &cookie.Vtbl.IUnknownVtbl)
	}
	// The caller transfers ownership of the cookie list to the handler.
	releaseWebview2Com(unsafe.Pointer(result), &result.Vtbl.IUnknownVtbl)

	if !found {
		return
	}
	c.sendEvent(WebviewLoginEvent{CookieString: strings.Join(pairs, "; ")})
}

// releaseWebview2Com releases a WebView2 COM interface through the Release slot
// of its IUnknown vtable. The wailsapp bindings only generate AddRef() for the
// WebView2 interface types, so the vtable entry is called directly; every
// generated vtbl embeds IUnknownVtbl as its first field, making the slot
// layout identical across all interfaces.
func releaseWebview2Com(this unsafe.Pointer, vtbl *webview2.IUnknownVtbl) {
	if this == nil || vtbl == nil {
		return
	}
	vtbl.CallRelease(this)
}

// cleanup runs after the message loop exits and tears down the window and the
// WebView2 COM objects. It is idempotent: normal closes already destroyed the
// window (WM_DESTROY) and every COM field is nil-ed out here, so a repeated
// call is a no-op.
func (c *webviewLoginController) cleanup() {
	c.mu.Lock()
	hwnd := c.hwnd
	c.hwnd = 0
	controller := c.controller
	core := c.core
	cookieManager := c.cookieManager
	c.controller = nil
	c.core = nil
	c.cookieManager = nil
	c.mu.Unlock()

	if hwnd != 0 && getWebviewLoginWindowContext(hwnd) != nil {
		// The window survived the loop (abnormal exit, e.g. GetMessage
		// failure). Notify the page only if the close was not requested, then
		// destroy the window; WM_DESTROY clears the context entry.
		c.mu.Lock()
		notify := !c.closed
		c.closed = true
		c.mu.Unlock()
		user32DestroyWindow.Call(hwnd)
		deleteWebviewLoginWindowContext(hwnd)
		if notify {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
	}

	if controller != nil {
		// Close() notifies the browser process that the controller is going
		// away; we are still on the UI thread here.
		_ = controller.Close()
		releaseWebview2Com(unsafe.Pointer(controller), &controller.Vtbl.IUnknownVtbl)
	}
	if core != nil {
		releaseWebview2Com(unsafe.Pointer(core), &core.Vtbl.IUnknownVtbl)
	}
	if cookieManager != nil {
		releaseWebview2Com(unsafe.Pointer(cookieManager), &cookieManager.Vtbl.IUnknownVtbl)
	}
}

// errWebviewLoginClosed reports that the login window was closed while the
// WebView2 environment was being set up.
var errWebviewLoginClosed = errors.New("webview login closed")

// webviewLoginWndProc is the window procedure for the login window.
func webviewLoginWndProc(hwnd, msg, wp, lp uintptr) uintptr {
	switch msg {
	case wmTimer:
		if wp != cookiePollTimerID {
			ret, _, _ := user32DefWindowProcW.Call(hwnd, msg, wp, lp)
			return ret
		}
		if c := getWebviewLoginWindowContext(hwnd); c != nil {
			c.pollOnce()
		}
		return 0
	case wmClose:
		if c := getWebviewLoginWindowContext(hwnd); c != nil {
			c.mu.Lock()
			notify := !c.closed
			c.closed = true
			c.mu.Unlock()
			if notify {
				// The window was closed by the user (Close() posts WM_CLOSE
				// only after setting closed, so this path is user-initiated).
				c.sendEvent(WebviewLoginEvent{WindowClosed: true})
			}
		}
		user32DestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		deleteWebviewLoginWindowContext(hwnd)
		user32KillTimer.Call(hwnd, cookiePollTimerID)
		user32PostQuitMessage.Call(0)
		return 0
	case wmGetMinMaxInfo:
		mmi := (*webviewLoginW32MinMaxInfo)(unsafe.Pointer(lp))
		mmi.ptMinTrackSize = webviewLoginW32Point{x: minWebviewLoginWindowWidth, y: minWebviewLoginWindowHeight}
		return 0
	default:
		ret, _, _ := user32DefWindowProcW.Call(hwnd, msg, wp, lp)
		return ret
	}
}

// webview2EnvHandler receives the created WebView2 environment. It is wrapped
// by webviewloader's combridge machinery, so it only implements the business
// method (IUnknown is provided by combridge). The environment pointer itself is
// owned by the webviewloader wrapper, which releases it right after this
// callback returns, so it is never stored.
type webview2EnvHandler struct {
	ctrl *webviewLoginController
	hwnd uintptr
}

func (h *webview2EnvHandler) EnvironmentCompleted(errorCode webviewloader.HRESULT, createdEnvironment *webviewloader.ICoreWebView2Environment) webviewloader.HRESULT {
	c := h.ctrl
	if errorCode < 0 || createdEnvironment == nil {
		slog.Error("WebView2 环境创建失败", slog.Any("errorCode", errorCode))
		if !c.closed {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		return 0
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0
	}
	handler := c.controllerHandler
	c.mu.Unlock()

	// combridge.IUnknownImpl and webview2.ICoreWebView2Environment share the
	// same single-pointer-to-vtable layout, so a plain pointer cast is valid.
	env := (*webview2.ICoreWebView2Environment)(unsafe.Pointer(createdEnvironment))
	if err := env.CreateCoreWebView2Controller(webview2.HWND(h.hwnd), handler); err != nil {
		slog.Error("创建 WebView2 登录窗口失败", slogx.Error(err))
		c.sendEvent(WebviewLoginEvent{WindowClosed: true})
	}
	return 0
}

// webview2ControllerHandler receives the created WebView2 controller, resolves
// the cookie manager, navigates to the login page and arms the polling timer.
type webview2ControllerHandler struct {
	ctrl *webviewLoginController
	hwnd uintptr
}

func (h *webview2ControllerHandler) QueryInterface(_, _ uintptr) uintptr { return 0 }
func (h *webview2ControllerHandler) AddRef() uintptr                     { return 1 }
func (h *webview2ControllerHandler) Release() uintptr                    { return 1 }

func (h *webview2ControllerHandler) CreateCoreWebView2ControllerCompleted(errorCode uintptr, result *webview2.ICoreWebView2Controller) uintptr {
	c := h.ctrl
	if errorCode != 0 || result == nil {
		slog.Error("WebView2 登录窗口创建回调失败", slog.Any("errorCode", errorCode))
		if !c.closed {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		return 0
	}

	c.mu.Lock()
	if c.closed {
		// The window was closed before the controller finished creating. The
		// completed handler passes a borrowed reference owned by WebView2,
		// which releases it after the callback returns; do not Release here.
		c.mu.Unlock()
		return 0
	}
	// The completed handler receives a borrowed reference: WebView2 releases
	// its own reference right after the callback returns, which would destroy
	// the controller while cleanup() still holds the pointer. AddRef before
	// storing it (mirrors wails' chromium.go).
	result.AddRef()
	c.controller = result
	c.mu.Unlock()

	// Fill the parent window's client area.
	var rc webviewLoginW32Rect
	if ret, _, _ := user32GetClientRect.Call(h.hwnd, uintptr(unsafe.Pointer(&rc))); ret != 0 {
		_ = result.PutBounds(webview2.RECT{Left: 0, Top: 0, Right: rc.right, Bottom: rc.bottom})
	}

	core, err := result.GetCoreWebView2()
	if err != nil || core == nil {
		slog.Error("获取 WebView2 core 失败", slogx.Error(err))
		if !c.closed {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		return 0
	}

	core2 := core.GetICoreWebView2_2()
	if core2 == nil {
		slog.Error("获取 ICoreWebView2_2 接口失败")
		// core is a borrowed reference owned by WebView2; no Release here.
		if !c.closed {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		return 0
	}

	cookieMgr, err := core2.GetCookieManager()
	// QueryInterface increments the reference count, so release the temporary
	// ICoreWebView2_2 interface regardless of the outcome.
	releaseWebview2Com(unsafe.Pointer(core2), &core2.Vtbl.IUnknownVtbl)
	if err != nil || cookieMgr == nil {
		slog.Error("获取 WebView2 CookieManager 失败", slogx.Error(err))
		// core is a borrowed reference owned by WebView2; no Release here.
		if !c.closed {
			c.sendEvent(WebviewLoginEvent{WindowClosed: true})
		}
		return 0
	}

	c.mu.Lock()
	if c.closed {
		// The controller is owned by c.controller (released by cleanup()), but
		// core and the cookie manager were not stored. core is borrowed
		// (WebView2 releases it); the cookie manager was AddRef'd by
		// GetCookieManager, so Release it here.
		c.mu.Unlock()
		releaseWebview2Com(unsafe.Pointer(cookieMgr), &cookieMgr.Vtbl.IUnknownVtbl)
		return 0
	}
	// Like the controller, the core webview handed back by the controller is
	// a borrowed reference: AddRef before storing so cleanup() can Release it
	// safely (mirrors wails' chromium.go). The cookie manager from
	// GetCookieManager is already AddRef'd (caller releases it), so it needs
	// no extra AddRef.
	core.AddRef()
	c.core = core
	c.cookieManager = cookieMgr
	c.mu.Unlock()

	_ = core.Navigate(loginURL)

	// Arm the cookie polling timer. The wParam of the resulting WM_TIMER
	// messages carries cookiePollTimerID.
	user32SetTimer.Call(h.hwnd, cookiePollTimerID, uintptr(webviewLoginPollInterval/time.Millisecond), 0)
	return 0
}

// webview2CookiesHandler receives the cookie list of each GetCookies call.
type webview2CookiesHandler struct {
	ctrl *webviewLoginController
}

func (h *webview2CookiesHandler) QueryInterface(_, _ uintptr) uintptr { return 0 }
func (h *webview2CookiesHandler) AddRef() uintptr                     { return 1 }
func (h *webview2CookiesHandler) Release() uintptr                    { return 1 }

func (h *webview2CookiesHandler) GetCookiesCompleted(errorCode uintptr, result *webview2.ICoreWebView2CookieList) uintptr {
	c := h.ctrl

	// Runs on the UI thread, so inFlight needs no lock.
	c.inFlight = false

	if errorCode != 0 || result == nil {
		if result != nil {
			releaseWebview2Com(unsafe.Pointer(result), &result.Vtbl.IUnknownVtbl)
		}
		return 0
	}
	c.handleCookies(result)
	return 0
}
