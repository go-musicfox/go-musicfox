package ui

import "time"

// Constants shared by all platform implementations of the WebView login
// window. Platform-specific files must not redefine them.
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
