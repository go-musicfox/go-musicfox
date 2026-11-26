//go:build !darwin && !windows && !linux

package ui

// webviewLoginAvailable reports whether the native WebView login window is
// available on the current platform. It is implemented on macOS (WKWebView),
// Windows (WebView2) and Linux (WebKitGTK); unsupported platforms fall back
// to this stub. The login entry is hidden when it reports false.
func webviewLoginAvailable() bool {
	return false
}

// webviewLoginController is a no-op stub on unsupported platforms. The login
// entry button is hidden when the WebView is unavailable.
type webviewLoginController struct{}

func newWebviewLoginController() *webviewLoginController {
	return &webviewLoginController{}
}

func (c *webviewLoginController) Open() {}

func (c *webviewLoginController) Close() {}

func (c *webviewLoginController) PopEvent() (WebviewLoginEvent, bool) {
	return WebviewLoginEvent{}, false
}
