package ui

import (
	"strconv"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

type SearchType uint32

const (
	StNull       SearchType = 0
	StSingleSong SearchType = 1
	StAlbum      SearchType = 10
	StSinger     SearchType = 100
	StPlaylist   SearchType = 1000
	StUser       SearchType = 1002
	StLyric      SearchType = 1006
	StRadio      SearchType = 1009
)

const PageTypeSearch model.PageType = "search"

type tickSearchMsg struct{}

func tickSearch(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return tickSearchMsg{}
	})
}

type SearchPage struct {
	netease   *Netease
	menuTitle *model.MenuItem

	index        int
	wordsInput   textinput.Model
	submitButton string
	tips         string
	searchType   SearchType
	result       interface{}
}

func NewSearchPage(netease *Netease) (search *SearchPage) {
	search = &SearchPage{
		netease:      netease,
		menuTitle:    &model.MenuItem{Title: "搜索"},
		wordsInput:   textinput.New(),
		submitButton: model.GetBlurredSubmitButton(),
	}
	search.wordsInput.Placeholder = " 输入关键词"
	search.wordsInput.Focus()
	search.wordsInput.Prompt = model.GetFocusedPrompt()
	search.wordsInput.TextStyle = util.GetPrimaryFontStyle()
	search.wordsInput.CharLimit = 32
	return
}

