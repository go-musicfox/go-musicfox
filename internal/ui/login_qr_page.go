package ui

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/mattn/go-runewidth"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/core/qrlogin"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

const QRLoginPageType model.PageType = "qr_login"

type qrGeneratedMsg struct {
	qrView     string
	uniKey     string
	qrCodePath string
}
type qrStatusMsg struct {
	code float64
	resp []byte
}
type qrErrorMsg struct{ err error }

// tickPollingMsg 用于触发轮询
type tickPollingMsg struct{}

func tickPolling(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return tickPollingMsg{}
	})
}

// QRLoginPage 二维码登录页面
type QRLoginPage struct {
	netease *Netease
	from    model.Page

	uniKey     string
	qrCodeView string
	qrCodePath string
	isExpired  bool
	statusMsg  string
	loading    *model.Loading
	AfterLogin LoginCallback

	backBtnHovered bool
	backBtnRowY    int
	backBtnStartX  int
	backBtnEndX    int
	mousePointer   string
}

func (p *QRLoginPage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return true
}

func (p *QRLoginPage) Msg() tea.Msg {
	return tickLoginMsg{}
}

func NewQRLoginPage(netease *Netease, from model.Page, afterLogin LoginCallback) *QRLoginPage {
	page := &QRLoginPage{
		netease:    netease,
		from:       from,
		AfterLogin: afterLogin,
		statusMsg:  model.T(MsgQRLoginGenerating),
		isExpired:  false,
	}
	page.loading = model.NewLoading(netease.MustMain(), &model.MenuItem{Title: model.T(MsgQRLoginPageTitle)})
	page.loading.DisplayNotOnlyOnMain()
	return page
}

func (p *QRLoginPage) Init() tea.Cmd {
	p.loading.Start()
	return p.generateQRCodeCmd
}

func (p *QRLoginPage) Type() model.PageType {
	return QRLoginPageType
}

func (p *QRLoginPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
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
				return newPage, p.netease.RerenderCmd(true)
			}
			if mouse.Y == p.backBtnRowY && mouse.X >= p.backBtnStartX && mouse.X < p.backBtnEndX {
				return p.back()
			}
		}
		return p, nil

	case qrGeneratedMsg:
		p.loading.Complete()
		p.qrCodeView = msg.qrView
		p.uniKey = msg.uniKey
		p.qrCodePath = msg.qrCodePath
		p.isExpired = false
		p.statusMsg = model.T(MsgQRLoginScanPrompt)
		return p, tickPolling(time.Second)

	case tickPollingMsg:
		if p.uniKey == "" {
			return p, nil
		}
		return p, p.pollQRStatusCmd

	case qrStatusMsg:
		switch int(msg.code) {
		case 803: // 登录成功
			p.statusMsg = model.T(MsgQRLoginSuccess)
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
			}
			cmd := p.netease.RerenderCmd(true)
			if newPage := p.loginSuccessHandle(p.netease); newPage != nil {
				return newPage, cmd
			}
			return p.netease.MustMain(), cmd
		case 800: // 已失效
			p.statusMsg = model.T(MsgQRLoginExpiredAction)
			p.isExpired = true
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
				p.qrCodePath = ""
			}
			return p, nil
		case 801: // 等待扫码
			p.statusMsg = model.T(MsgQRLoginWaitingScan)
			return p, tickPolling(time.Second)
		case 802: // 已扫码待确认
			p.statusMsg = model.T(MsgQRLoginWaitingConfirm)
			return p, tickPolling(time.Second)
		default:
			p.statusMsg = fmt.Sprintf(model.T(MsgQRLoginUnknownStatus), int(msg.code))
			return p, nil
		}

	case qrErrorMsg:
		p.loading.Complete()
		p.statusMsg = util.SetFgStyle(model.T(MsgQRLoginError)+msg.err.Error(), lipgloss.BrightRed)
		return p, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "b", "esc", "q":
			return p.back()
		case "v":
			if p.qrCodePath != "" && !p.isExpired {
				err := open.Start(p.qrCodePath)
				if err != nil {
					p.statusMsg = util.SetFgStyle(model.T(MsgQRLoginOpenImageFailed)+err.Error(), lipgloss.BrightRed)
				}
			}
		}
	}

	return p, nil
}

func (p *QRLoginPage) back() (model.Page, tea.Cmd) {
	if p.qrCodePath != "" {
		_ = os.Remove(p.qrCodePath)
	}
	return p.from, p.netease.RerenderCmd(true)
}

