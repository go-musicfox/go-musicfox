//go:build darwin

// Package webkit provides a minimal purego/objc bridge to WebKit for the
// WebView login window (WKWebView + WKHTTPCookieStore). No CGO is involved.
package webkit

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var (
	importOnce sync.Once
	webKitLib  uintptr
)

func importFramework() {
	importOnce.Do(func() {
		var err error
		webKitLib, err = purego.Dlopen("/System/Library/Frameworks/WebKit.framework/WebKit", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
	})
}

func init() {
	importFramework()
	class_WKWebViewConfiguration = objc.GetClass("WKWebViewConfiguration")
	class_WKWebsiteDataStore = objc.GetClass("WKWebsiteDataStore")
	class_WKWebView = objc.GetClass("WKWebView")
	class_WKHTTPCookieStore = objc.GetClass("WKHTTPCookieStore")
	class_NSHTTPCookie = objc.GetClass("NSHTTPCookie")
	class_NSURLRequest = objc.GetClass("NSURLRequest")

	// getAllCookiesWithCompletionHandler: was renamed to getAllCookies: on
	// newer macOS (e.g. 27); resolve the name at runtime so both work.
	sel_instancesRespondToSelector := objc.RegisterName("instancesRespondToSelector:")
	if objc.ID(class_WKHTTPCookieStore).Send(sel_instancesRespondToSelector, objc.RegisterName("getAllCookiesWithCompletionHandler:")) != 0 {
		sel_getAllCookies = objc.RegisterName("getAllCookiesWithCompletionHandler:")
	} else {
		sel_getAllCookies = objc.RegisterName("getAllCookies:")
	}
}

var (
	class_WKWebViewConfiguration objc.Class
	class_WKWebsiteDataStore     objc.Class
	class_WKWebView              objc.Class
	class_WKHTTPCookieStore      objc.Class
	class_NSHTTPCookie           objc.Class
	class_NSURLRequest           objc.Class
)

var (
	sel_initWithFrameConfiguration         = objc.RegisterName("initWithFrame:configuration:")
	sel_setWebsiteDataStore                = objc.RegisterName("setWebsiteDataStore:")
	sel_nonPersistentDataStore             = objc.RegisterName("nonPersistentDataStore")
	sel_loadRequest                        = objc.RegisterName("loadRequest:")
	sel_httpCookieStore                    = objc.RegisterName("httpCookieStore")
	sel_getAllCookies                      objc.SEL // resolved at init
	sel_requestWithURL                     = objc.RegisterName("requestWithURL:")
	sel_name                               = objc.RegisterName("name")
	sel_value                              = objc.RegisterName("value")
)

// WKWebViewConfiguration mirrors the WKWebViewConfiguration class.
type WKWebViewConfiguration struct {
	core.NSObject
}

func WKWebViewConfiguration_alloc() WKWebViewConfiguration {
	return WKWebViewConfiguration{
		core.NSObject{
			ID: objc.ID(class_WKWebViewConfiguration).Send(macdriver.SEL_alloc),
		},
	}
}

func (c WKWebViewConfiguration) Init() WKWebViewConfiguration {
	c.ID = c.Send(macdriver.SEL_init)
	return c
}

func (c WKWebViewConfiguration) SetWebsiteDataStore(store WKWebsiteDataStore) {
	c.Send(sel_setWebsiteDataStore, store.ID)
}

// WKWebsiteDataStore mirrors the WKWebsiteDataStore class.
type WKWebsiteDataStore struct {
	core.NSObject
}

// NonPersistentDataStore returns an in-memory data store (never persisted to
// disk).
func WKWebsiteDataStore_NonPersistentDataStore() WKWebsiteDataStore {
	return WKWebsiteDataStore{
		core.NSObject{
			ID: objc.ID(class_WKWebsiteDataStore).Send(sel_nonPersistentDataStore),
		},
	}
}

// HttpCookieStore returns the cookie store associated with the data store.
// httpCookieStore is a property of WKWebsiteDataStore, not of WKWebView.
func (s WKWebsiteDataStore) HttpCookieStore() WKHTTPCookieStore {
	return WKHTTPCookieStore{
		core.NSObject{
			ID: s.Send(sel_httpCookieStore),
		},
	}
}

// WKWebView mirrors the WKWebView class.
type WKWebView struct {
	core.NSObject
}

func WKWebView_alloc() WKWebView {
	return WKWebView{
		core.NSObject{
			ID: objc.ID(class_WKWebView).Send(macdriver.SEL_alloc),
		},
	}
}

func (v WKWebView) InitWithFrameConfiguration(frame cocoa.NSRect, configuration WKWebViewConfiguration) WKWebView {
	v.ID = v.Send(sel_initWithFrameConfiguration,
		frame.Origin.X, frame.Origin.Y,
		frame.Size.Width, frame.Size.Height,
		configuration.ID,
	)
	return v
}

func (v WKWebView) LoadRequest(request NSURLRequest) {
	v.Send(sel_loadRequest, request.ID)
}

// WKHTTPCookieStore mirrors the WKHTTPCookieStore class.
type WKHTTPCookieStore struct {
	core.NSObject
}

// GetAllCookiesWithCompletionHandler asynchronously fetches all cookies and
// invokes the completion block with an NSArray of NSHTTPCookie objects. The
// block is called on a private WebKit queue, not the main thread. The actual
// selector is resolved at init (getAllCookiesWithCompletionHandler: on older
// macOS, getAllCookies: on newer macOS).
func (s WKHTTPCookieStore) GetAllCookiesWithCompletionHandler(handler objc.Block) {
	s.Send(sel_getAllCookies, handler)
}

// NSHTTPCookie mirrors the NSHTTPCookie class. Only the readonly name/value
// properties are bridged.
type NSHTTPCookie struct {
	core.NSObject
}

func (c NSHTTPCookie) Name() core.NSString {
	return core.NSString{NSObject: core.NSObject{ID: c.Send(sel_name)}}
}

func (c NSHTTPCookie) Value() core.NSString {
	return core.NSString{NSObject: core.NSObject{ID: c.Send(sel_value)}}
}

// NSURLRequest mirrors the NSURLRequest class.
type NSURLRequest struct {
	core.NSObject
}

func NSURLRequest_RequestWithURL(url core.NSURL) NSURLRequest {
	return NSURLRequest{
		core.NSObject{
			ID: objc.ID(class_NSURLRequest).Send(sel_requestWithURL, url.ID),
		},
	}
}
