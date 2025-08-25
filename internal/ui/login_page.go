package ui

import (
	"log/slog"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

const LoginPageType model.PageType = "login"

const (
	submitIndex  = 2 // skip account and password input
	qrLoginIndex = 3
)

// login tick
type tickLoginMsg struct{}

func tickLogin(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return tickLoginMsg{}
	})
}

type LoginPage struct {
	netease *Netease

	menuTitle     *model.MenuItem
	index         int
	accountInput  textinput.Model
	passwordInput textinput.Model
	submitButton  string
	qrLoginButton string
	qrLoginStep   int
	qrLoginUniKey string
	tips          string
	AfterLogin    LoginCallback
}

func NewLoginPage(netease *Netease) (login *LoginPage) {
	accountInput := textinput.New()
	accountInput.Placeholder = " 手机号或邮箱"
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

	login = &LoginPage{
		netease:       netease,
		menuTitle:     &model.MenuItem{Title: "用户登录", Subtitle: "手机号或邮箱"},
		accountInput:  accountInput,
		passwordInput: passwordInput,
		submitButton:  model.GetBlurredSubmitButton(),
	}
	login.qrLoginButton = model.GetBlurredButton(login.qrButtonTextByStep())

	return
}

func (l *LoginPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LoginPage) Type() model.PageType {
	return LoginPageType
}