func (p *QRLoginPage) View(a *model.App) string {
	var builder strings.Builder

	var top int
	mainPage := p.netease.MustMain()
	builder.WriteString(pageTitleView(a, mainPage, &top))
	topBefore := top
	builder.WriteString(pageMenuTitleViewWithBack(a, mainPage, &top, &model.MenuItem{Title: model.T(MsgQRLoginPageTitle)}, p.backBtnHovered))
	p.backBtnRowY = pageMenuTitleRow(a, mainPage, topBefore)
	p.backBtnStartX = max(0, mainPage.MenuStartColumn()-pageBackButtonWidth)
	p.backBtnEndX = p.backBtnStartX + pageBackButtonWidth
	builder.WriteString("\n\n")
	top += 2

	if p.qrCodeView != "" {
		qrLines := strings.Split(strings.TrimSuffix(p.qrCodeView, "\n"), "\n")
		if len(qrLines) == 0 {
			return builder.String() // 安全检查
		}

		qrWidth := runewidth.StringWidth(qrLines[0])
		padding := (a.WindowWidth() - qrWidth) / 2
		if padding < 0 {
			padding = 0
		}
		space := style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", padding))

		if p.isExpired {
			expiredMsg := model.T(MsgQRLoginExpired)
			msgWidth := runewidth.StringWidth(expiredMsg)
			msgPaddingLen := (qrWidth - msgWidth) / 2
			if msgPaddingLen < 0 {
				msgPaddingLen = 0
			}
			msgPadding := style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", msgPaddingLen))

			fullMsgLine := msgPadding + expiredMsg + msgPadding
			for runewidth.StringWidth(fullMsgLine) < qrWidth {
				fullMsgLine += " "
			}

			middleIndex := len(qrLines) / 2
			if middleIndex > 0 && middleIndex < len(qrLines) {
				qrLines[middleIndex] = fullMsgLine
			}
		}

		for _, line := range qrLines {
			builder.WriteString(space)
			if p.isExpired {
				builder.WriteString(util.SetFgStyle(line, lipgloss.BrightRed))
			} else {
				builder.WriteString(line)
			}
			builder.WriteString("\n")
		}
	} else {
		builder.WriteString(strings.Repeat("\n", 10))
	}
	builder.WriteString("\n")

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

	if p.qrCodePath != "" {
		viewTip := "Press 'v' to show image of qrcode"
		padding = (a.WindowWidth() - runewidth.StringWidth(viewTip)) / 2
		if padding < 0 {
			padding = 0
		}
		builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", padding)))
		builder.WriteString(util.SetFgStyle(viewTip, lipgloss.BrightBlack))
		builder.WriteString("\n")
	} else {
		builder.WriteString("\n")
	}

	return finishCustomPageView(&builder, a)
}

// generateQRCodeCmd 异步获取和生成二维码
func (p *QRLoginPage) generateQRCodeCmd() tea.Msg {
	cookieJar := neteaseutil.GetGlobalCookieJar()
	uniKey, url, err := qrlogin.GetKey(cookieJar)
	if err != nil {
		return qrErrorMsg{err}
	}

	// 生成二维码
	path, buffer, err := app.GenQRCode("qrcode.png", url)
	if err != nil {
		return qrErrorMsg{err}
	}

	return qrGeneratedMsg{
		qrView:     buffer.String(),
		uniKey:     uniKey,
		qrCodePath: path,
	}
}

// pollQRStatusCmd 轮询二维码状态
func (p *QRLoginPage) pollQRStatusCmd() tea.Msg {
	cookieJar := neteaseutil.GetGlobalCookieJar()
	code, resp, err := qrlogin.CheckStatus(p.uniKey, cookieJar)
	if err != nil {
		return qrErrorMsg{err}
	}
	return qrStatusMsg{code: code, resp: resp}
}

// loginSuccessHandle 登录成功函数
func (p *QRLoginPage) loginSuccessHandle(n *Netease) model.Page {
	// 先保存 cookie，确保登录成功后 cookie 被持久化
	// 即使后续 LoginCallback 失败（AccountInfo 失败），cookie 也已保存
	if err := n.engine.CompleteQRLogin(core.AppCookieJar()); err != nil {
		slog.Error("login callback error", slogx.Error(err))
	}

	var newPage model.Page
	if p.AfterLogin != nil {
		newPage = p.AfterLogin()
	}
	return newPage
}
