package ui

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/mattn/go-runewidth"
)

const LastfmCustomApiPageType model.PageType = "lastfm_custom_api"

type LastfmCustomApiPage struct {
	netease *Netease

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
}

func NewLastfmCustomApiPage(netease *Netease) *LastfmCustomApiPage {
	keyInput := textinput.New()
	keyInput.Placeholder = " Key"
	keyInput.Focus()
	keyInput.Prompt = model.GetFocusedPrompt()
	s := textinput.DefaultStyles(true)
	s.Focused.Text = util.GetPrimaryFontStyle()
	keyInput.SetStyles(s)
	keyInput.CharLimit = 32

	secretInput := textinput.New()
	secretInput.Placeholder = " Secret"
	secretInput.Prompt = "> "
	secretInput.EchoMode = textinput.EchoPassword
	secretInput.EchoCharacter = '•'
	secretInput.CharLimit = 32

	page := &LastfmCustomApiPage{
		netease:      netease,
		menuTitle:    &model.MenuItem{Title: "Lastfm API account"},
		keyInput:     keyInput,
		secretInput:  secretInput,
		submitButton: model.GetBlurredSubmitButton(),

		reloadText:  "重载",
		clearText:   "清空",
		submitIndex: 2,
		reloadIndex: 3,
		clearIndex:  4,
	}
	page.reloadButton = model.GetBlurredButton(page.reloadText)
	page.clearButton = model.GetBlurredButton(page.clearText)
	page.reloadApiAccount()
	page.tips = ""
	return page

}

func (l *LastfmCustomApiPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LastfmCustomApiPage) Type() model.PageType {
	return LastfmCustomApiPageType
}

func (l *LastfmCustomApiPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	inputs := []*textinput.Model{
		&l.keyInput,
		&l.secretInput,
	}

	var (
		key tea.KeyMsg
		ok  bool
	)

	if key, ok = msg.(tea.KeyMsg); !ok {
		return l.updateAccountInputs(msg)
	}

	switch key.String() {
	case "b":
		if l.index != l.submitIndex && l.index != l.clearIndex {
			return l.updateAccountInputs(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		return l.netease.MustMain(), l.netease.RerenderCmd(true)
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
				return l.updateAccountInputs(msg)
			}
			if s == "left" && l.index >= l.submitIndex {
				l.index--
			} else if s == "right" && l.index <= l.clearIndex {
				l.index++
			}
		case "up", "shift+tab":
			l.index--
		default:
			l.index++
		}

		if l.index > l.clearIndex {
			l.index = 0
		} else if l.index < 0 {
			l.index = l.clearIndex
		}

		for i := 0; i <= len(inputs)-1; i++ {
			if i != l.index {
				// Remove focused state
				inputs[i].Blur()
				inputs[i].Prompt = model.GetBlurredPrompt()
				s := textinput.DefaultStyles(true)
				s.Focused.Text = lipgloss.NewStyle()
				inputs[i].SetStyles(s)
				continue
			}
			// Set focused state
			inputs[i].Focus()
			inputs[i].Prompt = model.GetFocusedPrompt()
			s := textinput.DefaultStyles(true)
			s.Focused.Text = util.GetPrimaryFontStyle()
			inputs[i].SetStyles(s)
		}

		// l.keyInput = *inputs[0]
		// l.secretInput = *inputs[1]

		l.tips = ""

		if l.index == submitIndex {
			l.tips = util.SetFgStyle("保存至数据库，优先使用此值", lipgloss.BrightBlue)
			l.submitButton = model.GetFocusedSubmitButton()
		} else {
			l.submitButton = model.GetBlurredSubmitButton()
		}

		if l.index == l.reloadIndex {
			l.tips = util.SetFgStyle("从数据库或本次启动时的配置文件中加载 API account", lipgloss.BrightBlue)
			l.reloadButton = model.GetFocusedButton(l.reloadText)
		} else {
			l.reloadButton = model.GetBlurredButton(l.reloadText)
		}

		if l.index == l.clearIndex {
			l.tips = util.SetFgStyle("清除当前值及已设置值", lipgloss.BrightBlue)
			l.clearButton = model.GetFocusedButton(l.clearText)
		} else {
			l.clearButton = model.GetBlurredButton(l.clearText)
		}

		return l, nil
	}

	// Handle character input and blinks
	return l.updateAccountInputs(msg)
}

