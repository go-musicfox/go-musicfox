package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/mattn/go-runewidth"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/utils/app"
)

const VerifyPageType model.PageType = "verify"

type verifyGeneratedMsg struct {
	qrCode     string
	qrView     string
	qrCodePath string
}
type verifyStatusMsg struct {
	status float64
}
type verifyErrorMsg struct{ err error }

// VerifyPage 人机验证页面（登录接口返回 -462 时渲染验证二维码）
type VerifyPage struct {
	netease    *Netease
	from       model.Page
	verify     *app.VerifyData
	onVerified func() (model.Page, tea.Cmd) // 验证成功后的回调（重试登录）

	qrCode     string
	qrCodeView string
	qrCodePath string
	isExpired  bool
	statusMsg  string
	loading    *model.Loading

	backBtnHovered bool
	backBtnRowY    int
	backBtnStartX  int
	backBtnEndX    int
	mousePointer   string
}

func (p *VerifyPage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return true
}

func (p *VerifyPage) Msg() tea.Msg {
	return tickLoginMsg{}
}

func NewVerifyPage(netease *Netease, from model.Page, verify *app.VerifyData, onVerified func() (model.Page, tea.Cmd)) *VerifyPage {
	page := &VerifyPage{
		netease:    netease,
		from:       from,
		verify:     verify,
		onVerified: onVerified,
		statusMsg:  model.T(MsgVerifyGenerating),
		isExpired:  false,
	}
	page.loading = model.NewLoading(netease.MustMain(), &model.MenuItem{Title: model.T(MsgVerifyPageTitle)})
	page.loading.DisplayNotOnlyOnMain()
	return page
}

func (p *VerifyPage) Init() tea.Cmd {
	p.loading.Start()
	return p.generateVerifyQRCodeCmd
}

func (p *VerifyPage) Type() model.PageType {
	return VerifyPageType
}

func (p *VerifyPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
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

	case verifyGeneratedMsg:
		p.loading.Complete()
		p.qrCode = msg.qrCode
		p.qrCodeView = msg.qrView
		p.qrCodePath = msg.qrCodePath
		p.isExpired = false
		p.statusMsg = model.T(MsgVerifyScanPrompt)
		return p, tickPolling(time.Second)

	case tickPollingMsg:
		if p.qrCode == "" {
			return p, nil
		}
		return p, p.pollVerifyStatusCmd

	case verifyStatusMsg:
		switch int(msg.status) {
		case 20: // 验证成功，自动重试原登录
			p.statusMsg = model.T(MsgVerifySuccess)
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
			}
			if p.onVerified != nil {
				return p.onVerified()
			}
			return p.back()
		case 0: // 等待扫码
			p.statusMsg = model.T(MsgVerifyWaitingScan)
			return p, tickPolling(time.Second)
		case 10: // 已扫码待确认
			p.statusMsg = model.T(MsgVerifyWaitingConfirm)
			return p, tickPolling(time.Second)
		case 21: // 已失效
			p.statusMsg = model.T(MsgVerifyExpiredAction)
			p.isExpired = true
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
				p.qrCodePath = ""
			}
			return p, nil
		default:
			p.statusMsg = fmt.Sprintf(model.T(MsgVerifyUnknownStatus), int(msg.status))
			return p, nil
		}

	case verifyErrorMsg:
		p.loading.Complete()
		p.statusMsg = util.SetFgStyle(model.T(MsgVerifyError)+msg.err.Error(), lipgloss.BrightRed)
		return p, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "b", "esc", "q":
			return p.back()
		case "v":
			if p.qrCodePath != "" && !p.isExpired {
				err := open.Start(p.qrCodePath)
				if err != nil {
					p.statusMsg = util.SetFgStyle(model.T(MsgVerifyOpenImageFailed)+err.Error(), lipgloss.BrightRed)
				}
			}
		}
	}

	return p, nil
}

func (p *VerifyPage) back() (model.Page, tea.Cmd) {
	if p.qrCodePath != "" {
		_ = os.Remove(p.qrCodePath)
	}
	return p.from, p.netease.RerenderCmd(true)
}

func (p *VerifyPage) View(a *model.App) string {
	var builder strings.Builder

	var top int
	mainPage := p.netease.MustMain()
	builder.WriteString(pageTitleView(a, mainPage, &top))
	topBefore := top
	builder.WriteString(pageMenuTitleViewWithBack(a, mainPage, &top, &model.MenuItem{Title: model.T(MsgVerifyPageTitle)}, p.backBtnHovered))
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
			expiredMsg := model.T(MsgVerifyExpired)
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

// generateVerifyQRCodeCmd 异步获取并生成验证二维码
func (p *VerifyPage) generateVerifyQRCodeCmd() tea.Msg {
	qrCode, err := app.GetVerifyQRCode(*p.verify)
	if err != nil {
		return verifyErrorMsg{err}
	}
	params, _ := json.Marshal(p.verify.Params)
	url := fmt.Sprintf("https://st.music.163.com/encrypt-pages?qrCode=%s&verifyToken=%s&verifyId=%.0f&verifyType=%.0f&params=%s",
		qrCode, p.verify.VerifyToken, p.verify.VerifyID, p.verify.VerifyType, params)
	path, buffer, err := app.GenQRCode("verify_qrcode.png", url)
	if err != nil {
		return verifyErrorMsg{err}
	}
	return verifyGeneratedMsg{qrCode: qrCode, qrView: buffer.String(), qrCodePath: path}
}

// pollVerifyStatusCmd 轮询验证二维码状态
func (p *VerifyPage) pollVerifyStatusCmd() tea.Msg {
	status, err := app.CheckVerifyQRCodeStatus(p.qrCode)
	if err != nil {
		return verifyErrorMsg{err}
	}
	return verifyStatusMsg{status: status}
}
