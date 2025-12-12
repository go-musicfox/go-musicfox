package ui

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/skratchdot/open-golang/open"
)

const LastfmAuthPageType model.PageType = "lastfm_auth"

type LastfmAuthPage struct {
	netease *Netease

	menuTitle       *model.MenuItem
	index           int
	accountInput    textinput.Model
	passwordInput   textinput.Model
	submitButton    string
	browserButton   string
	qrAuthButton    string
	browserAuthStep int
	tips            string
	AfterAction     func()

	submitIndex  int
	qrAuthIndex  int
	browserIndex int

	token      string
	url        string
	sessionKey string
}

func NewLastfmAuthPage(netease *Netease) *LastfmAuthPage {
	accountInput := textinput.New()
	accountInput.Placeholder = " 用户名或邮箱"
	accountInput.Focus()
	accountInput.Prompt = model.GetFocusedPrompt()
	accountInput.TextStyle = util.GetPrimaryFontStyle()
	accountInput.CharLimit = 32

	passwordInput := textinput.New()
	passwordInput.Placeholder = " 密码"
	passwordInput.Prompt = "> "
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 32

	page := &LastfmAuthPage{
		netease:       netease,
		menuTitle:     &model.MenuItem{Title: "Lastfm用户登录/授权"},
		accountInput:  accountInput,
		passwordInput: passwordInput,
		submitButton:  model.GetBlurredSubmitButton(),

		submitIndex:  2,
		qrAuthIndex:  3,
		browserIndex: 4,
	}
	page.qrAuthButton = model.GetBlurredButton(page.qrButtonTextByStep())
	page.browserButton = model.GetBlurredButton(page.browserButtonTextByStep())

	return page
}

func (l *LastfmAuthPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LastfmAuthPage) Type() model.PageType {
	return LastfmAuthPageType
}

