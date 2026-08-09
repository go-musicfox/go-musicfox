//go:build linux

// Package webkitgtk provides a minimal purego bridge to WebKitGTK for the
// WebView login window (WebKitWebView + WebKitCookieManager). Both the
// WebKitGTK 6.0 (GTK4) and 4.1 (GTK3) API versions are supported: the 6.0
// stack is preferred at runtime and the 4.1 stack is used as a fallback on
// older distributions (e.g. Debian 11/12, Ubuntu 22.04). When neither stack
// is available every exported function degrades to a no-op / zero value and
// the login page falls back to the form login. No CGO is involved; every C
// function is resolved at runtime with purego.Dlopen + Dlsym + RegisterFunc
// (missing symbols stay nil instead of panicking, see registerSymbol) so the
// binary links cleanly without WebKitGTK headers.
package webkitgtk

import (
	"log/slog"
	"unsafe"

	"github.com/ebitengine/purego"
)

// API version constants.
const (
	// Version6 is the WebKitGTK 6.0 API (GTK4).
	Version6 = 6
	// Version41 is the WebKitGTK 4.1 API (GTK3).
	Version41 = 41
	// Version40 is the legacy WebKitGTK 4.0 API (GTK3, pre-2.40): no
	// WebKitNetworkSession; the web view uses the default web context and
	// cookies come from the context's cookie manager.
	Version40 = 40

	// GTK_WINDOW_TOPLEVEL is the GtkWindowType for a regular window; it is
	// required by the GTK3 gtk_window_new (GTK4's gtk_window_new takes no
	// argument).
	GTK_WINDOW_TOPLEVEL = 0
)

// webkitVersion is the active API version: Version6, Version41, or 0 when no
// supported WebKitGTK stack could be loaded. It is written once during init()
// before any goroutine runs, so the exported getter needs no lock.
var webkitVersion int

// WebKit bindings (same symbols in libwebkitgtk-6.0.so.4 and
// libwebkit2gtk-4.1.so.0).
var (
	webkitNetworkSessionNewEphemeral       func() uintptr
	webkitNetworkSessionGetCookieManager   func(session uintptr) uintptr
	webkitWebViewNewWithNetworkSession     func(session uintptr) uintptr
	webkitWebViewLoadURI                   func(webView uintptr, uri *byte)
	webkitCookieManagerGetAllCookies       func(manager, cancellable, callback, userData uintptr)
	webkitCookieManagerGetAllCookiesFinish func(manager, result, err uintptr) uintptr
)

// WebKitGTK 4.0 bindings (libwebkit2gtk-4.0.so.37, pre-2.40 API): the legacy
// web view creation and cookie manager access without WebKitNetworkSession.
var (
	webkitWebViewNew                    func() uintptr
	webkitWebContextGetDefault          func() uintptr
	webkitWebContextGetCookieManager    func(ctx uintptr) uintptr
	webkitCookieManagerGetCookiesAsync  func(manager, cancellable, callback, userData uintptr)
	webkitCookieManagerGetCookiesFinish func(manager, result, err uintptr) uintptr
)

// GTK bindings shared by both GTK4 and GTK3 (identical signatures; registered
// from whichever GTK library loaded).
var (
	gtkInitCheck            func(argc, argv uintptr) bool
	gtkWindowSetDefaultSize func(window uintptr, width, height int32)
	gtkWindowSetTitle       func(window uintptr, title *byte)
	gtkWindowPresent        func(window uintptr)
	gtkWindowClose          func(window uintptr)
	gtkMain                 func()
	gtkMainQuit             func()
)

// GTK4 bindings (libgtk-4.so.1).
var (
	gtkWindowNew4     func() uintptr
	gtkWindowSetChild func(window, widget uintptr)
)

// GTK3 bindings (libgtk-3.so.0).
var (
	gtkWindowNew3    func(windowType int32) uintptr
	gtkContainerAdd  func(container, widget uintptr)
	gtkWidgetShowAll func(widget uintptr)
)

// GObject/GLib bindings (shared by both versions).
var (
	gSignalConnectData func(instance uintptr, detailedSignal *byte, cHandler, data, destroyData uintptr, connectFlags int32) uint
	gObjectUnref       func(object uintptr)
	gTimeoutAdd        func(interval uint32, function, data uintptr) uint
	gListLength        func(list uintptr) uint
	gListNthData       func(list uintptr, n uint) uintptr
	gListFree          func(list uintptr)
	gErrorFree         func(err uintptr)
)

