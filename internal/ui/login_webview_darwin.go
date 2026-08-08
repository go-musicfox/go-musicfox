//go:build darwin

package ui

import (
	"strings"
	"sync"
	"time"

	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/webkit"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

// webviewLoginAvailable reports whether the native WKWebView login window is
// available on the current platform (macOS only).
const webviewLoginAvailable = true

const (
	// loginURL is the target page loaded by the WebView login window.
	loginURL = "https://music.163.com/#/login"
	// loginWindowTitle is the title of the WebView login window.
	loginWindowTitle = "网易云音乐登录"
	// webviewLoginWindowWidth/Height is the default login window size.
	webviewLoginWindowWidth, webviewLoginWindowHeight = 800.0, 600.0
	// webviewLoginPollInterval is the cookie polling interval.
	webviewLoginPollInterval = time.Second
	// musicUCookieName marks a successful login.
	musicUCookieName = "MUSIC_U"
)

// webviewLoginController opens the native WKWebView login window and polls the
// WKHTTPCookieStore until the MUSIC_U cookie appears. All ObjC window
// operations are dispatched to the main thread via the helper class; the
// polling goroutine and the WebKit completion block only touch the cookie
// store and the event channel, never the UI.
type webviewLoginController struct {
	mu sync.Mutex

	events chan WebviewLoginEvent

	window       cocoa.NSWindow
	cookieStore  webkit.WKHTTPCookieStore
	cookiesBlock objc.Block
	inFlight     bool

	// Retained ObjC objects, released deterministically in
	// handleWebviewCloseWindow (see login_webview_helper_darwin.go).
	config       webkit.WKWebViewConfiguration
	webView      webkit.WKWebView
	observerName core.NSString

	opened      bool
	closed      bool
	stopPolling chan struct{}
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

// Open creates the login window on the main thread and starts the cookie
// polling goroutine.
func (c *webviewLoginController) Open() {
	c.mu.Lock()
	if c.closed || c.opened {
		c.mu.Unlock()
		return
	}
	c.opened = true
	c.mu.Unlock()

	setWebviewLoginCtrl(c)
	webviewDispatchSync(sel_webviewCreateWindow)

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.stopPolling = make(chan struct{})
	c.mu.Unlock()

	errorx.Go(c.pollCookies)
}

// Close stops polling, closes the login window on the main thread and restores
// the global Prohibited activation policy. Safe to call multiple times; every
// exit path (success, user close, cancel) goes through here.
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

	setWebviewLoginCtrl(c)
	webviewDispatchSync(sel_webviewCloseWindow)
	setWebviewLoginCtrl(nil)
}

// pollCookies serializes cookie polls with an in-flight flag. It runs on its
// own goroutine and never touches UI objects.
func (c *webviewLoginController) pollCookies() {
	ticker := time.NewTicker(webviewLoginPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopPolling:
			return
		case <-ticker.C:
			c.pollOnce()
		}
	}
}

func (c *webviewLoginController) pollOnce() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.cookieStore.ID == 0 || c.cookiesBlock == 0 || c.inFlight {
		return
	}
	c.inFlight = true
	core.Autorelease(func() {
		c.cookieStore.GetAllCookiesWithCompletionHandler(c.cookiesBlock)
	})
}

// newCookiesBlock returns an ObjC block that marshals the cookies of each
// getAllCookiesWithCompletionHandler: call into a Go string and forwards
// events. It captures c instead of consulting the package-level controller, so
// a late callback from a previous login session operates on its own (already
// closed) controller and is harmless instead of corrupting the current one.
func (c *webviewLoginController) newCookiesBlock() objc.Block {
	return objc.NewBlock(func(_ objc.Block, cookies objc.ID) {
		c.handleCookies(cookies)
	})
}

// handleCookies runs on a private WebKit queue after each
// getAllCookiesWithCompletionHandler: call. It only marshals the cookies into
// a Go string and forwards events through the channel — no UI operations are
// allowed here.
func (c *webviewLoginController) handleCookies(cookies objc.ID) {
	core.Autorelease(func() {
		var (
			pairs []string
			found bool
		)
		arr := core.NSArray{NSObject: core.NSObject{ID: cookies}}
		count := arr.Count()
		for i := uint(0); i < count; i++ {
			cookie := webkit.NSHTTPCookie{NSObject: arr.ObjectAtIndex(i)}
			name := cookie.Name().String()
			value := cookie.Value().String()
			pairs = append(pairs, name+"="+value)
			if name == musicUCookieName {
				found = true
			}
		}

		c.mu.Lock()
		c.inFlight = false
		c.mu.Unlock()

		if !found {
			return
		}
		c.sendEvent(WebviewLoginEvent{CookieString: strings.Join(pairs, "; ")})
	})
}