func (s *SearchPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (s *SearchPage) Type() model.PageType {
	return PageTypeSearch
}

func (s *SearchPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	if _, ok := msg.(tickSearchMsg); ok {
		return s, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return s.updateSearchInputs(msg)
	}

	switch key.String() {
	case "esc":
		s.Reset()
		return s.netease.MustMain(), s.netease.RerenderCmd(true)

	// Cycle between inputs
	case "tab", "shift+tab", "enter", "up", "down":
		if s.searchType == StNull {
			return s, nil
		}

		inputs := []textinput.Model{
			s.wordsInput,
		}

		k := key.String()

		// Did the user press enter while the submit button was focused?
		// If so, exit.
		if k == "enter" && s.index == len(inputs) {
			return s.enterHandler()
		}

		// Cycle indexes
		if k == "up" || k == "shift+tab" {
			s.index--
		} else {
			s.index++
		}

		if s.index > len(inputs) {
			s.index = 0
		} else if s.index < 0 {
			s.index = len(inputs)
		}

		for i := 0; i <= len(inputs)-1; i++ {
			if i == s.index {
				// Set focused state
				inputs[i].Focus()
				inputs[i].Prompt = model.GetFocusedPrompt()
				inputs[i].TextStyle = util.GetPrimaryFontStyle()
				continue
			}
			// Remove focused state
			inputs[i].Blur()
			inputs[i].Prompt = model.GetBlurredPrompt()
			inputs[i].TextStyle = lipgloss.NewStyle()
		}

		s.wordsInput = inputs[0]
		if s.index == len(inputs) {
			s.submitButton = model.GetFocusedSubmitButton()
		} else {
			s.submitButton = model.GetBlurredSubmitButton()
		}

		return s, nil
	}

	// Handle character input and blinks
	return s.updateSearchInputs(msg)
}

func (s *SearchPage) enterHandler() (model.Page, tea.Cmd) {
	if len(s.wordsInput.Value()) <= 0 {
		s.tips = util.SetFgStyle("关键词不得为空", termenv.ANSIBrightRed)
		return s, nil
	}
	loading := NewLoading(s.netease, s.menuTitle)
	loading.start()
	defer loading.complete()

	var (
		code     float64
		response []byte
	)
	searchService := service.SearchService{
		S:     s.wordsInput.Value(),
		Type:  strconv.Itoa(int(s.searchType)),
		Limit: strconv.Itoa(constants.SearchPageSize),
	}
	code, response = searchService.Search()

	codeType := utils.CheckCode(code)
	switch codeType {
	case utils.UnknownError:
		s.tips = util.SetFgStyle("未知错误，请稍后再试~", termenv.ANSIBrightRed)
		return s, tickSearch(time.Nanosecond)
	case utils.NetworkError:
		s.tips = util.SetFgStyle("网络异常，请稍后再试~", termenv.ANSIBrightRed)
		return s, tickSearch(time.Nanosecond)
	case utils.Success:
		s.result = response
		switch s.searchType {
		case StSingleSong:
			s.result = utils.GetSongsOfSearchResult(response)
		case StAlbum:
			s.result = utils.GetAlbumsOfSearchResult(response)
		case StSinger:
			s.result = utils.GetArtistsOfSearchResult(response)
		case StPlaylist:
			s.result = utils.GetPlaylistsOfSearchResult(response)
		case StUser:
			s.result = utils.GetUsersOfSearchResult(response)
		case StLyric:
			s.result = utils.GetSongsOfSearchResult(response)
		case StRadio:
			s.result = utils.GetDjRadiosOfSearchResult(response)
		}
		s.netease.MustMain().EnterMenu(nil, nil)
	}

	s.Reset()
	return s.netease.MustMain(), s.netease.Tick(time.Nanosecond)
}

func (s *SearchPage) View(a *model.App) string {
	var (
		builder strings.Builder
		top     int // 距离顶部的行数
		main    = s.netease.MustMain()
	)

	// title
	if configs.ConfigRegistry.MainShowTitle {
		builder.WriteString(main.TitleView(a, &top))
	} else {
		top++
	}

	// menu title
	menuViews := main.CurMenu().MenuViews()
	if main.SelectedIndex() < len(menuViews) {
		typeMenu := menuViews[main.SelectedIndex()]
		s.menuTitle.Subtitle = typeMenu.Title
	}
	builder.WriteString(main.MenuTitleView(a, &top, s.menuTitle))
	builder.WriteString("\n\n\n")
	top += 2

	inputs := []textinput.Model{
		s.wordsInput,
	}

	for i, input := range inputs {
		if main.MenuStartColumn() > 0 {
			builder.WriteString(strings.Repeat(" ", main.MenuStartColumn()))
		}

		builder.WriteString(input.View())

		var valueLen int
		if input.Value() == "" {
			valueLen = runewidth.StringWidth(input.Placeholder)
		} else {
			valueLen = runewidth.StringWidth(input.Value())
		}
		if spaceLen := a.WindowWidth() - main.MenuStartColumn() - valueLen - 3; spaceLen > 0 {
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
	if main.MenuStartColumn() > 0 {
		builder.WriteString(strings.Repeat(" ", main.MenuStartColumn()))
	}
	builder.WriteString(s.tips)
	builder.WriteString("\n\n")
	top++
	if main.MenuStartColumn() > 0 {
		builder.WriteString(strings.Repeat(" ", main.MenuStartColumn()))
	}
	builder.WriteString(s.submitButton)
	spaceLen := a.WindowWidth() - main.MenuStartColumn() - runewidth.StringWidth(model.SubmitText)
	if spaceLen > 0 {
		builder.WriteString(strings.Repeat(" ", spaceLen))
	}
	builder.WriteString("\n")

	if a.WindowHeight() > top+3 {
		builder.WriteString(strings.Repeat("\n", a.WindowHeight()-top-3))
	}

	return builder.String()
}

func (s *SearchPage) Msg() tea.Msg {
	return &tickSearchMsg{}
}

func (s *SearchPage) Reset() {
	s.tips = ""
	s.wordsInput.SetValue("")
	s.wordsInput.Reset()
	s.index = 0
	s.wordsInput.Focus()
	s.wordsInput.Prompt = model.GetFocusedPrompt()
	s.wordsInput.TextStyle = util.GetPrimaryFontStyle()
	s.wordsInput.CharLimit = 32
	s.submitButton = model.GetBlurredSubmitButton()
}

func (s *SearchPage) updateSearchInputs(msg tea.Msg) (model.Page, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	s.wordsInput, cmd = s.wordsInput.Update(msg)
	cmds = append(cmds, cmd)
	return s, tea.Batch(cmds...)
}