// Soup bindings (libsoup-3.0, shared by both versions).
var (
	soupCookieGetName  func(cookie uintptr) *byte
	soupCookieGetValue func(cookie uintptr) *byte
	soupCookieFree     func(cookie uintptr)
)

func init() {
	// The GObject/GLib/Soup libraries are shared by both stacks; load them
	// once. Each failure is logged by dlopen so the reason (missing library
	// vs. missing symbols) is visible in musicfox.log.
	sharedOK :=
		dlopen("libgobject-2.0.so.0", func(lib uintptr) {
			registerSymbol(&gSignalConnectData, lib, "g_signal_connect_data")
			registerSymbol(&gObjectUnref, lib, "g_object_unref")
		}) &&
			dlopen("libglib-2.0.so.0", func(lib uintptr) {
				registerSymbol(&gTimeoutAdd, lib, "g_timeout_add")
				registerSymbol(&gListLength, lib, "g_list_length")
				registerSymbol(&gListNthData, lib, "g_list_nth_data")
				registerSymbol(&gListFree, lib, "g_list_free")
				registerSymbol(&gErrorFree, lib, "g_error_free")
			}) &&
			dlopen("libsoup-3.0.so.0", func(lib uintptr) {
				registerSymbol(&soupCookieGetName, lib, "soup_cookie_get_name")
				registerSymbol(&soupCookieGetValue, lib, "soup_cookie_get_value")
				registerSymbol(&soupCookieFree, lib, "soup_cookie_free")
			})

	// Prefer the WebKitGTK 6.0 (GTK4) stack. A stack is only usable when the
	// library actually exports the WebKit API we call: an older library may
	// load fine (RTLD_LAZY) while missing newer symbols (e.g.
	// webkit_network_session_new_ephemeral, added in WebKitGTK 2.40). Those
	// symbols stay nil through registerSymbol, so a nil check here decides
	// whether the stack really works.
	webkitUsable := func() bool {
		return webkitNetworkSessionNewEphemeral != nil && webkitWebViewNewWithNetworkSession != nil && webkitCookieManagerGetAllCookies != nil
	}
	if sharedOK {
		// Prefer the WebKitGTK 6.0 (GTK4) stack.
		if dlopen("libwebkitgtk-6.0.so.4", registerWebKit) {
			if webkitUsable() && dlopen("libgtk-4.so.1", registerGTK4) {
				webkitVersion = Version6
				return
			}
			if !webkitUsable() {
				slog.Warn("webkitgtk: libwebkitgtk-6.0.so.4 loaded but missing required symbols (webkit_network_session_new_ephemeral or friends); falling back to 4.1")
			}
		}
		// Fall back to WebKitGTK 4.1 (GTK3).
		if dlopen("libwebkit2gtk-4.1.so.0", registerWebKit) && webkitUsable() && dlopen("libgtk-3.so.0", registerGTK3) {
			webkitVersion = Version41
			return
		}
		// Last resort: legacy WebKitGTK 4.0 (GTK3, < 2.40), preinstalled on
		// Debian 11 and Ubuntu 22.04 GNOME desktops (via gnome-software).
		// It has no WebKitNetworkSession; the legacy symbols drive the web
		// view and cookie manager instead.
		webkitUsable40 := func() bool {
			return webkitWebViewNew != nil && webkitWebContextGetDefault != nil && webkitCookieManagerGetCookiesAsync != nil
		}
		if dlopen("libwebkit2gtk-4.0.so.37", registerWebKit40) && webkitUsable40() && dlopen("libgtk-3.so.0", registerGTK3) {
			webkitVersion = Version40
			return
		}
	}
	// No supported stack is available: keep version 0. The exported functions
	// degrade to no-ops / zero values through their nil guards.
	webkitVersion = 0
}

// dlopen loads a shared library and registers its functions through the given
// callback. Failures are non-fatal: the callback is only invoked on success.
// The error is logged so musicfox.log shows whether the library is missing or
// a dependency of it failed to resolve.
func dlopen(name string, register func(lib uintptr)) bool {
	lib, err := purego.Dlopen(name, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		slog.Debug("webkitgtk: dlopen failed", "lib", name, "error", err)
		return false
	}
	register(lib)
	return true
}

