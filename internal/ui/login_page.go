package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/netease-music/service"
	neteaseutil "github.com/go-musicfox/netease-music/util"

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

	idxTabAccount = -2 // 账号登录 Tab 的焦点索引
	idxTabCookie  = -1 // Cookie 登录 Tab 的焦点索引
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
	tabs          *model.Tabs
	accountInput  textinput.Model
	passwordInput textinput.Model
	cookieInput   textinput.Model
	submitButton  string
	qrLoginButton string
	qrLoginStep   int
	tips          string
	AfterLogin    LoginCallback

	// 以下字段用于鼠标点击区域的计算与命中
	accountRowY   int // 账号输入框所在的行号（1-based）
	passwordRowY  int // 密码输入框所在的行号（1-based）
	cookieRowY    int // Cookie输入框所在的行号（1-based）
	inputStartX   int // 输入框起始 X（0-based）
	inputEndX     int // 输入框结束 X（0-based，闭区间）
	buttonsRowY   int // 提交/扫码按钮所在行号（1-based）
	submitStartX  int // 提交按钮起始 X（0-based）
	submitEndX    int // 提交按钮结束 X（0-based，闭区间）
	qrStartX      int // 扫码按钮起始 X（0-based）
	qrEndX        int // 扫码按钮结束 X（0-based，闭区间）
	cookieStartX  int // Cookie按钮起始 X（0-based）
	cookieEndX    int // Cookie按钮结束 X（0-based，闭区间）
	tabStartX     int // Tabs 起始 X（0-based）
	tabsStartRowY int // Tabs 起始行（1-based）
	tabsEndRowY   int // Tabs 结束行（1-based）
	backBtnRowY   int // 返回按钮所在行号（1-based）
	backBtnStartX int
	backBtnEndX   int

	// Hover 状态跟踪
	backBtnHovered  bool
	hoveredTab      int // 悬停的 Tab 索引（0=账号登录，1=Cookie，-1=无）
	hoveredInputBox int // 悬停的输入框（0=account，1=password，2=cookie，-1=无）
	hoveredButton   int // 悬停的按钮（0=提交，1=扫码，-1=无）
	mousePointer    string
}

// 执行登录操作回显的信息结构体
type LoginMsg struct {
	err error
}

func NewLoginPage(netease *Netease) (login *LoginPage) {
	accountInput := textinput.New()
	accountInput.Placeholder = model.T(MsgLoginAccountPlaceholder)
	accountInput.CharLimit = 32

	passwordInput := textinput.New()
	passwordInput.Placeholder = model.T(MsgLoginPasswordPlaceholder)
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 32

	cookieInput := textinput.New()
	cookieInput.Placeholder = model.T(MsgLoginCookiePlaceholder)
	cookieInput.CharLimit = 5000

	login = &LoginPage{
		netease:         netease,
		menuTitle:       &model.MenuItem{Title: model.T(MsgLoginPageTitle)},
		tabs:            model.NewTabs([]string{model.T(MsgLoginAccountTab), model.T(MsgLoginCookieTab)}),
		accountInput:    accountInput,
		passwordInput:   passwordInput,
		cookieInput:     cookieInput,
		submitButton:    pageSubmitButton(false),
		qrLoginButton:   pageButton(model.T(MsgLoginQRCodeButton), false),
		hoveredTab:      -1,
		hoveredInputBox: -1,
		hoveredButton:   -1,
		mousePointer:    "default",
	}
	login.updateTabStyle()
	login.applyFocus()
	return login
}

func (l *LoginPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LoginPage) Type() model.PageType {
	return LoginPageType
}

