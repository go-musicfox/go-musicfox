//go:build !darwin

package ui

// webviewLoginAvailable reports whether the native WKWebView login window is
// available on the current platform. It is only implemented on macOS.
const webviewLoginAvailable = false

// webviewLoginController is a no-op stub on non-macOS platforms. The login
// entry button stays visible and shows an unsupported message when activated.
type webviewLoginController struct{}

func newWebviewLoginController() *webviewLoginController {
	return &webviewLoginController{}
}

func (c *webviewLoginController) Open() {}

func (c *webviewLoginController) Close() {}

func (c *webviewLoginController) PopEvent() (WebviewLoginEvent, bool) {
	return WebviewLoginEvent{}, false
}