// registerSymbol binds a single C symbol to the given Go function pointer.
// Unlike purego.RegisterLibFunc it never panics: a missing symbol (e.g. an
// older library version without a newer API) leaves the function pointer nil,
// and every exported wrapper degrades to a no-op / zero value through its nil
// guard. dlopen may succeed on a library that lacks a symbol (RTLD_LAZY only
// fails on the missing library itself), so this is the last line of defence
// that keeps the WebView feature optional.
func registerSymbol(fptr any, lib uintptr, name string) {
	sym, err := purego.Dlsym(lib, name)
	if err != nil {
		return
	}
	purego.RegisterFunc(fptr, sym)
}

// registerWebKit binds the WebKit functions that are identical in both API
// versions.
func registerWebKit(lib uintptr) {
	registerSymbol(&webkitNetworkSessionNewEphemeral, lib, "webkit_network_session_new_ephemeral")
	registerSymbol(&webkitNetworkSessionGetCookieManager, lib, "webkit_network_session_get_cookie_manager")
	registerSymbol(&webkitWebViewNewWithNetworkSession, lib, "webkit_web_view_new_with_network_session")
	registerSymbol(&webkitWebViewLoadURI, lib, "webkit_web_view_load_uri")
	registerSymbol(&webkitCookieManagerGetAllCookies, lib, "webkit_cookie_manager_get_all_cookies")
	registerSymbol(&webkitCookieManagerGetAllCookiesFinish, lib, "webkit_cookie_manager_get_all_cookies_finish")
}

// registerWebKit40 binds the legacy 4.0 API (WebKitGTK < 2.40) symbols.
func registerWebKit40(lib uintptr) {
	registerSymbol(&webkitWebViewNew, lib, "webkit_web_view_new")
	registerSymbol(&webkitWebContextGetDefault, lib, "webkit_web_context_get_default")
	registerSymbol(&webkitWebContextGetCookieManager, lib, "webkit_web_context_get_cookie_manager")
	registerSymbol(&webkitWebViewLoadURI, lib, "webkit_web_view_load_uri")
	registerSymbol(&webkitCookieManagerGetCookiesAsync, lib, "webkit_cookie_manager_get_cookies_async")
	registerSymbol(&webkitCookieManagerGetCookiesFinish, lib, "webkit_cookie_manager_get_cookies_finish")
}

// registerGTKCommon binds the GTK functions that share a signature between
// GTK4 and GTK3.
func registerGTKCommon(lib uintptr) {
	registerSymbol(&gtkInitCheck, lib, "gtk_init_check")
	registerSymbol(&gtkWindowSetDefaultSize, lib, "gtk_window_set_default_size")
	registerSymbol(&gtkWindowSetTitle, lib, "gtk_window_set_title")
	registerSymbol(&gtkWindowPresent, lib, "gtk_window_present")
	registerSymbol(&gtkWindowClose, lib, "gtk_window_close")
	registerSymbol(&gtkMain, lib, "gtk_main")
	registerSymbol(&gtkMainQuit, lib, "gtk_main_quit")
}

// registerGTK4 binds the GTK4-specific functions.
func registerGTK4(lib uintptr) {
	registerGTKCommon(lib)
	registerSymbol(&gtkWindowNew4, lib, "gtk_window_new")
	registerSymbol(&gtkWindowSetChild, lib, "gtk_window_set_child")
}

// registerGTK3 binds the GTK3-specific functions.
func registerGTK3(lib uintptr) {
	registerGTKCommon(lib)
	registerSymbol(&gtkWindowNew3, lib, "gtk_window_new")
	registerSymbol(&gtkContainerAdd, lib, "gtk_container_add")
	registerSymbol(&gtkWidgetShowAll, lib, "gtk_widget_show_all")
}

// WebKit.

// WebKitGTKVersion returns the active API version: Version6, Version41, or 0
// when no supported WebKitGTK stack could be loaded.
func WebKitGTKVersion() int {
	return webkitVersion
}

func NetworkSessionNewEphemeral() uintptr {
	if webkitNetworkSessionNewEphemeral == nil {
		return 0
	}
	return webkitNetworkSessionNewEphemeral()
}

