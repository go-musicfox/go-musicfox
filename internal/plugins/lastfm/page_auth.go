package lastfm

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	lastfm "github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

const LastfmAuthPageType model.PageType = "lastfm_auth"

type LastfmAuthPage struct {
	svc ui.MenuServices // service accessor (Phase 3.9 plugin boundary)

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

	// 鼠标交互
	backBtnHovered bool
	backBtnRowY    int
	backBtnStartX  int
	backBtnEndX    int

	accountRowY   int
	passwordRowY  int
	inputStartX   int
	inputEndX     int
	buttonsRowY   int
	submitStartX  int
	submitEndX    int
	qrAuthStartX  int
	qrAuthEndX    int
	browserStartX int
	browserEndX   int

	hoveredInput  int
	hoveredButton int
	mousePointer  string
}

func NewLastfmAuthPage(svc ui.MenuServices) *LastfmAuthPage {
	accountInput := textinput.New()
	accountInput.Placeholder = " 用户名或邮箱"
	accountInput.CharLimit = 32
	accountInput.SetStyles(ui.PageInputStyles())

	passwordInput := textinput.New()
	passwordInput.Placeholder = " 密码"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 32
	passwordInput.SetStyles(ui.PageInputStyles())

	page := &LastfmAuthPage{
		svc:           svc,
		menuTitle:     &model.MenuItem{Title: "Lastfm用户登录/授权"},
		accountInput:  accountInput,
		passwordInput: passwordInput,

		submitIndex:  2,
		qrAuthIndex:  3,
		browserIndex: 4,

		hoveredInput:  -1,
		hoveredButton: -1,
		mousePointer:  "default",
	}
	page.applyFocus()
	return page
}

func (l *LastfmAuthPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LastfmAuthPage) Type() model.PageType {
	return LastfmAuthPageType
}

