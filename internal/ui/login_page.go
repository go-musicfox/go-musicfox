package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-musicfox/netease-music/service"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	apputils "github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

const LoginPageType model.PageType = "login"

const (
	submitIndex  = 2 // skip account and password input
	qrLoginIndex = 3

	tabAccount = 0
	tabCookie  = 1
	tabCount   = 2

	idxTabAccount = -2 // 账号登录 Tab 的焦点索引
	idxTabCookie  = -1 // Cookie 登录 Tab 的焦点索引
)

var (
	tabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			Foreground(lipgloss.Color(termenv.ANSIBrightBlack.String())).
			BorderForeground(lipgloss.Color(termenv.ANSIBrightBlack.String())).
			Padding(0, 0)

	activeTabStyleGetter = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			Foreground(lipgloss.Color(configs.AppConfig.Theme.PrimaryColor)).
			BorderForeground(lipgloss.Color(configs.AppConfig.Theme.PrimaryColor)).
			Padding(0, 0)
	})
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
	tabIndex      int
	accountInput  textinput.Model
	passwordInput textinput.Model
	cookieInput   textinput.Model
	submitButton  string
	qrLoginButton string
	qrLoginStep   int
	tips          string
	AfterLogin    LoginCallback

	// 以下字段用于鼠标点击区域的计算与命中
	accountRowY  int // 账号输入框所在的行号（1-based）
	passwordRowY int // 密码输入框所在的行号（1-based）
	cookieRowY   int // Cookie输入框所在的行号（1-based）
	buttonsRowY  int // 提交/扫码按钮所在行号（1-based）
	submitStartX int // 提交按钮起始 X（0-based）
	submitEndX   int // 提交按钮结束 X（0-based，闭区间）
	qrStartX     int // 扫码按钮起始 X（0-based）
	qrEndX       int // 扫码按钮结束 X（0-based，闭区间）
	cookieStartX int // Cookie按钮起始 X（0-based）
	cookieEndX   int // Cookie按钮结束 X（0-based，闭区间）
	tabStartX    int // Tab 区域起始 X（0-based）
	tabEndX      int // Tab 区域结束 X（0-based，闭区间）
	tabsRowY     int // Tab 所在行号（1-based）
}