func NetworkSessionGetCookieManager(session uintptr) uintptr {
	if webkitNetworkSessionGetCookieManager == nil {
		return 0
	}
	return webkitNetworkSessionGetCookieManager(session)
}

func WebViewNewWithNetworkSession(session uintptr) uintptr {
	if webkitWebViewNewWithNetworkSession == nil {
		return 0
	}
	return webkitWebViewNewWithNetworkSession(session)
}

// WebViewNew creates a WebKitWebView on the legacy 4.0 API (uses the default
// web context).
func WebViewNew() uintptr {
	if webkitWebViewNew == nil {
		return 0
	}
	return webkitWebViewNew()
}

// WebContextGetDefault returns the default WebKitWebContext (borrowed
// reference) on the legacy 4.0 API.
func WebContextGetDefault() uintptr {
	if webkitWebContextGetDefault == nil {
		return 0
	}
	return webkitWebContextGetDefault()
}

// WebContextGetCookieManager returns the cookie manager (borrowed reference)
// of a web context on the legacy 4.0 API.
func WebContextGetCookieManager(ctx uintptr) uintptr {
	if webkitWebContextGetCookieManager == nil {
		return 0
	}
	return webkitWebContextGetCookieManager(ctx)
}

func WebViewLoadURI(webView uintptr, uri string) {
	if webkitWebViewLoadURI == nil {
		return
	}
	webkitWebViewLoadURI(webView, cString(uri))
}

func CookieManagerGetAllCookies(manager, cancellable, callback, userData uintptr) {
	if webkitCookieManagerGetAllCookies == nil {
		return
	}
	webkitCookieManagerGetAllCookies(manager, cancellable, callback, userData)
}

func CookieManagerGetAllCookiesFinish(manager, result, err uintptr) uintptr {
	if webkitCookieManagerGetAllCookiesFinish == nil {
		return 0
	}
	return webkitCookieManagerGetAllCookiesFinish(manager, result, err)
}

// CookieManagerGetCookiesAsync starts an asynchronous cookie fetch on the
// legacy 4.0 API.
func CookieManagerGetCookiesAsync(manager, cancellable, callback, userData uintptr) {
	if webkitCookieManagerGetCookiesAsync == nil {
		return
	}
	webkitCookieManagerGetCookiesAsync(manager, cancellable, callback, userData)
}

// CookieManagerGetCookiesFinish completes an asynchronous cookie fetch on the
// legacy 4.0 API and returns the GList of SoupCookie.
func CookieManagerGetCookiesFinish(manager, result, err uintptr) uintptr {
	if webkitCookieManagerGetCookiesFinish == nil {
		return 0
	}
	return webkitCookieManagerGetCookiesFinish(manager, result, err)
}

// GTK.

// GtkInitCheck initializes GTK without aborting the process on failure (e.g. a
// stale DISPLAY). It reports whether the initialization succeeded; it also
// returns false when no GTK library is loaded.
func GtkInitCheck() bool {
	if gtkInitCheck == nil {
		return false
	}
	return gtkInitCheck(0, 0)
}

// GtkWindowNew creates a new toplevel window: gtk_window_new() on GTK4,
// gtk_window_new(GTK_WINDOW_TOPLEVEL) on GTK3.
func GtkWindowNew() uintptr {
	switch webkitVersion {
	case Version6:
		if gtkWindowNew4 == nil {
			return 0
		}
		return gtkWindowNew4()
	case Version41, Version40:
		if gtkWindowNew3 == nil {
			return 0
		}
		return gtkWindowNew3(GTK_WINDOW_TOPLEVEL)
	default:
		return 0
	}
}

// GtkWindowSetChild attaches a child widget with the GTK4 API
// (gtk_window_set_child). It is a no-op on GTK3; use GtkWindowAddChild for a
// version-neutral attachment.
func GtkWindowSetChild(window, widget uintptr) {
	if gtkWindowSetChild == nil {
		return
	}
	gtkWindowSetChild(window, widget)
}

// GtkWindowAddChild attaches the child widget to the window:
// gtk_window_set_child on GTK4, gtk_container_add on GTK3.
func GtkWindowAddChild(window, child uintptr) {
	switch webkitVersion {
	case Version6:
		if gtkWindowSetChild != nil {
			gtkWindowSetChild(window, child)
		}
	case Version41, Version40:
		if gtkContainerAdd != nil {
			gtkContainerAdd(window, child)
		}
	}
}