func (l *LoginPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	if _, ok := msg.(tickLoginMsg); ok {
		return l, nil
	}

	if loginMsg, ok := msg.(LoginMsg); ok {
		if loginMsg.err != nil {
			l.tips = util.SetFgStyle(loginMsg.err.Error(), lipgloss.BrightRed)
			return l, nil
		}
		if newPage := l.loginSuccessHandle(l.netease); newPage != nil {
			return newPage, tea.Batch(tea.ClearScreen, model.TickMain(time.Nanosecond), l.netease.RerenderCmd(true))
		}
		return l.netease.MustMain(), model.TickMain(time.Nanosecond)
	}

	if mouseMsg, ok := msg.(tea.MouseMotionMsg); ok {
		mouse := mouseMsg.Mouse()
		y := mouse.Y + 1
		x := mouse.X
		oldBackButton := l.backBtnHovered
		oldTab := l.hoveredTab
		oldInput := l.hoveredInputBox
		oldButton := l.hoveredButton
		oldPointer := l.mousePointer

		// Delegate breadcrumb hover (use raw 0-based mouse.Y, not 1-based y)
		bcChanged, bcOver := pageBreadcrumbMotion(a, l.netease.MustMain(), mouse.X, mouse.Y)

		l.backBtnHovered = y == l.backBtnRowY && x >= l.backBtnStartX && x < l.backBtnEndX
		l.hoveredTab = l.tabAt(x, y)
		l.tabs.SetHovered(l.hoveredTab)
		l.hoveredInputBox = -1
		l.hoveredButton = -1
		if x >= l.inputStartX && x <= l.inputEndX {
			if l.activeTab() == tabAccount {
				switch y {
				case l.accountRowY:
					l.hoveredInputBox = 0
				case l.passwordRowY:
					l.hoveredInputBox = 1
				}
			} else if y == l.cookieRowY {
				l.hoveredInputBox = 2
			}
		}
		if y == l.buttonsRowY {
			if x >= l.submitStartX && x <= l.submitEndX {
				l.hoveredButton = 0
			} else if l.activeTab() == tabAccount && x >= l.qrStartX && x <= l.qrEndX {
				l.hoveredButton = 1
			}
		}

		l.mousePointer = "default"
		if l.hoveredInputBox >= 0 {
			l.mousePointer = "text"
		} else if l.backBtnHovered || l.hoveredTab >= 0 || l.hoveredButton >= 0 || bcOver {
			l.mousePointer = "pointer"
		}
		if l.backBtnHovered != oldBackButton || l.hoveredTab != oldTab || l.hoveredInputBox != oldInput || l.hoveredButton != oldButton || l.mousePointer != oldPointer || bcChanged {
			return l, tea.Sequence(tickLogin(time.Nanosecond), a.SetMousePointer(l.mousePointer))
		}
		return l.updateActiveInput(msg)
	}

	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		mouse := clickMsg.Mouse()
		if mouse.Button != tea.MouseLeft {
			return l.updateActiveInput(msg)
		}
		// Breadcrumb click delegates to Main (use raw 0-based mouse.Y)
		if newPage := pageBreadcrumbClick(a, l.netease.MustMain(), mouse.X, mouse.Y); newPage != nil {
			l.tips = ""
			l.qrLoginStep = 0
			l.qrLoginButton = pageButton(l.qrButtonTextByStep(), false)
			return newPage, l.netease.RerenderCmd(true)
		}
		y := mouse.Y + 1
		x := mouse.X
		if y == l.backBtnRowY && x >= l.backBtnStartX && x < l.backBtnEndX {
			l.tips = ""
			l.qrLoginStep = 0
			l.qrLoginButton = pageButton(l.qrButtonTextByStep(), false)
			return l.netease.MustMain(), l.netease.RerenderCmd(true)
		}

		if tab := l.tabAt(x, y); tab >= 0 {
			l.tabs.SetActive(tab)
			l.index = l.tabFocusIndex()
			l.updateTabStyle()
			l.applyFocus()
			return l, tickLogin(time.Nanosecond)
		}

		if l.activeTab() == tabAccount {
			if x >= l.inputStartX && x <= l.inputEndX {
				switch y {
				case l.accountRowY:
					l.index = 0
					l.applyFocus()
					setPageInputCursor(&l.accountInput, x, l.inputStartX)
					return l, tickLogin(time.Nanosecond)
				case l.passwordRowY:
					l.index = 1
					l.applyFocus()
					setPageInputCursor(&l.passwordInput, x, l.inputStartX)
					return l, tickLogin(time.Nanosecond)
				}
			}
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
			if y == l.cookieRowY && x >= l.inputStartX && x <= l.inputEndX {
				l.index = 0
				l.applyFocus()
				setPageInputCursor(&l.cookieInput, x, l.inputStartX)
				return l, tickLogin(time.Nanosecond)
			}
			if y == l.buttonsRowY && x >= l.submitStartX && x <= l.submitEndX {
				l.index = submitIndex
				return l.enterHandler()
			}
		}
		return l.updateActiveInput(msg)
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return l.updateActiveInput(msg)
	}
	if l.index < 0 && l.updateTabs(msg, key.String()) {
		return l, nil
	}
	keyName := key.String()
	if l.index == 0 || (l.activeTab() == tabAccount && l.index == 1) {
		switch keyName {
		case "left", "right", "[", "]":
			return l.updateActiveInput(msg)
		}
	}

	switch keyName {
	case "b":
		if l.index != submitIndex && l.index != qrLoginIndex {
			return l.updateActiveInput(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		l.qrLoginStep = 0
		l.qrLoginButton = pageButton(l.qrButtonTextByStep(), l.index == qrLoginIndex)
		return l.netease.MustMain(), l.netease.RerenderCmd(true)
	case "tab", "shift+tab", "enter", "up", "down", "left", "right":
		if keyName == "enter" && l.index >= submitIndex {
			return l.enterHandler()
		}

		switch keyName {
		case "up", "shift+tab":
			switch l.index {
			case idxTabAccount, idxTabCookie:
			case 0:
				l.index = l.tabFocusIndex()
			case 1:
				l.index = 0
			case submitIndex:
				if l.activeTab() == tabCookie {
					l.index = 0
				} else {
					l.index = 1
				}
			case qrLoginIndex:
				l.index = 1
			}
		case "down", "tab", "enter":
			switch l.index {
			case idxTabAccount, idxTabCookie:
				l.index = 0
			case 0:
				if l.activeTab() == tabCookie {
					l.index = submitIndex
				} else {
					l.index = 1
				}
			case 1:
				l.index = submitIndex
			case submitIndex, qrLoginIndex:
				l.index = l.tabFocusIndex()
			}
		case "left":
			switch l.index {
			case qrLoginIndex:
				l.index = submitIndex
			case submitIndex:
				if l.activeTab() == tabCookie {
					l.index = 0
				}
			}
		case "right":
			switch l.index {
			case submitIndex:
				if l.activeTab() == tabAccount {
					l.index = qrLoginIndex
				}
			case 0:
				if l.activeTab() == tabCookie {
					l.index = submitIndex
				}
			}
		}
		l.applyFocus()
		return l, nil
	}

	return l.updateActiveInput(msg)
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
		write(pageTitleView(a, mainPage, &top))
	} else {
		write("\n")
		top++
	}

	// menu title
	topBefore := top
	write(pageMenuTitleViewWithBack(a, mainPage, &top, l.menuTitle, l.backBtnHovered))
	l.backBtnRowY = pageMenuTitleRow(a, mainPage, topBefore) + 1
	l.backBtnStartX = max(0, mainPage.MenuStartColumn()-pageBackButtonWidth)
	l.backBtnEndX = l.backBtnStartX + pageBackButtonWidth
	write("\n")
	top++

	write("\n")
	l.tabStartX = max(0, mainPage.MenuStartColumn())
	l.tabsStartRowY = curRow()
	l.tabs.SetSize(max(0, a.WindowWidth()-l.tabStartX), 3)
	l.tabs.SetHovered(l.hoveredTab)
	tabRow := l.tabs.View()
	if l.tabStartX > 0 {
		tabRow = lipgloss.NewStyle().PaddingLeft(l.tabStartX).Render(tabRow)
	}
	write(tabRow)
	l.tabsEndRowY = l.tabsStartRowY + lipgloss.Height(tabRow) - 1
	write("\n\n")

	l.inputStartX = max(0, mainPage.MenuStartColumn())
	l.inputEndX = max(l.inputStartX, a.WindowWidth()-1)
	if l.activeTab() == tabAccount {
		l.renderAccountLoginView(a, &builder, &top, mainPage, write, curRow)
	} else {
		l.renderCookieLoginView(a, &builder, &top, mainPage, write, curRow)
	}

	fillPageHeight(&builder, a.WindowHeight())

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

		input.SetWidth(max(1, a.WindowWidth()-l.inputStartX-lipgloss.Width(input.Prompt)))
		write(pageInputView(*input, l.hoveredInputBox == i))

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
	submitButtonView := l.submitButton
	if l.hoveredButton == 0 {
		submitButtonView = pageButtonHoverView(pageSubmitText())
	}
	qrButtonView := l.qrLoginButton
	if l.hoveredButton == 1 {
		qrButtonView = pageButtonHoverView(l.qrButtonTextByStep())
	}

	submitW := lipgloss.Width(submitButtonView)
	l.submitStartX = submitX
	l.submitEndX = submitX + submitW - 1

	write(submitButtonView)

	btnBlank := "    "
	write(btnBlank)
	// 扫码按钮坐标
	qrX := submitX + submitW + lipgloss.Width(btnBlank)
	qrW := lipgloss.Width(qrButtonView)
	l.qrStartX = qrX
	l.qrEndX = qrX + qrW - 1

	write(qrButtonView)

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - lipgloss.Width(submitButtonView) - lipgloss.Width(qrButtonView) - lipgloss.Width(btnBlank)
	if spaceLen > 0 {
		write(strings.Repeat(" ", spaceLen))
	}
	write("\n")
}

