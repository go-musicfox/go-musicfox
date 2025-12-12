package ui

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

const LastfmQRAuthPageType model.PageType = "lastfm_qr_auth"

type lastfmQRGeneratedMsg struct {
	qrView     string
	qrCodePath string
	token      string
	url        string
}

type lastfmQRErrorMsg struct{ err error }

// LastfmQRAuthPage Last.fm 二维码授权页面
type LastfmQRAuthPage struct {
	netease *Netease
	from    model.Page

	token       string
	url         string
	qrCodeView  string
	qrCodePath  string
	statusMsg   string
	loading     *model.Loading
	AfterAction func()
}

func (p *LastfmQRAuthPage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return true
}

func (p *LastfmQRAuthPage) Msg() tea.Msg {
	return nil
}

func NewLastfmQRAuthPage(netease *Netease, from model.Page, afterAction func()) *LastfmQRAuthPage {
	page := &LastfmQRAuthPage{
		netease:     netease,
		from:        from,
		AfterAction: afterAction,
		statusMsg:   "正在生成二维码，请稍候...",
	}
	page.loading = model.NewLoading(netease.MustMain(), &model.MenuItem{Title: "Last.fm 二维码授权"})
	page.loading.DisplayNotOnlyOnMain()
	return page
}

func (p *LastfmQRAuthPage) Init() tea.Cmd {
	p.loading.Start()
	return p.generateQRCodeCmd
}

func (p *LastfmQRAuthPage) Type() model.PageType {
	return LastfmQRAuthPageType
}

func (p *LastfmQRAuthPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case lastfmQRGeneratedMsg:
		p.loading.Complete()
		p.qrCodeView = msg.qrView
		p.qrCodePath = msg.qrCodePath
		p.token = msg.token
		p.url = msg.url
		p.statusMsg = "请使用可扫码设备扫码并在浏览器授权"
		return p, nil

	case lastfmQRErrorMsg:
		p.loading.Complete()
		p.statusMsg = util.SetFgStyle("发生错误: "+msg.err.Error(), termenv.ANSIBrightRed)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "b", "esc":
			if p.qrCodePath != "" {
				_ = os.Remove(p.qrCodePath)
			}
			return p.from, p.netease.RerenderCmd(true)
		case "v":
			if p.qrCodePath != "" {
				err := open.Start(p.qrCodePath)
				if err != nil {
					p.statusMsg = util.SetFgStyle("打开二维码失败: "+err.Error(), termenv.ANSIBrightRed)
				}
			}
		case "enter":
			// 用户确认已授权，尝试获取 session
			return p.confirmAuth()
		}
	}

	return p, nil
}

func (p *LastfmQRAuthPage) View(a *model.App) string {
	var builder strings.Builder

	var top int
	mainPage := p.netease.MustMain()
	builder.WriteString(mainPage.TitleView(a, &top))
	builder.WriteString(mainPage.MenuTitleView(a, &top, &model.MenuItem{Title: "Last.fm 二维码授权"}))
	builder.WriteString("\n\n")
	top += 2

	if p.qrCodeView != "" {
		qrLines := strings.Split(strings.TrimSuffix(p.qrCodeView, "\n"), "\n")
		if len(qrLines) == 0 {
			return builder.String()
		}

		qrWidth := runewidth.StringWidth(qrLines[0])
		padding := (a.WindowWidth() - qrWidth) / 2
		if padding < 0 {
			padding = 0
		}
		space := strings.Repeat(" ", padding)

		for _, line := range qrLines {
			builder.WriteString(space)
			builder.WriteString(line)
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

	if p.qrCodeView != "" {
		confirmTip := "授权完成后按 'Enter' 继续"
		padding = (a.WindowWidth() - runewidth.StringWidth(confirmTip)) / 2
		if padding < 0 {
			padding = 0
		}
		builder.WriteString(strings.Repeat(" ", padding))
		builder.WriteString(util.SetFgStyle(confirmTip, termenv.ANSIBrightBlue))
		builder.WriteString("\n")
	}

	bottomTip := "按 'b' 或 'esc' 返回"
	padding = (a.WindowWidth() - runewidth.StringWidth(bottomTip)) / 2
	if padding < 0 {
		padding = 0
	}
	builder.WriteString(strings.Repeat(" ", padding))
	builder.WriteString(util.SetFgStyle(bottomTip, termenv.ANSIBrightBlack))
	builder.WriteString("\n")

	if p.qrCodePath != "" {
		viewTip := "按 'v' 键可在外部查看器中打开二维码图片"
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
func (p *LastfmQRAuthPage) generateQRCodeCmd() tea.Msg {
	if !lastfm.IsAvailable() {
		return lastfmQRErrorMsg{fmt.Errorf("请确保正确设置 API key 及 secret")}
	}

	token, url, err := p.netease.lastfm.GetAuthUrlWithToken()
	if err != nil {
		slog.Error("获取 Last.fm 授权 token 失败", slogx.Error(err))
		return lastfmQRErrorMsg{fmt.Errorf("token 或 url 获取失败")}
	}

	slog.Info("lastfm auth url", slog.String("url", url))

	// 生成二维码
	path, buffer, err := app.GenQRCode("qrcode_lastfm.png", url)
	if err != nil {
		slog.Error("生成二维码失败", slogx.Error(err))
		return lastfmQRErrorMsg{err}
	}

	return lastfmQRGeneratedMsg{
		qrView:     buffer.String(),
		qrCodePath: path,
		token:      token,
		url:        url,
	}
}

// confirmAuth 用户确认已授权，尝试获取 session
func (p *LastfmQRAuthPage) confirmAuth() (model.Page, tea.Cmd) {
	loading := model.NewLoading(p.netease.MustMain(), &model.MenuItem{Title: "Last.fm 二维码授权"})
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	// 获取 session key
	sessionKey, err := p.netease.lastfm.GetSession(p.token)
	if err != nil {
		p.statusMsg = util.SetFgStyle("获取授权失败，请确认已在浏览器中完成授权", termenv.ANSIBrightRed)
		slog.Error("sessionKey 获取失败", slogx.Error(err))
		return p, nil
	}

	// 获取用户信息
	user, err := p.netease.lastfm.GetUserInfo(map[string]any{})
	if err != nil {
		p.statusMsg = util.SetFgStyle("用户信息获取失败", termenv.ANSIBrightRed)
		slog.Error("用户信息获取失败", slogx.Error(err))
		return p, nil
	}

	// 保存用户信息
	p.netease.lastfm.InitUserInfo(&storage.LastfmUser{
		Id:         user.Id,
		Name:       user.Name,
		RealName:   user.RealName,
		Url:        user.Url,
		SessionKey: sessionKey,
	})

	// 清理二维码文件
	if p.qrCodePath != "" {
		_ = os.Remove(p.qrCodePath)
	}

	// 执行回调
	if p.AfterAction != nil {
		p.AfterAction()
	}

	// 显示成功通知
	notify.Notify(notify.NotifyContent{
		Title:   "授权成功",
		Text:    fmt.Sprintf("Last.fm 用户 %s 授权成功", p.netease.lastfm.UserName()),
		GroupId: types.GroupID,
	})

	return p.netease.MustMain(), p.netease.RerenderCmd(true)
}