// GtkShowWidget makes the widget visible. GTK3 widgets are not shown by
// default even when their toplevel window is presented, so this calls
// gtk_widget_show_all; GTK4 shows children automatically and this is a no-op.
func GtkShowWidget(widget uintptr) {
	if (webkitVersion != Version41 && webkitVersion != Version40) || gtkWidgetShowAll == nil {
		return
	}
	gtkWidgetShowAll(widget)
}

func GtkWindowSetDefaultSize(window uintptr, width, height int32) {
	if gtkWindowSetDefaultSize == nil {
		return
	}
	gtkWindowSetDefaultSize(window, width, height)
}

func GtkWindowSetTitle(window uintptr, title string) {
	if gtkWindowSetTitle == nil {
		return
	}
	gtkWindowSetTitle(window, cString(title))
}

func GtkWindowPresent(window uintptr) {
	if gtkWindowPresent == nil {
		return
	}
	gtkWindowPresent(window)
}

func GtkWindowClose(window uintptr) {
	if gtkWindowClose == nil {
		return
	}
	gtkWindowClose(window)
}

func GtkMain() {
	if gtkMain == nil {
		return
	}
	gtkMain()
}

func GtkMainQuit() {
	if gtkMainQuit == nil {
		return
	}
	gtkMainQuit()
}

// GObject/GLib.

// GSignalConnectCloseRequest connects the window-close signal of the active
// API version: "close-request" (GTK4) or "delete-event" (GTK3). The callback
// must have the signature func(window uintptr, extra uintptr) int32; for
// "delete-event" the second argument is the GdkEvent pointer and for
// "close-request" it is the user_data, both of which the controller ignores.
// Returning 0 (FALSE) allows the window to close and GTK destroys it in both
// versions.
func GSignalConnectCloseRequest(window uintptr, cb uintptr) uint {
	if gSignalConnectData == nil {
		return 0
	}
	signal := "close-request"
	if webkitVersion == Version41 || webkitVersion == Version40 {
		signal = "delete-event"
	}
	return gSignalConnectData(window, cString(signal), cb, 0, 0, 0)
}

func GSignalConnectData(instance uintptr, detailedSignal string, cHandler, data, destroyData uintptr, connectFlags int32) uint {
	if gSignalConnectData == nil {
		return 0
	}
	return gSignalConnectData(instance, cString(detailedSignal), cHandler, data, destroyData, connectFlags)
}

func GObjectUnref(object uintptr) {
	if gObjectUnref == nil {
		return
	}
	gObjectUnref(object)
}

func GTimeoutAdd(interval uint32, function, data uintptr) uint {
	if gTimeoutAdd == nil {
		return 0
	}
	return gTimeoutAdd(interval, function, data)
}

func GListLength(list uintptr) uint {
	if gListLength == nil {
		return 0
	}
	return gListLength(list)
}

func GListNthData(list uintptr, n uint) uintptr {
	if gListNthData == nil {
		return 0
	}
	return gListNthData(list, n)
}

func GListFree(list uintptr) {
	if gListFree == nil {
		return
	}
	gListFree(list)
}

func GErrorFree(err uintptr) {
	if gErrorFree == nil {
		return
	}
	gErrorFree(err)
}

// Soup.

func SoupCookieGetName(cookie uintptr) string {
	if soupCookieGetName == nil {
		return ""
	}
	return goString(soupCookieGetName(cookie))
}

func SoupCookieGetValue(cookie uintptr) string {
	if soupCookieGetValue == nil {
		return ""
	}
	return goString(soupCookieGetValue(cookie))
}

func SoupCookieFree(cookie uintptr) {
	if soupCookieFree == nil {
		return
	}
	soupCookieFree(cookie)
}

// cString returns a NUL-terminated C string for the given Go string. The
// returned buffer is only valid for the duration of the synchronous call it is
// passed to; GTK/WebKit copy the strings they need.
func cString(s string) *byte {
	b := append([]byte(s), 0)
	return &b[0]
}

// goString converts a NUL-terminated C string into a Go string.
func goString(p *byte) string {
	if p == nil {
		return ""
	}
	n := 0
	for ptr := unsafe.Pointer(p); *(*byte)(unsafe.Add(ptr, n)) != 0; n++ {
	}
	return unsafe.String(p, n)
}