// 执行登录操作回显的信息结构体
type LoginMsg struct {
	err error
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

	cookieInput := textinput.New()
	cookieInput.Placeholder = " 请输入 Cookie"
	cookieInput.Prompt = "> "
	cookieInput.CharLimit = 5000

	login = &LoginPage{
		netease:       netease,
		menuTitle:     &model.MenuItem{Title: "用户登录", Subtitle: "手机号或邮箱"},
		accountInput:  accountInput,
		passwordInput: passwordInput,
		cookieInput:   cookieInput,
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

	var (
		key tea.KeyMsg
		ok  bool
	)

	if _, ok = msg.(tickLoginMsg); ok {
		return l, nil
	}

	if loginMsg, ok := msg.(LoginMsg); ok {
		if loginMsg.err != nil {
			l.tips = util.SetFgStyle(loginMsg.err.Error(), termenv.ANSIBrightRed)
			return l, nil
		}

		if newPage := l.loginSuccessHandle(l.netease); newPage != nil {
			return newPage, tea.Batch(
				tea.ClearScreen,
				model.TickMain(time.Nanosecond),
				l.netease.RerenderCmd(true),
			)
		}
		return l.netease.MustMain(), model.TickMain(time.Nanosecond)
	}

	// 鼠标事件处理
	if mouse, ok := msg.(tea.MouseMsg); ok {
		// 仅处理左键按下的点击
		if mouse.Button == tea.MouseButtonLeft && mouse.Action == tea.MouseActionPress {
			// 行坐标转换为 1-based 与 View 中记录的行号一致
			y := mouse.Y + 1
			x := mouse.X

			// 点击 Tab 区域
			if y == l.tabsRowY && x >= l.tabStartX && x <= l.tabEndX {
				tabWidth1 := lipgloss.Width(activeTabStyleGetter().Render("手机号/邮箱登录"))

				if x < l.tabStartX+tabWidth1 {
					l.tabIndex = tabAccount
					l.index = idxTabAccount
				} else {
					l.tabIndex = tabCookie
					l.index = idxTabCookie
				}
				l.updateTabStyle()
			}

			if l.tabIndex == tabAccount {
				// 点击输入框：设置焦点
				if y == l.accountRowY {
					l.index = 0
					l.focusAccountInputs()
					return l, tickLogin(time.Nanosecond)
				}
				if y == l.passwordRowY {
					l.index = 1
					l.focusAccountInputs()
					return l, tickLogin(time.Nanosecond)
				}

				// 点击按钮：触发提交或扫码登录
				if y == l.buttonsRowY {
					if x >= l.submitStartX && x <= l.submitEndX {
						l.index = submitIndex
						return l.enterHandler()
					}
					if x >= l.qrStartX && x <= l.qrEndX {
						l.index = qrLoginIndex
						return l.enterHandler()
					}
				}
			} else {
				// Cookie 登录模式
				if y == l.cookieRowY {
					l.index = 0
					l.cookieInput.Focus()
					l.cookieInput.Prompt = model.GetFocusedPrompt()
					l.cookieInput.TextStyle = util.GetPrimaryFontStyle()
					l.submitButton = model.GetBlurredSubmitButton()
					return l, tickLogin(time.Nanosecond)
				}

				// 点击按钮：触发 Cookie 登录
				if y == l.buttonsRowY {
					if x >= l.submitStartX && x <= l.submitEndX {
						l.index = submitIndex
						return l.enterHandler()
					}
				}
			}
		}
		// 其他鼠标事件交给输入框以便光标闪烁等
		if l.tabIndex == tabAccount {
			return l.updateLoginInputs(msg)
		}
		return l.updateCookieInput(msg)
	}

	if key, ok = msg.(tea.KeyMsg); !ok {
		if l.tabIndex == tabAccount {
			return l.updateLoginInputs(msg)
		}
		return l.updateCookieInput(msg)
	}

	switch key.String() {
	case "b":
		if l.index != submitIndex && l.index != qrLoginIndex {
			if l.tabIndex == tabAccount {
				return l.updateLoginInputs(msg)
			}
			return l.updateCookieInput(msg)
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
	case "tab", "shift+tab", "enter", "up", "down", "left", "right", "]", "[":
		s := key.String()

		// Did the user press enter while the submit button was focused?
		// If so, exit.
		if s == "enter" && l.index >= submitIndex {
			return l.enterHandler()
		}

		// 焦点切换：Tab/Shift+Tab/Left/Right 在输入框和按钮间切换
		switch s {
		case "up", "shift+tab":
			switch l.index {
			case idxTabAccount, idxTabCookie:
			case 0:
				if l.tabIndex == tabAccount {
					l.index = idxTabAccount
				} else {
					l.index = idxTabCookie
				}

			case 1:
				l.index = 0

			case submitIndex:
				if l.tabIndex == tabCookie {
					l.index = 0
				} else {
					l.index = 1
				}

			case qrLoginIndex:
				l.index = 1
			}

		case "down", "tab":
			switch l.index {
			case idxTabAccount:
				l.index = 0
			case idxTabCookie:
				l.index = 0

			case 0:
				if l.tabIndex == tabCookie {
					l.index = submitIndex
				} else {
					l.index = 1
				}

			case 1:
				l.index = submitIndex

			case submitIndex, qrLoginIndex:
				if l.tabIndex == tabAccount {
					l.index = idxTabAccount
				} else {
					l.index = idxTabCookie
				}
			}

		case "left":
			switch l.index {
			case idxTabCookie:
				l.index = idxTabAccount
				l.tabIndex = tabAccount
				l.updateTabStyle()

			case qrLoginIndex:
				l.index = submitIndex
			case submitIndex:
				if l.tabIndex == tabCookie {
					l.index = 0
				}
			}

		case "right":
			switch l.index {
			case idxTabAccount:
				l.index = idxTabCookie
				l.tabIndex = tabCookie
				l.updateTabStyle()
			case submitIndex:
				if l.tabIndex == tabAccount {
					l.index = qrLoginIndex
				}
			case 0:
				if l.tabIndex == tabCookie {
					l.index = submitIndex
				}
			}
		}

		// 全部失焦
		if l.index < 0 {
			l.accountInput.Blur()
			l.passwordInput.Blur()
			l.cookieInput.Blur()

			l.accountInput.Prompt = model.GetBlurredPrompt()
			l.accountInput.TextStyle = lipgloss.NewStyle()
			l.passwordInput.Prompt = model.GetBlurredPrompt()
			l.passwordInput.TextStyle = lipgloss.NewStyle()
			l.cookieInput.Prompt = model.GetBlurredPrompt()
			l.cookieInput.TextStyle = lipgloss.NewStyle()

			l.submitButton = model.GetBlurredSubmitButton()
			l.qrLoginButton = model.GetBlurredButton(l.qrButtonTextByStep())

			return l, nil
		}

		if l.tabIndex == tabAccount {
			if l.index > qrLoginIndex {
				l.index = qrLoginIndex
			}

			inputs := []*textinput.Model{&l.accountInput, &l.passwordInput}
			for i := 0; i < len(inputs); i++ {
				if i == l.index {
					inputs[i].Focus()
					inputs[i].Prompt = model.GetFocusedPrompt()
					inputs[i].TextStyle = util.GetPrimaryFontStyle()
				} else {
					inputs[i].Blur()
					inputs[i].Prompt = model.GetBlurredPrompt()
					inputs[i].TextStyle = lipgloss.NewStyle()
				}
			}

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

		} else {
			if l.index == 1 {
				l.index = 0
			}

			switch l.index {
			case 0:
				l.cookieInput.Focus()
				l.cookieInput.Prompt = model.GetFocusedPrompt()
				l.cookieInput.TextStyle = util.GetPrimaryFontStyle()
				l.submitButton = model.GetBlurredSubmitButton()
			case submitIndex:
				l.cookieInput.Blur()
				l.cookieInput.Prompt = model.GetBlurredPrompt()
				l.cookieInput.TextStyle = lipgloss.NewStyle()
				l.submitButton = model.GetFocusedSubmitButton()
			}
		}

		return l, nil
	}

	// Handle character input and blinks
	if l.tabIndex == tabAccount {
		return l.updateLoginInputs(msg)
	}
	return l.updateCookieInput(msg)
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
	write("\n")
	top++

	// 记录 Tab 所在行（1-based）
	l.tabsRowY = curRow()

	write("\n")

	// Tab 渲染
	var accountStyle, cookieStyle lipgloss.Style
	if l.tabIndex == tabAccount {
		accountStyle = activeTabStyleGetter()
		cookieStyle = tabStyle
	} else {
		accountStyle = tabStyle
		cookieStyle = activeTabStyleGetter()
	}

	if l.index == idxTabAccount {
		accountStyle = accountStyle.Copy().Bold(true)
	}
	if l.index == idxTabCookie {
		cookieStyle = cookieStyle.Copy().Bold(true)
	}

	tab1 := accountStyle.Render("手机号/邮箱登录")
	tab2 := cookieStyle.Render("Cookie 登录")

	filledSpace := ""
	if mainPage.MenuStartColumn() > 0 {
		filledSpace = strings.Repeat(" ", mainPage.MenuStartColumn())
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, filledSpace, tab1, tab2)

	write(tabRow)

	// Add blank line between tab bar and form
	write("\n\n")

	// 记录 Tab 区域的起止 X 坐标（0-based）
	l.tabStartX = mainPage.MenuStartColumn()
	if l.tabStartX < 0 {
		l.tabStartX = 0
	}
	l.tabEndX = l.tabStartX + lipgloss.Width(tabRow) - 1

	if l.tabIndex == tabAccount {
		l.renderAccountLoginView(a, &builder, &top, mainPage, write, curRow)
	} else {
		l.renderCookieLoginView(a, &builder, &top, mainPage, write, curRow)
	}

	if a.WindowHeight() > top+3 {
		write(strings.Repeat("\n", a.WindowHeight()-top-3))
	}

	return builder.String()
}

func (l *LoginPage) renderAccountLoginView(a *model.App, builder *strings.Builder, top *int, mainPage *model.Main, write func(string), curRow func() int) {
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

		(*top)++

		if i < len(inputs)-1 {
			write("\n\n")
			(*top)++
		}
	}

	write("\n\n")
	(*top)++
	if mainPage.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	write(l.tips)
	write("\n\n")
	(*top)++
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
}

func (l *LoginPage) renderCookieLoginView(a *model.App, builder *strings.Builder, top *int, mainPage *model.Main, write func(string), curRow func() int) {
	if mainPage.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}

	write(l.cookieInput.View())

	var valueLen int
	if l.cookieInput.Value() == "" {
		valueLen = runewidth.StringWidth(l.cookieInput.Placeholder)
	} else {
		valueLen = runewidth.StringWidth(l.cookieInput.Value())
	}
	if spaceLen := l.netease.WindowWidth() - mainPage.MenuStartColumn() - valueLen - 3; spaceLen > 0 {
		write(strings.Repeat(" ", spaceLen))
	}

	// 记录输入框所在行号（1-based）
	l.cookieRowY = curRow()

	(*top)++

	write("\n\n")
	(*top)++
	if mainPage.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	write(l.tips)
	write("\n\n")
	(*top)++
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

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - lipgloss.Width(l.submitButton)
	if spaceLen > 0 {
		write(strings.Repeat(" ", spaceLen))
	}
	write("\n")
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
		if l.tabIndex == tabCookie {
			return l.loginByCookie()
		}
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

// 登录api返回信息的结构体
type loginResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
}