func (l *LoginPage) renderCookieLoginView(a *model.App, builder *strings.Builder, top *int, mainPage *model.Main, write func(string), curRow func() int) {
	if mainPage.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}

	l.cookieInput.SetWidth(max(1, a.WindowWidth()-l.inputStartX-lipgloss.Width(l.cookieInput.Prompt)))
	write(pageInputView(l.cookieInput, l.hoveredInputBox == 2))

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
	submitButtonView := l.submitButton
	if l.hoveredButton == 0 {
		submitButtonView = pageButtonHoverView(pageSubmitText())
	}

	submitW := lipgloss.Width(submitButtonView)
	l.submitStartX = submitX
	l.submitEndX = submitX + submitW - 1

	write(submitButtonView)

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - lipgloss.Width(submitButtonView)
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
		return model.T(MsgLoginQRCodeContinue)
	case 0:
		fallthrough
	default:
		return model.T(MsgLoginQRCodeButton)
	}
}

func (l *LoginPage) enterHandler() (model.Page, tea.Cmd) {
	loading := model.NewLoading(l.netease.MustMain(), l.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case submitIndex:
		if l.activeTab() == tabCookie {
			return l.loginByCookie()
		}
		// 提交
		if len(l.accountInput.Value()) <= 0 || len(l.passwordInput.Value()) <= 0 {
			l.tips = util.SetFgStyle(model.T(MsgLoginCredentialRequired), lipgloss.BrightRed)
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
			l.tips = util.SetFgStyle(model.T(MsgLoginFailed)+err.Error(), lipgloss.BrightRed)
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
			resp.Msg = fmt.Sprintf(model.T(MsgLoginUnknownError), resp.Code)
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
			return LoginMsg{err: fmt.Errorf("%s", fmt.Sprintf(model.T(MsgLoginUnknownError), int(code)))}
		case _struct.NetworkError:
			slog.Error("登录失败, 网络异常", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("%s", model.T(MsgLoginNetworkError))}
		case _struct.TooManyRequests:
			slog.Error("登录失败, 请求过于频繁", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("%s", model.T(MsgLoginTooManyRequests))}
		case _struct.Success:
			// http状态码200， 但是：
			// 账号密码错误时api状态码为502
			// 请求频繁时api状态码为-462
			// 低版本时api状态码为8821
			// 需要二阶段验证时api状态码为8830
			switch resp.Code {
			case -462:
				slog.Error("登录失败, 请求过于频繁", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("%s", model.T(MsgLoginTooManyRequests))}
			case 502:
				slog.Error("登录失败, 账号或密码错误", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("%s", model.T(MsgLoginInvalidCredentials))}
			case 8821:
				slog.Error("登录失败, 客户端版本过低", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("%s", model.T(MsgLoginVersionTooOld))}
			case 8830:
				slog.Error("登录失败, 需要二阶段验证", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("%s", model.T(MsgLogin2FANotSupported))}
			case 200:
				// 登录成功
				return LoginMsg{err: nil}
			default:
				slog.Error("登录失败, api状态码异常", slogx.Error(resp.Message))
				return LoginMsg{err: fmt.Errorf("登录失败: %s", resp.Message)}
			}
		default:
			slog.Error("登录失败, 未知错误", slogx.Error(resp.Message))
			return LoginMsg{err: fmt.Errorf("%s", fmt.Sprintf(model.T(MsgLoginUnknownError), int(code)))}
		}
	}
}