func (l *LastfmAuthPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	var (
		key tea.KeyMsg
		ok  bool
	)

	if key, ok = msg.(tea.KeyMsg); !ok {
		return l.updateLoginInputs(msg)
	}

	switch key.String() {
	case "b":
		if l.index != submitIndex && l.index != l.qrAuthIndex {
			return l.updateLoginInputs(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		return l.netease.MustMain(), l.netease.RerenderCmd(true)
	case "tab", "shift+tab", "enter", "up", "down", "left", "right":
		s := key.String()

		// Did the user press enter while the submit button was focused?
		// If so, exit.
		if s == "enter" && l.index >= submitIndex {
			return l.enterHandler()
		}

		// 当focus在button上时，左右按键的特殊处理
		switch s {
		case "left", "right":
			if l.index < submitIndex {
				return l.updateLoginInputs(msg)
			}
			if s == "left" && l.index >= l.submitIndex {
				l.index--
			} else if s == "right" && l.index <= l.browserIndex {
				l.index++
			}
		case "up", "shift+tab":
			l.index--
		default:
			l.index++
		}

		if l.index > l.browserIndex {
			l.index = 0
		} else if l.index < 0 {
			l.index = l.browserIndex
		}

		for i := 0; i <= len(inputs)-1; i++ {
			if i != l.index {
				// Remove focused state
				inputs[i].Blur()
				inputs[i].Prompt = model.GetBlurredPrompt()
				inputs[i].TextStyle = lipgloss.NewStyle()
				continue
			}
			// Set focused state
			inputs[i].Focus()
			inputs[i].Prompt = model.GetFocusedPrompt()
			inputs[i].TextStyle = util.GetPrimaryFontStyle()
		}

		// l.accountInput = *inputs[0]
		// l.passwordInput = *inputs[1]

		l.tips = ""

		if l.index == l.submitIndex {
			l.tips = util.SetFgStyle("使用账号密码登录并授权", termenv.ANSIBrightBlue)
			l.submitButton = model.GetFocusedSubmitButton()
		} else {
			l.submitButton = model.GetBlurredSubmitButton()
		}

		if l.index == l.qrAuthIndex {
			l.tips = util.SetFgStyle("请使用可扫码设备扫码并在浏览器授权", termenv.ANSIBrightBlue)
			l.qrAuthButton = model.GetFocusedButton(l.qrButtonTextByStep())
		} else {
			l.qrAuthButton = model.GetBlurredButton(l.qrButtonTextByStep())
		}

		if l.index == l.browserIndex {
			l.tips = util.SetFgStyle("在默认浏览器中打开链接并授权", termenv.ANSIBrightBlue)
			l.browserButton = model.GetFocusedButton(l.browserButtonTextByStep())
		} else {
			l.browserButton = model.GetBlurredButton(l.browserButtonTextByStep())
		}

		return l, nil
	}

	// Handle character input and blinks
	return l.updateLoginInputs(msg)
}

func (l *LastfmAuthPage) View(a *model.App) string {
	var (
		builder  strings.Builder
		top      int // 距离顶部的行数
		mainPage = l.netease.MustMain()
	)

	// title
	if configs.AppConfig.Theme.ShowTitle {
		builder.WriteString(mainPage.TitleView(a, &top))
	} else {
		top++
	}

	// menu title
	builder.WriteString(mainPage.MenuTitleView(a, &top, l.menuTitle))
	builder.WriteString("\n\n\n")
	top += 2

	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	for i, input := range inputs {
		if mainPage.MenuStartColumn() > 0 {
			builder.WriteString(strings.Repeat(" ", mainPage.MenuStartColumn()))
		}

		builder.WriteString(input.View())

		var valueLen int
		if input.Value() == "" {
			valueLen = runewidth.StringWidth(input.Placeholder)
		} else {
			valueLen = runewidth.StringWidth(input.Value())
		}
		if spaceLen := l.netease.WindowWidth() - mainPage.MenuStartColumn() - valueLen - 3; spaceLen > 0 {
			builder.WriteString(strings.Repeat(" ", spaceLen))
		}

		top++

		if i < len(inputs)-1 {
			builder.WriteString("\n\n")
			top++
		}
	}

	builder.WriteString("\n\n")
	top++
	if mainPage.MenuStartColumn() > 0 {
		builder.WriteString(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	builder.WriteString(l.tips)
	builder.WriteString("\n\n")
	top++
	if mainPage.MenuStartColumn() > 0 {
		builder.WriteString(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	builder.WriteString(l.submitButton)

	btnBlank := "    "
	builder.WriteString(btnBlank)
	builder.WriteString(l.qrAuthButton)

	builder.WriteString(btnBlank)
	builder.WriteString(l.browserButton)

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - runewidth.StringWidth(types.SubmitText) - runewidth.StringWidth(l.qrButtonTextByStep()) - runewidth.StringWidth(l.browserButtonTextByStep()) - len(btnBlank)*2
	if spaceLen > 0 {
		builder.WriteString(strings.Repeat(" ", spaceLen))
	}
	builder.WriteString("\n")

	if a.WindowHeight() > top+3 {
		builder.WriteString(strings.Repeat("\n", a.WindowHeight()-top-3))
	}

	return builder.String()
}

func (l *LastfmAuthPage) Msg() tea.Msg {
	return nil
}

func (l *LastfmAuthPage) updateLoginInputs(msg tea.Msg) (model.Page, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	l.accountInput, cmd = l.accountInput.Update(msg)
	cmds = append(cmds, cmd)

	l.passwordInput, cmd = l.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l *LastfmAuthPage) qrButtonTextByStep() string {
	return "扫码授权"
}

func (l *LastfmAuthPage) browserButtonTextByStep() string {
	switch l.browserAuthStep {
	case 1:
		return "已授权，继续"
	case 0:
		fallthrough
	default:
		return "浏览器授权"
	}
}

func (l *LastfmAuthPage) enterHandler() (model.Page, tea.Cmd) {
	loading := model.NewLoading(l.netease.MustMain(), l.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case submitIndex:
		// 提交
		// 简单的账号密码判断
		if len(l.accountInput.Value()) < 2 || len(l.accountInput.Value()) > 15 || len(l.passwordInput.Value()) < 6 {
			l.tips = util.SetFgStyle("请正确输入账号或密码", termenv.ANSIBrightRed)
			return l, nil
		}
		return l.authByLogin()
	case l.qrAuthIndex:
		// 扫码授权
		return l.authByQRCode()
	case l.browserIndex:
		// 浏览器授权
		return l.authByBrower()
	}

	return l, tickLogin(time.Nanosecond)
}

func (l *LastfmAuthPage) getAuthUrlWithToken() bool {
	if l.token != "" && l.url != "" {
		return true
	}

	if !lastfm.IsAvailable() {
		l.tips = util.SetFgStyle("请确保正确设置 API key 及 secret", termenv.ANSIBrightRed)
		return false
	}

	var err error
	if l.token, l.url, err = l.netease.lastfm.GetAuthUrlWithToken(); err != nil {
		slog.Info("token", slog.Any("token", l.token))
		slog.Info("url", slog.Any("url", l.url))
		l.tips = util.SetFgStyle("token 或 url 获取失败", termenv.ANSIBrightRed)
		slog.Error("token 或 url 获取失败", slog.Any("error", err))
		return false
	}
	slog.Info("lastfm auth url", slog.String("url", l.url))
	return true
}

func (l *LastfmAuthPage) getSessionKey() bool {
	if l.sessionKey != "" {
		return true
	}

	var err error
	if l.sessionKey, err = l.netease.lastfm.GetSession(l.token); err != nil {
		l.tips = util.SetFgStyle("sessionKey 获取失败", termenv.ANSIBrightRed)
		slog.Error("sessionKey 获取失败", slogx.Error(err))
		return false
	}
	return true
}

func (l *LastfmAuthPage) initUserInfo() bool {
	user, err := l.netease.lastfm.GetUserInfo(map[string]any{})
	if err != nil {
		l.tips = util.SetFgStyle("用户信息获取失败", termenv.ANSIBrightRed)
		slog.Error("用户信息获取失败", slogx.Error(err))
		return false
	}

	l.netease.lastfm.InitUserInfo(&storage.LastfmUser{
		Id:         user.Id,
		Name:       user.Name,
		RealName:   user.RealName,
		Url:        user.Url,
		SessionKey: l.sessionKey,
	})
	return true
}

func (l *LastfmAuthPage) authByLogin() (model.Page, tea.Cmd) {
	var err error
	l.sessionKey, err = l.netease.lastfm.Login(l.accountInput.Value(), l.passwordInput.Value())
	if err != nil {
		l.tips = util.SetFgStyle("登录失败，请检查", termenv.ANSIBrightRed)
		slog.Error("登录失败", slogx.Error(err))
		return l, nil
	}

	if !l.initUserInfo() {
		return l, nil
	}
	return l.authSuccessHandle()
}

func (l *LastfmAuthPage) authByQRCode() (model.Page, tea.Cmd) {
	qrPage := NewLastfmQRAuthPage(l.netease, l, l.AfterAction)
	return qrPage, qrPage.Init()
}

func (l *LastfmAuthPage) authByBrower() (model.Page, tea.Cmd) {
	if l.browserAuthStep == 0 {
		if !l.getAuthUrlWithToken() {
			return l, nil
		}

		if err := open.Start(l.url); err != nil {
			l.tips = util.SetFgStyle("认证页打开失败，请确认浏览器是否工作", termenv.ANSIBrightRed)
			slog.Error("认证页打开失败", slogx.Error(err))
			return l, nil
		}
		l.tips = util.SetFgStyle("请在浏览器中授权后继续，若未正确跳转，请更换认证方式", termenv.ANSIBrightBlue)
		l.browserAuthStep++
		l.browserButton = model.GetFocusedButton(l.browserButtonTextByStep())
		return l, nil
	}

	if !l.getSessionKey() {
		return l, nil
	}
	if !l.initUserInfo() {
		return l, nil
	}
	return l.authSuccessHandle()
}

func (l *LastfmAuthPage) authSuccessHandle() (model.Page, tea.Cmd) {
	if l.AfterAction != nil {
		l.AfterAction()
	}

	notify.Notify(notify.NotifyContent{
		Title: "授权成功",
		// Text:    "Last.fm 授权成功",
		Text:    fmt.Sprintf("Last.fm 用户 %s 授权成功", l.netease.lastfm.UserName()),
		GroupId: types.GroupID,
	})
	return l.netease.MustMain(), model.TickMain(time.Second)
}