func (l *LastfmAuthPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	// 处理鼠标移动事件
	if mouseMsg, ok := msg.(tea.MouseMotionMsg); ok {
		mouse := mouseMsg.Mouse()
		mainPage := l.svc.MustMain()
		oldBackHovered := l.backBtnHovered
		oldInputHovered := l.hoveredInput
		oldButtonHovered := l.hoveredButton
		oldPointer := l.mousePointer

		breadcrumbChanged, breadcrumbHovered := ui.PageBreadcrumbMotion(a, mainPage, mouse.X, mouse.Y)
		l.backBtnHovered = mouse.Y == l.backBtnRowY && mouse.X >= l.backBtnStartX && mouse.X < l.backBtnEndX
		l.hoveredInput = -1
		if mouse.X >= l.inputStartX && mouse.X <= l.inputEndX {
			switch mouse.Y {
			case l.accountRowY:
				l.hoveredInput = 0
			case l.passwordRowY:
				l.hoveredInput = 1
			}
		}
		l.hoveredButton = -1
		if mouse.Y == l.buttonsRowY {
			switch {
			case mouse.X >= l.submitStartX && mouse.X <= l.submitEndX:
				l.hoveredButton = 0
			case mouse.X >= l.qrAuthStartX && mouse.X <= l.qrAuthEndX:
				l.hoveredButton = 1
			case mouse.X >= l.browserStartX && mouse.X <= l.browserEndX:
				l.hoveredButton = 2
			}
		}

		l.mousePointer = "default"
		if l.hoveredInput >= 0 {
			l.mousePointer = "text"
		} else if l.backBtnHovered || l.hoveredButton >= 0 || breadcrumbHovered {
			l.mousePointer = "pointer"
		}
		if l.backBtnHovered != oldBackHovered || l.hoveredInput != oldInputHovered || l.hoveredButton != oldButtonHovered || l.mousePointer != oldPointer || breadcrumbChanged {
			return l, tea.Sequence(ui.TickLogin(time.Nanosecond), a.SetMousePointer(l.mousePointer))
		}
		return l.updateLoginInputs(msg)
	}

	// 处理鼠标点击事件
	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		mouse := clickMsg.Mouse()
		if mouse.Button != tea.MouseLeft {
			return l.updateLoginInputs(msg)
		}
		mainPage := l.svc.MustMain()
		if newPage := ui.PageBreadcrumbClick(a, mainPage, mouse.X, mouse.Y); newPage != nil {
			l.tips = ""
			return newPage, l.svc.App().RerenderCmd(true)
		}
		if mouse.Y == l.backBtnRowY && mouse.X >= l.backBtnStartX && mouse.X < l.backBtnEndX {
			l.tips = ""
			return mainPage, l.svc.App().RerenderCmd(true)
		}
		if mouse.X >= l.inputStartX && mouse.X <= l.inputEndX {
			switch mouse.Y {
			case l.accountRowY:
				l.index = 0
				l.applyFocus()
				ui.SetPageInputCursor(&l.accountInput, mouse.X, l.inputStartX)
				return l, ui.TickLogin(time.Nanosecond)
			case l.passwordRowY:
				l.index = 1
				l.applyFocus()
				ui.SetPageInputCursor(&l.passwordInput, mouse.X, l.inputStartX)
				return l, ui.TickLogin(time.Nanosecond)
			}
		}
		if mouse.Y == l.buttonsRowY {
			switch {
			case mouse.X >= l.submitStartX && mouse.X <= l.submitEndX:
				l.index = l.submitIndex
			case mouse.X >= l.qrAuthStartX && mouse.X <= l.qrAuthEndX:
				l.index = l.qrAuthIndex
			case mouse.X >= l.browserStartX && mouse.X <= l.browserEndX:
				l.index = l.browserIndex
			default:
				return l.updateLoginInputs(msg)
			}
			l.applyFocus()
			return l.enterHandler()
		}
		return l.updateLoginInputs(msg)
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
		if l.index < l.submitIndex {
			return l.updateLoginInputs(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		return l.svc.MustMain(), l.svc.App().RerenderCmd(true)
	case "tab", "shift+tab", "enter", "up", "down", "left", "right":
		s := key.String()

		// Did the user press enter while the submit button was focused?
		// If so, exit.
		if s == "enter" && l.index >= l.submitIndex {
			return l.enterHandler()
		}

		// 当focus在button上时，左右按键的特殊处理
		switch s {
		case "left", "right":
			if l.index < l.submitIndex {
				return l.updateLoginInputs(msg)
			}
			if s == "left" && l.index > l.submitIndex {
				l.index--
			} else if s == "right" && l.index < l.browserIndex {
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

		l.applyFocus()
		return l, nil
	}

	// Handle character input and blinks
	return l.updateLoginInputs(msg)
}

func (l *LastfmAuthPage) View(a *model.App) string {
	var (
		builder  strings.Builder
		top      int
		mainPage = l.svc.MustMain()
	)

	lineCount := 0
	write := func(text string) {
		builder.WriteString(text)
		lineCount += strings.Count(text, "\n")
	}

	// title
	if configs.AppConfig.Theme.ShowTitle {
		write(ui.PageTitleView(a, mainPage, &top))
	} else {
		write("\n")
		top++
	}

	// menu title
	topBefore := top
	write(ui.PageMenuTitleViewWithBack(a, mainPage, &top, l.menuTitle, l.backBtnHovered))
	l.backBtnRowY = ui.PageMenuTitleRow(a, mainPage, topBefore)
	l.backBtnStartX = max(0, mainPage.MenuStartColumn()-ui.PageBackButtonWidth)
	l.backBtnEndX = l.backBtnStartX + ui.PageBackButtonWidth
	write("\n\n")

	l.inputStartX = max(0, mainPage.MenuStartColumn())
	l.inputEndX = max(l.inputStartX, a.WindowWidth()-1)

	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	for i, input := range inputs {
		if l.inputStartX > 0 {
			write(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", l.inputStartX)))
		}
		if i == 0 {
			l.accountRowY = lineCount
		} else {
			l.passwordRowY = lineCount
		}
		input.SetWidth(max(1, a.WindowWidth()-l.inputStartX-lipgloss.Width(input.Prompt)))
		write(ui.PageInputView(*input, l.hoveredInput == i))

		if i < len(inputs)-1 {
			write("\n\n")
		}
	}

	write("\n\n")
	if l.inputStartX > 0 {
		write(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", l.inputStartX)))
	}
	write(l.tips)
	write("\n\n")
	if l.inputStartX > 0 {
		write(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", l.inputStartX)))
	}

	l.buttonsRowY = lineCount
	submitButtonView := l.submitButton
	if l.hoveredButton == 0 {
		submitButtonView = ui.PageButtonHoverView(ui.PageSubmitText())
	}
	qrAuthButtonView := l.qrAuthButton
	if l.hoveredButton == 1 {
		qrAuthButtonView = ui.PageButtonHoverView(l.qrButtonTextByStep())
	}
	browserButtonView := l.browserButton
	if l.hoveredButton == 2 {
		browserButtonView = ui.PageButtonHoverView(l.browserButtonTextByStep())
	}

	buttonGap := "    "
	l.submitStartX = l.inputStartX
	l.submitEndX = l.submitStartX + lipgloss.Width(submitButtonView) - 1
	l.qrAuthStartX = l.submitEndX + lipgloss.Width(buttonGap) + 1
	l.qrAuthEndX = l.qrAuthStartX + lipgloss.Width(qrAuthButtonView) - 1
	l.browserStartX = l.qrAuthEndX + lipgloss.Width(buttonGap) + 1
	l.browserEndX = l.browserStartX + lipgloss.Width(browserButtonView) - 1

	write(submitButtonView)
	write(buttonGap)
	write(qrAuthButtonView)
	write(buttonGap)
	write(browserButtonView)

	spaceLen := a.WindowWidth() - l.inputStartX - lipgloss.Width(submitButtonView) - lipgloss.Width(qrAuthButtonView) - lipgloss.Width(browserButtonView) - 2*lipgloss.Width(buttonGap)
	if spaceLen > 0 {
		write(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", spaceLen)))
	}
	write("\n")

	return ui.FinishCustomPageView(&builder, a)
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

func (l *LastfmAuthPage) applyFocus() {
	ui.BlurPageInput(&l.accountInput)
	ui.BlurPageInput(&l.passwordInput)
	l.tips = ""

	switch l.index {
	case 0:
		ui.FocusPageInput(&l.accountInput)
	case 1:
		ui.FocusPageInput(&l.passwordInput)
	case l.submitIndex:
		l.tips = style.FG("使用账号密码登录并授权", lipgloss.BrightBlue)
	case l.qrAuthIndex:
		l.tips = style.FG("请使用可扫码设备扫码并在浏览器授权", lipgloss.BrightBlue)
	case l.browserIndex:
		l.tips = style.FG("在默认浏览器中打开链接并授权", lipgloss.BrightBlue)
	}

	l.submitButton = ui.PageSubmitButton(l.index == l.submitIndex)
	l.qrAuthButton = ui.PageButton(l.qrButtonTextByStep(), l.index == l.qrAuthIndex)
	l.browserButton = ui.PageButton(l.browserButtonTextByStep(), l.index == l.browserIndex)
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
	loading := model.NewLoading(l.svc.MustMain(), l.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case l.submitIndex:
		// 提交
		// 简单的账号密码判断
		if len(l.accountInput.Value()) < 2 || len(l.accountInput.Value()) > 15 || len(l.passwordInput.Value()) < 6 {
			l.tips = style.FG("请正确输入账号或密码", lipgloss.BrightRed)
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

	return l, ui.TickLogin(time.Nanosecond)
}

func (l *LastfmAuthPage) getAuthURLWithToken() bool {
	if l.token != "" && l.url != "" {
		return true
	}

	if !lastfm.IsAvailable() {
		l.tips = style.FG("请确保正确设置 API key 及 secret", lipgloss.BrightRed)
		return false
	}

	var err error
	if l.token, l.url, err = l.svc.Lastfm().GetAuthUrlWithToken(); err != nil {
		slog.Info("token", slog.Any("token", l.token))
		slog.Info("url", slog.Any("url", l.url))
		l.tips = style.FG("token 或 url 获取失败", lipgloss.BrightRed)
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
	if l.sessionKey, err = l.svc.Lastfm().GetSession(l.token); err != nil {
		l.tips = style.FG("sessionKey 获取失败", lipgloss.BrightRed)
		slog.Error("sessionKey 获取失败", slogx.Error(err))
		return false
	}
	return true
}

func (l *LastfmAuthPage) initUserInfo() bool {
	user, err := l.svc.Lastfm().GetUserInfo(map[string]any{})
	if err != nil {
		l.tips = style.FG("用户信息获取失败", lipgloss.BrightRed)
		slog.Error("用户信息获取失败", slogx.Error(err))
		return false
	}

	l.svc.Lastfm().InitUserInfo(&storage.LastfmUser{
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
	l.sessionKey, err = l.svc.Lastfm().Login(l.accountInput.Value(), l.passwordInput.Value())
	if err != nil {
		l.tips = style.FG("登录失败，请检查", lipgloss.BrightRed)
		slog.Error("登录失败", slogx.Error(err))
		return l, nil
	}

	if !l.initUserInfo() {
		return l, nil
	}
	return l.authSuccessHandle()
}

func (l *LastfmAuthPage) authByQRCode() (model.Page, tea.Cmd) {
	qrPage := NewLastfmQRAuthPage(l.svc, l, l.AfterAction)
	return qrPage, qrPage.Init()
}

func (l *LastfmAuthPage) authByBrower() (model.Page, tea.Cmd) {
	if l.browserAuthStep == 0 {
		if !l.getAuthURLWithToken() {
			return l, nil
		}

		if err := open.Start(l.url); err != nil {
			l.tips = style.FG("认证页打开失败，请确认浏览器是否工作", lipgloss.BrightRed)
			slog.Error("认证页打开失败", slogx.Error(err))
			return l, nil
		}
		l.tips = style.FG("请在浏览器中授权后继续，若未正确跳转，请更换认证方式", lipgloss.BrightBlue)
		l.browserAuthStep++
		l.browserButton = util.GetFocusedButton(l.browserButtonTextByStep())
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
		Text:    fmt.Sprintf("Last.fm 用户 %s 授权成功", l.svc.Lastfm().UserName()),
		GroupId: types.GroupID,
	})
	return l.svc.MustMain(), model.TickMain(time.Second)
}
