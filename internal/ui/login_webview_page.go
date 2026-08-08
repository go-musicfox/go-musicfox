package ui

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/mattn/go-runewidth"

	apputils "github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

const WebviewLoginPageType model.PageType = "webview_login"

// WebviewLoginEvent is emitted by the WebView login controller.
type WebviewLoginEvent struct {
	// CookieString carries the standard "name=value; name2=value2" cookie
	// string once a successful login (MUSIC_U cookie) is detected.
	CookieString string
	// WindowClosed reports that the user manually closed the login window.
	WindowClosed bool
}

// webviewLoginResultMsg carries the result of the cookie verification chain.
type webviewLoginResultMsg struct {
	err error
}

// WebviewLoginPage is the TUI page for the macOS WebView login flow. It opens
// a native WKWebView window (macOS only) and polls the controller for login
// events via tea.Tick, mirroring the QR login page state machine.
type WebviewLoginPage struct {
	netease *Netease
	from    model.Page

	controller *webviewLoginController
	statusMsg  string
	checking   bool
	AfterLogin LoginCallback

	backBtnHovered bool
	backBtnRowY    int
	backBtnStartX  int
	backBtnEndX    int
	mousePointer   string
}

func (p *WebviewLoginPage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return true
}

func (p *WebviewLoginPage) Msg() tea.Msg {
	return tickPollingMsg{}
}

func NewWebviewLoginPage(netease *Netease, from model.Page, afterLogin LoginCallback) *WebviewLoginPage {
	return &WebviewLoginPage{
		netease:    netease,
		from:       from,
		controller: newWebviewLoginController(),
		statusMsg:  model.T(MsgLoginWebviewWaiting),
		AfterLogin: afterLogin,
	}
}

func (p *WebviewLoginPage) Init() tea.Cmd {
	p.controller.Open()
	return tickPolling(time.Second)
}

func (p *WebviewLoginPage) Type() model.PageType {
	return WebviewLoginPageType
}

func (p *WebviewLoginPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		oldHovered := p.backBtnHovered
		oldPointer := p.mousePointer
		bcChanged, bcOver := pageBreadcrumbMotion(a, p.netease.MustMain(), mouse.X, mouse.Y)
		p.backBtnHovered = mouse.Y == p.backBtnRowY && mouse.X >= p.backBtnStartX && mouse.X < p.backBtnEndX
		p.mousePointer = "default"
		if p.backBtnHovered || bcOver {
			p.mousePointer = "pointer"
		}
		if p.backBtnHovered != oldHovered || p.mousePointer != oldPointer || bcChanged {
			return p, tea.Sequence(p.netease.RerenderCmd(true), a.SetMousePointer(p.mousePointer))
		}
		return p, nil

	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		if mouse.Button == tea.MouseLeft {
			if newPage := pageBreadcrumbClick(a, p.netease.MustMain(), mouse.X, mouse.Y); newPage != nil {
				// The breadcrumb click bypasses p.back(), so close the native
				// login window explicitly: otherwise the window, the polling
				// goroutine and the Accessory activation policy leak.
				p.controller.Close()
				return newPage, p.netease.RerenderCmd(true)
			}
			if mouse.Y == p.backBtnRowY && mouse.X >= p.backBtnStartX && mouse.X < p.backBtnEndX {
				return p.back()
			}
		}
		return p, nil

	case tickPollingMsg:
		ev, ok := p.controller.PopEvent()
		if !ok {
			return p, tickPolling(time.Second)
		}
		if ev.WindowClosed {
			p.statusMsg = model.T(MsgLoginWebviewCancelled)
			return p.back()
		}
		if p.checking {
			// A verification chain is already in flight; ignore duplicates.
			return p, tickPolling(time.Second)
		}
		p.checking = true
		p.statusMsg = model.T(MsgLoginWebviewVerifying)
		return p, webviewCheckCookieCmd(ev.CookieString)

	case webviewLoginResultMsg:
		p.checking = false
		if msg.err != nil {
			// Cookie invalid/expired: keep the window open and keep polling.
			p.statusMsg = util.SetFgStyle(model.T(MsgLoginWebviewFailed)+msg.err.Error(), lipgloss.BrightRed)
			return p, tickPolling(time.Second)
		}
		p.statusMsg = model.T(MsgLoginWebviewSuccess)
		// Synchronous page switch: Close() dispatches to the main thread with
		// waitUntilDone:YES, so the native window is already gone by the time
		// the new page renders. Do not change this to an asynchronous close.
		p.controller.Close()
		cmd := p.netease.RerenderCmd(true)
		if newPage := p.loginSuccessHandle(p.netease); newPage != nil {
			return newPage, cmd
		}
		return p.netease.MustMain(), cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "b", "esc", "q":
			return p.back()
		}
	}

	return p, nil
}

