package ui

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

const LastfmCustomAPIPageType model.PageType = "lastfm_custom_api"

type LastfmCustomAPIPage struct {
	netease *Netease
	svc     *menuServices // service accessor (Phase 3.3.5)

	menuTitle    *model.MenuItem
	index        int
	keyInput     textinput.Model
	secretInput  textinput.Model
	submitButton string
	reloadButton string
	clearButton  string
	tips         string
	AfterAction  func()

	reloadText  string
	clearText   string
	submitIndex int
	reloadIndex int
	clearIndex  int

	backBtnHovered bool
	backBtnRowY    int
	backBtnStartX  int

	keyRowY      int
	secretRowY   int
	inputStartX  int
	inputEndX    int
	buttonsRowY  int
	submitStartX int
	submitEndX   int
	reloadStartX int
	reloadEndX   int
	clearStartX  int
	clearEndX    int

	hoveredInput  int
	hoveredButton int
	mousePointer  string
}

func NewLastfmCustomAPIPage(netease *Netease) *LastfmCustomAPIPage {
	page := newLastfmCustomAPIPage(netease)
	page.reloadAPIAccount()
	return page
}

func newLastfmCustomAPIPage(netease *Netease) *LastfmCustomAPIPage {
	keyInput := textinput.New()
	keyInput.Placeholder = " Key"
	keyInput.CharLimit = 32
	keyInput.SetStyles(pageInputStyles())

	secretInput := textinput.New()
	secretInput.Placeholder = " Secret"
	secretInput.EchoMode = textinput.EchoPassword
	secretInput.EchoCharacter = '•'
	secretInput.CharLimit = 32
	secretInput.SetStyles(pageInputStyles())

	page := &LastfmCustomAPIPage{
		netease:       netease,
		svc:           newMenuServices(netease),
		menuTitle:     &model.MenuItem{Title: "Lastfm API account"},
		keyInput:      keyInput,
		secretInput:   secretInput,
		reloadText:    "重载",
		clearText:     "清空",
		submitIndex:   2,
		reloadIndex:   3,
		clearIndex:    4,
		hoveredInput:  -1,
		hoveredButton: -1,
		mousePointer:  "default",
	}
	page.applyFocus()
	return page
}

func (l *LastfmCustomAPIPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LastfmCustomAPIPage) Type() model.PageType {
	return LastfmCustomAPIPageType
}