func (l *LoginPage) loginByAccount() (model.Page, tea.Cmd) {
	var (
		code      float64
		bodyBytes []byte
		err       error
	)

	if strings.ContainsRune(l.accountInput.Value(), '@') {
		loginService := service.LoginEmailService{
			Email:    l.accountInput.Value(),
			Password: l.passwordInput.Value(),
		}
		code, bodyBytes = loginService.LoginEmail()
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
		code, bodyBytes, err = loginService.LoginCellphone()

		if err != nil {
			l.tips = util.SetFgStyle("使用账号密码登录失败："+err.Error(), termenv.ANSIBrightRed)
			slog.Error("使用账号密码登录失败", slogx.Error(err))
			return l, tickLogin(time.Nanosecond)
		}
	}

	var resp loginResponse
	// 尝试解析 body 获取具体的错误信息
	if jsonErr := json.Unmarshal(bodyBytes, &resp); jsonErr == nil {
		if resp.Msg == "" {
			resp.Msg = resp.Message
		}
		if resp.Msg == "" {
			resp.Msg = fmt.Sprintf("未知错误，请稍后再试！code: %d", resp.Code)
		}
	}

	return l, checkLoginCmd(code, resp)
}

func checkLoginCmd(code float64, resp loginResponse) tea.Cmd {
	return func() tea.Msg {
		codeType := _struct.CheckCode(code)
		switch codeType {
		case _struct.UnknownError:
			slog.Error("登录失败, 未知错误", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("未知错误，code: %d", int(code))}
		case _struct.NetworkError:
			slog.Error("登录失败, 网络异常", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("网络异常，请检查后重试")}
		case _struct.TooManyRequests:
			slog.Error("登录失败, 请求过于频繁", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("请求过于频繁，请稍后再试~")}
		case _struct.Success:
			// http状态码200， 但是：
			// 账号密码错误时api状态码为502
			// 请求频繁时api状态码为-462
			// 低版本时api状态码为8821
			// 需要二阶段验证时api状态码为8830
			switch resp.Code {
			case -462:
				slog.Error("登录失败, 请求过于频繁", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("请求过于频繁，请稍后再试~")}
			case 502:
				slog.Error("登录失败, 账号或密码错误", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("账号或密码错误，请重试")}
			case 8821:
				slog.Error("登录失败, 客户端版本过低", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("客户端版本过低，请升级到最新版本后重试")}
			case 8830:
				slog.Error("登录失败, 需要二阶段验证", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("账号需要二阶段验证，目前暂不支持该功能")}
			case 200:
				// 登录成功
				return LoginMsg{err: nil}
			default:
				slog.Error("登录失败, api状态码异常", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("登录失败: %s", resp.Message)}
			}
		default:
			slog.Error("登录失败, 未知错误", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("未知错误, code: %d", int(code))}
		}
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

func (l *LoginPage) updateTabStyle() {
	if l.tabIndex == tabAccount {
		l.menuTitle.Subtitle = "手机号/邮箱登录"
	} else {
		l.menuTitle.Subtitle = "Cookie 登录"
	}
}

func (l *LoginPage) focusAccountInputs() {
	l.accountInput.Focus()
	l.accountInput.Prompt = model.GetFocusedPrompt()
	l.accountInput.TextStyle = util.GetPrimaryFontStyle()

	l.passwordInput.Blur()
	l.passwordInput.Prompt = model.GetBlurredPrompt()
	l.passwordInput.TextStyle = lipgloss.NewStyle()

	l.submitButton = model.GetBlurredSubmitButton()
	l.qrLoginButton = model.GetBlurredButton(l.qrButtonTextByStep())
}

func (l *LoginPage) updateCookieInput(msg tea.Msg) (model.Page, tea.Cmd) {
	var cmd tea.Cmd
	l.cookieInput, cmd = l.cookieInput.Update(msg)
	return l, cmd
}

func checkCookieCmd(cookieStr string) tea.Cmd {
	return func() tea.Msg {
		err := apputils.ParseCookieFromStr(cookieStr, appCookieJar)
		if err != nil {
			return LoginMsg{err: fmt.Errorf("Cookie 格式错误: %w", err)}
		}

		// 正确的写法应该是立即用反序列化的cookie去刷新token
		neteaseutil.SetGlobalCookieJar(appCookieJar)
		jar, err := apputils.RefreshCookieJar()
		if err != nil {
			slog.Error("Cookie 登录失败", slogx.Error(err))
			return LoginMsg{err: fmt.Errorf("Cookie 登录失败: %w", err)}
		}

		slog.Info("使用 Cookie 登录成功")
		appCookieJar = jar
		neteaseutil.SetGlobalCookieJar(appCookieJar)
		err = appCookieJar.Save()
		if err != nil {
			slog.Warn("刷新token成功但保存 Cookie 失败", slogx.Error(err))
		}

		return LoginMsg{err: nil}
	}
}

func (l *LoginPage) loginByCookie() (model.Page, tea.Cmd) {
	cookieStr := l.cookieInput.Value()
	if len(cookieStr) <= 0 {
		l.tips = util.SetFgStyle("请输入 Cookie", termenv.ANSIBrightRed)
		return l, nil
	}

	l.tips = util.SetFgStyle("正在验证 Cookie...", termenv.ANSIBrightCyan)
	l.cookieInput.SetValue("")

	return l, checkCookieCmd(cookieStr)
}