// loginByQRCode 跳转到二维码登录界面
func (l *LoginPage) loginByQRCode() (model.Page, tea.Cmd) {
	qrPage := NewQRLoginPage(l.netease, l, l.AfterLogin)
	return qrPage, qrPage.Init()
}

func (l *LoginPage) loginSuccessHandle(n *Netease) model.Page {
	// 先保存 cookie，确保登录成功后 cookie 被持久化
	// 即使后续 LoginCallback 失败（AccountInfo 失败），cookie 也已保存
	if err := appCookieJar.Save(); err != nil {
		slog.Warn("持久化 Cookie 失败", slogx.Error(err))
	}

	if err := n.LoginCallback(); err != nil {
		slog.Error("login callback error", slogx.Error(err))
	}

	var newPage model.Page
	if l.AfterLogin != nil {
		newPage = l.AfterLogin()
	}
	return newPage
}

func (l *LoginPage) activeTab() int {
	return l.tabs.Active()
}

func (l *LoginPage) tabFocusIndex() int {
	return idxTabAccount + l.activeTab()
}

func (l *LoginPage) tabAt(x, y int) int {
	if y < l.tabsStartRowY || y > l.tabsEndRowY || x < l.tabStartX {
		return -1
	}
	offset := x - l.tabStartX
	accountWidth := lipgloss.Width(model.T(MsgLoginAccountTab)) + 4
	if offset < accountWidth {
		return tabAccount
	}
	if offset < accountWidth+lipgloss.Width(model.T(MsgLoginCookieTab))+4 {
		return tabCookie
	}
	return -1
}

