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

	"github.com/go-musicfox/go-musicfox/internal/configs"
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
	tips          string
	AfterLogin    LoginCallback

	// 以下字段用于鼠标点击区域的计算与命中
	accountRowY  int // 账号输入框所在的行号（1-based）
	passwordRowY int // 密码输入框所在的行号（1-based）
	buttonsRowY  int // 提交/扫码按钮所在行号（1-based）
	submitStartX int // 提交按钮起始 X（0-based）
	submitEndX   int // 提交按钮结束 X（0-based，闭区间）
	qrStartX     int // 扫码按钮起始 X（0-based）
	qrEndX       int // 扫码按钮结束 X（0-based，闭区间）
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

	// 鼠标事件处理
	if mouse, ok := msg.(tea.MouseMsg); ok {
		// 仅处理左键按下的点击
		if mouse.Button == tea.MouseButtonLeft && mouse.Action == tea.MouseActionPress {
			// 行坐标转换为 1-based 与 View 中记录的行号一致
			y := mouse.Y + 1
			x := mouse.X

			// 点击输入框：设置焦点
			if y == l.accountRowY {
				l.index = 0
				// 焦点与样式更新
				l.accountInput.Focus()
				l.accountInput.Prompt = model.GetFocusedPrompt()
				l.accountInput.TextStyle = util.GetPrimaryFontStyle()

				l.passwordInput.Blur()
				l.passwordInput.Prompt = model.GetBlurredPrompt()
				l.passwordInput.TextStyle = lipgloss.NewStyle()

				// 按钮样式同步
				l.submitButton = model.GetBlurredSubmitButton()
				l.qrLoginButton = model.GetBlurredButton(l.qrButtonTextByStep())
				return l, tickLogin(time.Nanosecond)
			}
			if y == l.passwordRowY {
				l.index = 1
				// 焦点与样式更新
				l.passwordInput.Focus()
				l.passwordInput.Prompt = model.GetFocusedPrompt()
				l.passwordInput.TextStyle = util.GetPrimaryFontStyle()

				l.accountInput.Blur()
				l.accountInput.Prompt = model.GetBlurredPrompt()
				l.accountInput.TextStyle = lipgloss.NewStyle()

				// 按钮样式同步
				l.submitButton = model.GetBlurredSubmitButton()
				l.qrLoginButton = model.GetBlurredButton(l.qrButtonTextByStep())
				return l, tickLogin(time.Nanosecond)
			}

			// 点击按钮：触发提交或扫码登录
			if y == l.buttonsRowY {
				// 命中提交按钮
				if x >= l.submitStartX && x <= l.submitEndX {
					l.index = submitIndex
					return l.enterHandler()
				}
				// 命中扫码按钮
				if x >= l.qrStartX && x <= l.qrEndX {
					l.index = qrLoginIndex
					return l.enterHandler()
				}
			}
		}
		// 其他鼠标事件交给输入框以便光标闪烁等
		return l.updateLoginInputs(mouse)
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

	lineCount := 0
	write := func(s string) {
		builder.WriteString(s)
		lineCount += strings.Count(s, "\n")
	}
	curRow := func() int { return lineCount + 1 }

	// title
	if configs.AppConfig.Theme.ShowTitle {
		write(mainPage.TitleView(a, &top))
	} else {
		top++
	}

	// menu title
	write(mainPage.MenuTitleView(a, &top, l.menuTitle))
	write("\n\n\n")
	top += 2

	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	for i, input := range inputs {
		if mainPage.MenuStartColumn() > 0 {
			write(strings.Repeat(" ", mainPage.MenuStartColumn()))
		}

		write(input.View())

		var valueLen int
		if input.Value() == "" {
			valueLen = runewidth.StringWidth(input.Placeholder)
		} else {
			valueLen = runewidth.StringWidth(input.Value())
		}
		if spaceLen := l.netease.WindowWidth() - mainPage.MenuStartColumn() - valueLen - 3; spaceLen > 0 {
			write(strings.Repeat(" ", spaceLen))
		}

		// 记录输入框所在行号（1-based）
		if i == 0 {
			l.accountRowY = curRow()
		} else if i == 1 {
			l.passwordRowY = curRow()
		}

		top++

		if i < len(inputs)-1 {
			write("\n\n")
			top++
		}
	}

	write("\n\n")
	top++
	if mainPage.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	write(l.tips)
	write("\n\n")
	top++
	if mainPage.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	// 记录按钮所在行（1-based）
	l.buttonsRowY = curRow()

	// 计算按钮的起止 X 坐标（0-based）
	submitX := mainPage.MenuStartColumn()
	if submitX < 0 {
		submitX = 0
	}
	submitW := lipgloss.Width(l.submitButton)
	l.submitStartX = submitX
	l.submitEndX = submitX + submitW - 1

	write(l.submitButton)

	btnBlank := "    "
	write(btnBlank)
	// 扫码按钮坐标
	qrX := submitX + submitW + lipgloss.Width(btnBlank)
	qrW := lipgloss.Width(l.qrLoginButton)
	l.qrStartX = qrX
	l.qrEndX = qrX + qrW - 1

	write(l.qrLoginButton)

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - lipgloss.Width(l.submitButton) - lipgloss.Width(l.qrLoginButton) - lipgloss.Width(btnBlank)
	if spaceLen > 0 {
		write(strings.Repeat(" ", spaceLen))
	}
	write("\n")

	if a.WindowHeight() > top+3 {
		write(strings.Repeat("\n", a.WindowHeight()-top-3))
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

// loginByQRCode 跳转到二维码登录界面
func (l *LoginPage) loginByQRCode() (model.Page, tea.Cmd) {
	qrPage := NewQRLoginPage(l.netease, l, l.AfterLogin)
	return qrPage, qrPage.Init()
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