func (l *LoginPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	var (
		key tea.KeyMsg
		ok  bool
	)

	if _, ok = msg.(tickLoginMsg); ok {
		return l, nil
	}

	if key, ok = msg.(tea.KeyMsg); !ok {
		return l.updateLoginInputs(msg)
	}

	switch key.String() {
	case "b":
		if l.index != submitIndex && l.index != qrLoginIndex {
			return l.updateLoginInputs(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		l.qrLoginStep = 0
		if l.index == qrLoginIndex {
			l.qrLoginButton = model.GetFocusedButton(l.qrButtonTextByStep())
		} else {
			l.qrLoginButton = model.GetBlurredButton(l.qrButtonTextByStep())
		}
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
			if s == "left" && l.index == qrLoginIndex {
				l.index--
			} else if s == "right" && l.index == submitIndex {
				l.index++
			}
		case "up", "shift+tab":
			l.index--
		default:
			l.index++
		}

		if l.index > qrLoginIndex {
			l.index = 0
		} else if l.index < 0 {
			l.index = qrLoginIndex
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

		if l.index == submitIndex {
			l.submitButton = model.GetFocusedSubmitButton()
		} else {
			l.submitButton = model.GetBlurredSubmitButton()
		}

		if l.index == qrLoginIndex {
			l.qrLoginButton = model.GetFocusedButton(l.qrButtonTextByStep())
		} else {
			l.qrLoginButton = model.GetBlurredButton(l.qrButtonTextByStep())
		}

		return l, nil
	}

	// Handle character input and blinks
	return l.updateLoginInputs(msg)
}

func (l *LoginPage) View(a *model.App) string {
	var (
		builder  strings.Builder
		top      int // 距离顶部的行数
		mainPage = l.netease.MustMain()
	)

	// title
	if configs.ConfigRegistry.Main.ShowTitle {
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
	builder.WriteString(l.qrLoginButton)

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - runewidth.StringWidth(types.SubmitText) - runewidth.StringWidth(l.qrButtonTextByStep()) - len(btnBlank)
	if spaceLen > 0 {
		builder.WriteString(strings.Repeat(" ", spaceLen))
	}
	builder.WriteString("\n")

	if a.WindowHeight() > top+3 {
		builder.WriteString(strings.Repeat("\n", a.WindowHeight()-top-3))
	}

	return builder.String()
}

func (l *LoginPage) Msg() tea.Msg {
	return tickLoginMsg{}
}

func (l *LoginPage) updateLoginInputs(msg tea.Msg) (model.Page, tea.Cmd) {
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

func (l *LoginPage) qrButtonTextByStep() string {
	switch l.qrLoginStep {
	case 1:
		return "已扫码登录，继续"
	case 0:
		fallthrough
	default:
		return "扫码登录"
	}
}

func (l *LoginPage) enterHandler() (model.Page, tea.Cmd) {
	loading := model.NewLoading(l.netease.MustMain(), l.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case submitIndex:
		// 提交
		if len(l.accountInput.Value()) <= 0 || len(l.passwordInput.Value()) <= 0 {
			l.tips = util.SetFgStyle("请输入账号或密码", termenv.ANSIBrightRed)
			return l, nil
		}
		return l.loginByAccount()
	case qrLoginIndex:
		// 扫码登录
		return l.loginByQRCode()
	}

	return l, tickLogin(time.Nanosecond)
}

func (l *LoginPage) loginByAccount() (model.Page, tea.Cmd) {
	var (
		code float64
		err  error
	)

	if strings.ContainsRune(l.accountInput.Value(), '@') {
		loginService := service.LoginEmailService{
			Email:    l.accountInput.Value(),
			Password: l.passwordInput.Value(),
		}
		code, _ = loginService.LoginEmail()
	} else {
		var (
			phone       = l.accountInput.Value()
			countryCode = "86"
		)
		if strings.HasPrefix(phone, "+") && strings.ContainsRune(phone, ' ') {
			if items := strings.Split(phone, " "); len(items) == 2 {
				countryCode, phone = strings.TrimLeft(items[0], "+"), items[1]
			}
		}
		loginService := service.LoginCellphoneService{
			Phone:       phone,
			Password:    l.passwordInput.Value(),
			Countrycode: countryCode,
		}
		code, _, err = loginService.LoginCellphone()
		if err != nil {
			l.tips = util.SetFgStyle("登录失败："+err.Error(), termenv.ANSIBrightRed)
			return l, tickLogin(time.Nanosecond)
		}
	}

	codeType := _struct.CheckCode(code)
	switch codeType {
	case _struct.UnknownError:
		l.tips = util.SetFgStyle("未知错误，请稍后再试~", termenv.ANSIBrightRed)
		return l, tickLogin(time.Nanosecond)
	case _struct.NetworkError:
		l.tips = util.SetFgStyle("网络异常，请稍后再试~", termenv.ANSIBrightRed)
		return l, tickLogin(time.Nanosecond)
	case _struct.Success:
		l.tips = ""
		if newPage := l.loginSuccessHandle(l.netease); newPage != nil {
			return newPage, l.netease.Tick(time.Nanosecond)
		}
		return l.netease.MustMain(), model.TickMain(time.Nanosecond)
	default:
		l.tips = util.SetFgStyle("你是个好人，但我们不合适(╬▔皿▔)凸 ", termenv.ANSIBrightRed) +
			util.SetFgStyle("(账号或密码错误)", termenv.ANSIBrightBlack)
		return l, tickLogin(time.Nanosecond)
	}
}

func (l *LoginPage) loginByQRCode() (model.Page, tea.Cmd) {
	qrService := service.LoginQRService{}
	if l.qrLoginStep == 0 {
		code, resp, url, err := qrService.GetKey()
		errHandler := func(err error) (model.Page, tea.Cmd) {
			l.tips = util.SetFgStyle("生成二维码失败，请稍候再试", termenv.ANSIBrightRed)
			if err != nil {
				slog.Error("生成二维码失败", slogx.Error(err))
			}
			return l, nil
		}
		if err != nil {
			return errHandler(err)
		}
		if code != 200 || url == "" {
			return errHandler(errors.Errorf("code: %f, resp: %s", code, string(resp)))
		}
		path, err := app.GenQRCode("qrcode.png", url)
		if err != nil {
			return errHandler(err)
		}
		_ = open.Start(path)
		l.tips = util.SetFgStyle("请扫描二维码(MUSICFOX_ROOT/qrcode.png)登录后，点击「继续」", termenv.ANSIBrightRed)
		l.qrLoginStep++
		l.qrLoginButton = model.GetFocusedButton(l.qrButtonTextByStep())
		l.qrLoginUniKey = qrService.UniKey
		return l, nil
	}
	qrService.UniKey = l.qrLoginUniKey
	errHandler := func(err error) (model.Page, tea.Cmd) {
		l.tips = util.SetFgStyle("校验二维码失败，请稍候再试", termenv.ANSIBrightRed)
		if err != nil {
			slog.Error("生成二维码失败", slogx.Error(err))
		}
		return l, nil
	}
	code, resp, err := qrService.CheckQR()
	if err != nil {
		return errHandler(err)
	}
	if code != 803 {
		return errHandler(errors.Errorf("checkQR code: %f, resp: %s", code, string(resp)))
	}
	l.tips = ""
	if newPage := l.loginSuccessHandle(l.netease); newPage != nil {
		return newPage, l.netease.Tick(time.Nanosecond)
	}
	return l.netease.MustMain(), model.TickMain(time.Nanosecond)
}

func (l *LoginPage) loginSuccessHandle(n *Netease) model.Page {
	if err := n.LoginCallback(); err != nil {
		slog.Error("login callback error", slogx.Error(err))
	}

	var newPage model.Page
	if l.AfterLogin != nil {
		newPage = l.AfterLogin()
	}
	return newPage
}
