package ui

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/skratchdot/open-golang/open"

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
		statusMsg:  "正在生成二维码，请稍候...",
		isExpired:  false,
	}
	page.loading = model.NewLoading(netease.MustMain(), &model.MenuItem{Title: "二维码登录"})
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

func (p *QRLoginPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case qrGeneratedMsg:
		p.loading.Complete()
		p.qrCodeView = msg.qrView
		p.uniKey = msg.uniKey
		p.qrCodePath = msg.qrCodePath
		p.isExpired = false
		p.statusMsg = "请使用网易云音乐APP扫码"
		return p, tickPolling(time.Second)

	case tickPollingMsg:
		if p.uniKey == "" {
			return p, nil
		}
		return p, p.pollQRStatusCmd

	case qrStatusMsg:
		switch int(msg.code) {
		case 803: // 登录成功
			p.statusMsg = "登录成功！"
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
			}
			cmd := p.netease.RerenderCmd(true)
			if newPage := p.loginSuccessHandle(p.netease); newPage != nil {
				return newPage, cmd
			}
			return p.netease.MustMain(), cmd
		case 800: // 已失效
			p.statusMsg = "二维码已失效，请按 'b' 或 'esc' 返回"
			p.isExpired = true
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
				p.qrCodePath = ""
			}
			return p, nil
		case 801: // 等待扫码
			p.statusMsg = "等待扫码..."
			return p, tickPolling(time.Second)
		case 802: // 已扫码待确认
			p.statusMsg = "已扫码，请在手机上确认登录"
			return p, tickPolling(time.Second)
		default:
			p.statusMsg = fmt.Sprintf("未知状态: %d，请返回重试", int(msg.code))
			return p, nil
		}

	case qrErrorMsg:
		p.loading.Complete()
		p.statusMsg = util.SetFgStyle("发生错误: "+msg.err.Error(), termenv.ANSIBrightRed)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "b", "esc", "q":
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
			}
			return p.from, p.netease.RerenderCmd(true)
		case "v":
			if p.qrCodePath != "" && !p.isExpired {
				err := open.Start(p.qrCodePath)
				if err != nil {
					p.statusMsg = util.SetFgStyle("打开二维码失败: "+err.Error(), termenv.ANSIBrightRed)
				}
			}
		}
	}

	return p, nil
}

func (p *QRLoginPage) View(a *model.App) string {
	var builder strings.Builder

	var top int
	mainPage := p.netease.MustMain()
	builder.WriteString(mainPage.TitleView(a, &top))
	builder.WriteString(mainPage.MenuTitleView(a, &top, &model.MenuItem{Title: "二维码登录"}))
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
		space := strings.Repeat(" ", padding)

		if p.isExpired {
			expiredMsg := "二维码已失效"
			msgWidth := runewidth.StringWidth(expiredMsg)
			msgPaddingLen := (qrWidth - msgWidth) / 2
			if msgPaddingLen < 0 {
				msgPaddingLen = 0
			}
			msgPadding := strings.Repeat(" ", msgPaddingLen)

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
				builder.WriteString(util.SetFgStyle(line, termenv.ANSIBrightRed))
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
	builder.WriteString(strings.Repeat(" ", padding))
	builder.WriteString(p.statusMsg)
	builder.WriteString("\n\n")

	bottomTip := "Press 'b' or 'esc' to return"
	padding = (a.WindowWidth() - runewidth.StringWidth(bottomTip)) / 2
	if padding < 0 {
		padding = 0
	}
	builder.WriteString(strings.Repeat(" ", padding))
	builder.WriteString(util.SetFgStyle(bottomTip, termenv.ANSIBrightBlack))
	builder.WriteString("\n")

	if p.qrCodePath != "" {
		viewTip := "Press 'v' to show image of qrcode"
		padding = (a.WindowWidth() - runewidth.StringWidth(viewTip)) / 2
		if padding < 0 {
			padding = 0
		}
		builder.WriteString(strings.Repeat(" ", padding))
		builder.WriteString(util.SetFgStyle(viewTip, termenv.ANSIBrightBlack))
		builder.WriteString("\n")
	} else {
		builder.WriteString("\n")
	}

	return builder.String()
}

// generateQRCodeCmd 异步获取和生成二维码
func (p *QRLoginPage) generateQRCodeCmd() tea.Msg {
	qrService := service.LoginQRService{}
	code, _, url, err := qrService.GetKey()
	if err != nil {
		return qrErrorMsg{err}
	}
	if code != 200 || url == "" {
		return qrErrorMsg{fmt.Errorf("无法获取二维码, code: %.0f", code)}
	}

	// 生成二维码
	path, buffer, err := app.GenQRCode("qrcode.png", url)
	if err != nil {
		return qrErrorMsg{err}
	}

	return qrGeneratedMsg{
		qrView:     buffer.String(),
		uniKey:     qrService.UniKey,
		qrCodePath: path,
	}
}

// pollQRStatusCmd 轮询二维码状态
func (p *QRLoginPage) pollQRStatusCmd() tea.Msg {
	qrService := service.LoginQRService{UniKey: p.uniKey}
	code, resp, err := qrService.CheckQR()
	if err != nil {
		return qrErrorMsg{err}
	}
	return qrStatusMsg{code: code, resp: resp}
}

// loginSuccessHandle 登录成功函数
func (p *QRLoginPage) loginSuccessHandle(n *Netease) model.Page {
	if err := n.LoginCallback(); err != nil {
		slog.Error("login callback error", slogx.Error(err))
	}

	var newPage model.Page
	if p.AfterLogin != nil {
		newPage = p.AfterLogin()
	}
	return newPage
}