func (l *LoginPage) updateTabs(msg tea.Msg, key string) bool {
	switch key {
	case "left", "right", "h", "l":
		l.tabs.Update(msg)
	default:
		return false
	}
	l.index = l.tabFocusIndex()
	l.updateTabStyle()
	l.applyFocus()
	return true
}

func (l *LoginPage) updateTabStyle() {
	if l.activeTab() == tabAccount {
		l.menuTitle.Subtitle = model.T(MsgLoginAccountTab)
	} else {
		l.menuTitle.Subtitle = model.T(MsgLoginCookieTab)
	}
}

func (l *LoginPage) applyFocus() {
	blurPageInput(&l.accountInput)
	blurPageInput(&l.passwordInput)
	blurPageInput(&l.cookieInput)
	if l.index < 0 {
		l.index = l.tabFocusIndex()
		l.tabs.Focus()
	} else {
		l.tabs.Blur()
		if l.activeTab() == tabAccount {
			switch l.index {
			case 0:
				focusPageInput(&l.accountInput)
			case 1:
				focusPageInput(&l.passwordInput)
			}
		} else if l.index == 0 {
			focusPageInput(&l.cookieInput)
		}
	}
	l.submitButton = pageSubmitButton(l.index == submitIndex)
	l.qrLoginButton = pageButton(l.qrButtonTextByStep(), l.index == qrLoginIndex)
}

func (l *LoginPage) updateActiveInput(msg tea.Msg) (model.Page, tea.Cmd) {
	if l.activeTab() == tabAccount {
		return l.updateLoginInputs(msg)
	}
	return l.updateCookieInput(msg)
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
			return LoginMsg{err: fmt.Errorf(model.T(MsgLoginCookieInvalid), err)}
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
		l.tips = util.SetFgStyle(model.T(MsgLoginCookieRequired), lipgloss.BrightRed)
		return l, nil
	}

	l.tips = util.SetFgStyle(model.T(MsgLoginCookieVerifying), lipgloss.BrightCyan)
	l.cookieInput.SetValue("")

	return l, checkCookieCmd(cookieStr)
}