func (l *LastfmCustomApiPage) View(a *model.App) string {
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
		&l.keyInput,
		&l.secretInput,
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
	builder.WriteString(l.reloadButton)

	builder.WriteString(btnBlank)
	builder.WriteString(l.clearButton)

	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - runewidth.StringWidth(types.SubmitText) - runewidth.StringWidth(l.clearText) - runewidth.StringWidth(l.reloadText) - len(btnBlank)*2
	if spaceLen > 0 {
		builder.WriteString(strings.Repeat(" ", spaceLen))
	}
	builder.WriteString("\n")

	if a.WindowHeight() > top+3 {
		builder.WriteString(strings.Repeat("\n", a.WindowHeight()-top-3))
	}

	return builder.String()
}

func (l *LastfmCustomApiPage) Msg() tea.Msg {
	return nil
}

func (l *LastfmCustomApiPage) updateAccountInputs(msg tea.Msg) (model.Page, tea.Cmd) {
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

func (l *LastfmCustomApiPage) enterHandler() (model.Page, tea.Cmd) {
	loading := model.NewLoading(l.netease.MustMain(), l.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case l.submitIndex:
		// 提交
		if len(l.keyInput.Value()) != 32 || len(l.secretInput.Value()) != 32 {
			l.tips = util.SetFgStyle("请输入正确的 API 账号或密码", lipgloss.BrightRed)
			return l, nil
		}
		l.netease.lastfm.SetApiAccount(l.keyInput.Value(), l.secretInput.Value())
		l.tips = util.SetFgStyle("已保存至数据库", lipgloss.BrightGreen)
	case l.reloadIndex:
		l.reloadApiAccount()
	case l.clearIndex:
		if len(l.keyInput.Value()) != 0 && len(l.secretInput.Value()) != 0 {
			l.keyInput.Reset()
			l.secretInput.Reset()
			l.tips = util.SetFgStyle("已清空，请重新填写, 为空时再次按下以清除数据库内 Api account", lipgloss.BrightRed)
		} else {
			l.netease.lastfm.ClearApiAccount()
			l.tips = util.SetFgStyle("已清除数据库内 Api account，需重新登录", lipgloss.BrightRed)
		}
	}
	if l.AfterAction != nil {
		l.AfterAction()
	}

	return l, tickLogin(time.Nanosecond)
}

func (l *LastfmCustomApiPage) reloadApiAccount() (model.Page, tea.Cmd) {
	// var key, secret string
	key, secret := l.netease.lastfm.GetApiAccount()
	if key != "" && secret != "" {
		l.keyInput.SetValue(key)
		l.secretInput.SetValue(secret)
		l.tips = util.SetFgStyle("已从已配置值(TUI 设置值)加载", lipgloss.BrightGreen)
	} else if configs.AppConfig.Reporter.Lastfm.Key != "" && configs.AppConfig.Reporter.Lastfm.Secret != "" {
		l.keyInput.SetValue(configs.AppConfig.Reporter.Lastfm.Key)
		l.secretInput.SetValue(configs.AppConfig.Reporter.Lastfm.Secret)
		l.tips = util.SetFgStyle("已从本次启动时的配置文件中加载", lipgloss.BrightGreen)
	} else {
		l.keyInput.Reset()
		l.secretInput.Reset()
		l.tips = util.SetFgStyle("未获取到内容，已重置", lipgloss.BrightGreen)
	}

	return l, nil
}