func (l *LastfmCustomAPIPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	if mouseMsg, ok := msg.(tea.MouseMotionMsg); ok {
		mouse := mouseMsg.Mouse()
		mainPage := l.netease.MustMain()
		oldBackHovered := l.backBtnHovered
		oldInputHovered := l.hoveredInput
		oldButtonHovered := l.hoveredButton
		oldPointer := l.mousePointer

		breadcrumbChanged, breadcrumbHovered := pageBreadcrumbMotion(a, mainPage, mouse.X, mouse.Y)
		l.backBtnHovered = mouse.Y == l.backBtnRowY && mouse.X >= l.backBtnStartX && mouse.X < l.backBtnStartX+pageBackButtonWidth
		l.hoveredInput = -1
		if mouse.X >= l.inputStartX && mouse.X <= l.inputEndX {
			switch mouse.Y {
			case l.keyRowY:
				l.hoveredInput = 0
			case l.secretRowY:
				l.hoveredInput = 1
			}
		}
		l.hoveredButton = -1
		if mouse.Y == l.buttonsRowY {
			switch {
			case mouse.X >= l.submitStartX && mouse.X <= l.submitEndX:
				l.hoveredButton = 0
			case mouse.X >= l.reloadStartX && mouse.X <= l.reloadEndX:
				l.hoveredButton = 1
			case mouse.X >= l.clearStartX && mouse.X <= l.clearEndX:
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
			return l, tea.Sequence(tickLogin(time.Nanosecond), a.SetMousePointer(l.mousePointer))
		}
		return l.updateAccountInputs(msg)
	}

	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		mouse := clickMsg.Mouse()
		if mouse.Button != tea.MouseLeft {
			return l.updateAccountInputs(msg)
		}
		mainPage := l.netease.MustMain()
		if newPage := pageBreadcrumbClick(a, mainPage, mouse.X, mouse.Y); newPage != nil {
			l.tips = ""
			return newPage, l.netease.RerenderCmd(true)
		}
		if mouse.Y == l.backBtnRowY && mouse.X >= l.backBtnStartX && mouse.X < l.backBtnStartX+pageBackButtonWidth {
			l.tips = ""
			return mainPage, l.netease.RerenderCmd(true)
		}
		if mouse.X >= l.inputStartX && mouse.X <= l.inputEndX {
			switch mouse.Y {
			case l.keyRowY:
				l.index = 0
				l.applyFocus()
				setPageInputCursor(&l.keyInput, mouse.X, l.inputStartX)
				return l, tickLogin(time.Nanosecond)
			case l.secretRowY:
				l.index = 1
				l.applyFocus()
				setPageInputCursor(&l.secretInput, mouse.X, l.inputStartX)
				return l, tickLogin(time.Nanosecond)
			}
		}
		if mouse.Y == l.buttonsRowY {
			switch {
			case mouse.X >= l.submitStartX && mouse.X <= l.submitEndX:
				l.index = l.submitIndex
			case mouse.X >= l.reloadStartX && mouse.X <= l.reloadEndX:
				l.index = l.reloadIndex
			case mouse.X >= l.clearStartX && mouse.X <= l.clearEndX:
				l.index = l.clearIndex
			default:
				return l.updateAccountInputs(msg)
			}
			l.applyFocus()
			return l.enterHandler()
		}
		return l.updateAccountInputs(msg)
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return l.updateAccountInputs(msg)
	}
	switch keyName := key.String(); keyName {
	case "b":
		if l.index < l.submitIndex {
			return l.updateAccountInputs(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		return l.netease.MustMain(), l.netease.RerenderCmd(true)
	case "tab", "shift+tab", "enter", "up", "down", "left", "right":
		if keyName == "enter" && l.index >= l.submitIndex {
			return l.enterHandler()
		}
		switch keyName {
		case "up", "shift+tab":
			l.index--
		case "left", "right":
			if l.index < l.submitIndex {
				return l.updateAccountInputs(msg)
			}
			if keyName == "left" {
				l.index--
			} else {
				l.index++
			}
		default:
			l.index++
		}
		if l.index > l.clearIndex {
			l.index = 0
		} else if l.index < 0 {
			l.index = l.clearIndex
		}
		l.applyFocus()
		return l, nil
	}
	return l.updateAccountInputs(msg)
}

func (l *LastfmCustomAPIPage) View(a *model.App) string {
	var (
		builder  strings.Builder
		top      int
		mainPage = l.netease.MustMain()
	)
	lineCount := 0
	write := func(text string) {
		builder.WriteString(text)
		lineCount += strings.Count(text, "\n")
	}

	if configs.AppConfig.Theme.ShowTitle {
		write(pageTitleView(a, mainPage, &top))
	} else {
		write("\n")
		top++
	}

	topBefore := top
	write(pageMenuTitleViewWithBack(a, mainPage, &top, l.menuTitle, l.backBtnHovered))
	l.backBtnRowY = pageMenuTitleRow(a, mainPage, topBefore)
	l.backBtnStartX = max(0, mainPage.MenuStartColumn()-pageBackButtonWidth)
	write("\n\n")

	l.inputStartX = max(0, mainPage.MenuStartColumn())
	l.inputEndX = max(l.inputStartX, a.WindowWidth()-1)
	inputs := []*textinput.Model{&l.keyInput, &l.secretInput}
	for i, input := range inputs {
		if l.inputStartX > 0 {
			write(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", l.inputStartX)))
		}
		if i == 0 {
			l.keyRowY = lineCount
		} else {
			l.secretRowY = lineCount
		}
		input.SetWidth(max(1, a.WindowWidth()-l.inputStartX-lipgloss.Width(input.Prompt)))
		write(pageInputView(*input, l.hoveredInput == i))
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
		submitButtonView = pageButtonHoverView(pageSubmitText())
	}
	reloadButtonView := l.reloadButton
	if l.hoveredButton == 1 {
		reloadButtonView = pageButtonHoverView(l.reloadText)
	}
	clearButtonView := l.clearButton
	if l.hoveredButton == 2 {
		clearButtonView = pageButtonHoverView(l.clearText)
	}
	buttonGap := "    "
	l.submitStartX = l.inputStartX
	l.submitEndX = l.submitStartX + lipgloss.Width(submitButtonView) - 1
	l.reloadStartX = l.submitEndX + lipgloss.Width(buttonGap) + 1
	l.reloadEndX = l.reloadStartX + lipgloss.Width(reloadButtonView) - 1
	l.clearStartX = l.reloadEndX + lipgloss.Width(buttonGap) + 1
	l.clearEndX = l.clearStartX + lipgloss.Width(clearButtonView) - 1
	write(submitButtonView)
	write(buttonGap)
	write(reloadButtonView)
	write(buttonGap)
	write(clearButtonView)
	if spaceLen := a.WindowWidth() - l.inputStartX - lipgloss.Width(submitButtonView) - lipgloss.Width(reloadButtonView) - lipgloss.Width(clearButtonView) - 2*lipgloss.Width(buttonGap); spaceLen > 0 {
		write(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", spaceLen)))
	}
	write("\n")

	return finishCustomPageView(&builder, a)
}

func (l *LastfmCustomAPIPage) Msg() tea.Msg {
	return nil
}

func (l *LastfmCustomAPIPage) updateAccountInputs(msg tea.Msg) (model.Page, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	l.keyInput, cmd = l.keyInput.Update(msg)
	cmds = append(cmds, cmd)

	l.secretInput, cmd = l.secretInput.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l *LastfmCustomAPIPage) applyFocus() {
	blurPageInput(&l.keyInput)
	blurPageInput(&l.secretInput)
	l.tips = ""
	switch l.index {
	case 0:
		focusPageInput(&l.keyInput)
	case 1:
		focusPageInput(&l.secretInput)
	case l.submitIndex:
		l.tips = style.FG("保存至数据库，优先使用此值", lipgloss.BrightBlue)
	case l.reloadIndex:
		l.tips = style.FG("从数据库或本次启动时的配置文件中加载 API account", lipgloss.BrightBlue)
	case l.clearIndex:
		l.tips = style.FG("清除当前值及已设置值", lipgloss.BrightBlue)
	}
	l.submitButton = pageSubmitButton(l.index == l.submitIndex)
	l.reloadButton = pageButton(l.reloadText, l.index == l.reloadIndex)
	l.clearButton = pageButton(l.clearText, l.index == l.clearIndex)
}

func (l *LastfmCustomAPIPage) enterHandler() (model.Page, tea.Cmd) {
	loading := model.NewLoading(l.netease.MustMain(), l.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case l.submitIndex:
		// 提交
		if len(l.keyInput.Value()) != 32 || len(l.secretInput.Value()) != 32 {
			l.tips = style.FG("请输入正确的 API 账号或密码", lipgloss.BrightRed)
			return l, nil
		}
		l.svc.Lastfm().SetApiAccount(l.keyInput.Value(), l.secretInput.Value())
		l.tips = style.FG("已保存至数据库", lipgloss.BrightGreen)
	case l.reloadIndex:
		l.reloadAPIAccount()
	case l.clearIndex:
		if len(l.keyInput.Value()) != 0 && len(l.secretInput.Value()) != 0 {
			l.keyInput.Reset()
			l.secretInput.Reset()
			l.tips = style.FG("已清空，请重新填写, 为空时再次按下以清除数据库内 Api account", lipgloss.BrightRed)
		} else {
			l.svc.Lastfm().ClearApiAccount()
			l.tips = style.FG("已清除数据库内 Api account，需重新登录", lipgloss.BrightRed)
		}
	}
	if l.AfterAction != nil {
		l.AfterAction()
	}

	return l, tickLogin(time.Nanosecond)
}

func (l *LastfmCustomAPIPage) reloadAPIAccount() (model.Page, tea.Cmd) {
	// var key, secret string
	key, secret := l.svc.Lastfm().GetApiAccount()
	if key != "" && secret != "" {
		l.keyInput.SetValue(key)
		l.secretInput.SetValue(secret)
		l.tips = style.FG("已从已配置值(TUI 设置值)加载", lipgloss.BrightGreen)
	} else if configs.AppConfig.Reporter.Lastfm.Key != "" && configs.AppConfig.Reporter.Lastfm.Secret != "" {
		l.keyInput.SetValue(configs.AppConfig.Reporter.Lastfm.Key)
		l.secretInput.SetValue(configs.AppConfig.Reporter.Lastfm.Secret)
		l.tips = style.FG("已从本次启动时的配置文件中加载", lipgloss.BrightGreen)
	} else {
		l.keyInput.Reset()
		l.secretInput.Reset()
		l.tips = style.FG("未获取到内容，已重置", lipgloss.BrightGreen)
	}

	return l, nil
}