func (p *WebviewLoginPage) back() (model.Page, tea.Cmd) {
	p.controller.Close()
	return p.from, p.netease.RerenderCmd(true)
}

func (p *WebviewLoginPage) View(a *model.App) string {
	var builder strings.Builder

	var top int
	mainPage := p.netease.MustMain()
	builder.WriteString(pageTitleView(a, mainPage, &top))
	topBefore := top
	builder.WriteString(pageMenuTitleViewWithBack(a, mainPage, &top, &model.MenuItem{Title: model.T(MsgLoginWebviewPageTitle)}, p.backBtnHovered))
	p.backBtnRowY = pageMenuTitleRow(a, mainPage, topBefore)
	p.backBtnStartX = max(0, mainPage.MenuStartColumn()-pageBackButtonWidth)
	p.backBtnEndX = p.backBtnStartX + pageBackButtonWidth
	builder.WriteString("\n\n")
	top += 2

	padding := (a.WindowWidth() - runewidth.StringWidth(p.statusMsg)) / 2
	if padding < 0 {
		padding = 0
	}
	builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", padding)))
	builder.WriteString(p.statusMsg)
	builder.WriteString("\n\n")

	bottomTip := "Press 'b' or 'esc' to return"
	padding = (a.WindowWidth() - runewidth.StringWidth(bottomTip)) / 2
	if padding < 0 {
		padding = 0
	}
	builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", padding)))
	builder.WriteString(util.SetFgStyle(bottomTip, lipgloss.BrightBlack))
	builder.WriteString("\n")

	return finishCustomPageView(&builder, a)
}

// webviewCheckCookieCmd parses the cookie string from the WebView and validates
// it against the existing login chain (mirrors checkCookieCmd in login_page.go).
// A failure means the cookie is stale; the page keeps the window open and
// continues polling instead of closing it.
func webviewCheckCookieCmd(cookieStr string) tea.Cmd {
	return func() tea.Msg {
		err := apputils.ParseCookieFromStr(cookieStr, appCookieJar)
		if err != nil {
			return webviewLoginResultMsg{err: fmt.Errorf(model.T(MsgLoginCookieInvalid), err)}
		}

		neteaseutil.SetGlobalCookieJar(appCookieJar)
		jar, err := apputils.RefreshCookieJar()
		if err != nil {
			slog.Error("WebView Cookie 登录失败", slogx.Error(err))
			return webviewLoginResultMsg{err: fmt.Errorf("Cookie 登录失败: %w", err)}
		}

		slog.Info("使用 WebView Cookie 登录成功")
		appCookieJar = jar
		neteaseutil.SetGlobalCookieJar(appCookieJar)
		if err = appCookieJar.Save(); err != nil {
			slog.Warn("刷新token成功但保存 Cookie 失败", slogx.Error(err))
		}

		return webviewLoginResultMsg{err: nil}
	}
}

// loginSuccessHandle runs the shared post-login chain: persist the cookie jar
// and fire the LoginCallback.
func (p *WebviewLoginPage) loginSuccessHandle(n *Netease) model.Page {
	if err := appCookieJar.Save(); err != nil {
		slog.Warn("持久化 Cookie 失败", slogx.Error(err))
	}

	if err := n.LoginCallback(); err != nil {
		slog.Error("login callback error", slogx.Error(err))
	}

	var newPage model.Page
	if p.AfterLogin != nil {
		newPage = p.AfterLogin()
	}
	return newPage
}
