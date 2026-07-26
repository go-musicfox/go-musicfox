package ui

import (
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
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

	backBtnHovered bool
	backBtnRowY    int // 0-based 屏幕行
	backBtnStartX  int // 0-based 起始列

	index        int
	wordsInput   textinput.Model
	submitButton string
	tips         string
	searchType   SearchType
	result       any

	// 鼠标点击区域坐标（用于 hover 和点击检测）
	inputRowY     int // 输入框所在行号（0-based）
	inputStartX   int // 输入框起始列（0-based）
	inputEndX     int // 输入框结束列（0-based，闭区间）
	submitRowY    int // 提交按钮所在行号（0-based）
	submitStartX  int // 提交按钮起始列（0-based）
	submitEndX    int // 提交按钮结束列（0-based，闭区间）
	hoveredInput  bool
	hoveredSubmit bool
	mousePointer  string
}

func NewSearchPage(netease *Netease) (search *SearchPage) {
	search = &SearchPage{
		netease:      netease,
		menuTitle:    &model.MenuItem{Title: model.T(MsgSearchPageTitle)},
		wordsInput:   textinput.New(),
		submitButton: pageSubmitButton(false),
	}
	search.wordsInput.Placeholder = model.T(MsgSearchPlaceholder)
	focusPageInput(&search.wordsInput)
	search.wordsInput.CharLimit = 32
	return
}

func (s *SearchPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (s *SearchPage) Type() model.PageType {
	return PageTypeSearch
}

func (s *SearchPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	if _, ok := msg.(tickSearchMsg); ok {
		return s, nil
	}

	if mouseMsg, ok := msg.(tea.MouseMotionMsg); ok {
		mouse := mouseMsg.Mouse()
		oldBackHovered := s.backBtnHovered
		oldInputHovered := s.hoveredInput
		oldSubmitHovered := s.hoveredSubmit
		oldPointer := s.mousePointer

		// Delegate breadcrumb hover to Main when top status bar is rendered
		bcChanged, bcOver := pageBreadcrumbMotion(a, s.netease.MustMain(), mouse.X, mouse.Y)

		s.backBtnHovered = mouse.Y == s.backBtnRowY && mouse.X >= s.backBtnStartX && mouse.X < s.backBtnStartX+pageBackButtonWidth
		s.hoveredInput = mouse.Y == s.inputRowY && mouse.X >= s.inputStartX && mouse.X <= s.inputEndX
		s.hoveredSubmit = mouse.Y == s.submitRowY && mouse.X >= s.submitStartX && mouse.X <= s.submitEndX
		s.mousePointer = "default"
		if s.hoveredInput {
			s.mousePointer = "text"
		} else if s.backBtnHovered || s.hoveredSubmit || bcOver {
			s.mousePointer = "pointer"
		}

		if s.backBtnHovered != oldBackHovered || s.hoveredInput != oldInputHovered || s.hoveredSubmit != oldSubmitHovered || s.mousePointer != oldPointer || bcChanged {
			return s, tea.Sequence(tickSearch(time.Nanosecond), a.SetMousePointer(s.mousePointer))
		}
		return s.updateSearchInputs(msg)
	}

	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		if clickMsg.Mouse().Button == tea.MouseLeft {
			mouse := clickMsg.Mouse()
			// Breadcrumb click delegates to Main navigation (top status bar only)
			if newPage := pageBreadcrumbClick(a, s.netease.MustMain(), mouse.X, mouse.Y); newPage != nil {
				s.Reset()
				return newPage, s.netease.RerenderCmd(true)
			}
			if mouse.Y == s.backBtnRowY && mouse.X >= s.backBtnStartX && mouse.X < s.backBtnStartX+pageBackButtonWidth {
				s.Reset()
				return s.netease.MustMain(), s.netease.RerenderCmd(true)
			}

			// 点击输入框：聚焦并将光标移动到点击位置
			if mouse.Y == s.inputRowY && mouse.X >= s.inputStartX && mouse.X <= s.inputEndX {
				s.index = 0
				focusPageInput(&s.wordsInput)
				setPageInputCursor(&s.wordsInput, mouse.X, s.inputStartX)
				s.submitButton = pageSubmitButton(false)
				return s, tickSearch(time.Nanosecond)
			}

			// 点击提交按钮：触发提交
			if mouse.Y == s.submitRowY && mouse.X >= s.submitStartX && mouse.X <= s.submitEndX {
				s.index = 1
				return s.enterHandler()
			}
		}
		return s.updateSearchInputs(msg)
	}

	key, ok := msg.(tea.KeyPressMsg)
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

		for i := range inputs {
			if i == s.index {
				focusPageInput(&inputs[i])
				continue
			}
			blurPageInput(&inputs[i])
		}

		s.wordsInput = inputs[0]
		s.submitButton = pageSubmitButton(s.index == len(inputs))

		return s, nil
	}

	// Handle character input and blinks
	return s.updateSearchInputs(msg)
}

func (s *SearchPage) enterHandler() (model.Page, tea.Cmd) {
	if len(s.wordsInput.Value()) <= 0 {
		s.tips = util.SetFgStyle(model.T(MsgSearchKeywordRequired), lipgloss.BrightRed)
		return s, nil
	}
	loading := model.NewLoading(s.netease.MustMain(), s.menuTitle)
	loading.DisplayNotOnlyOnMain()
	loading.Start()
	defer loading.Complete()

	var (
		code     float64
		response []byte
	)
	searchService := service.SearchService{
		S:     s.wordsInput.Value(),
		Type:  strconv.Itoa(int(s.searchType)),
		Limit: strconv.Itoa(types.SearchPageSize),
	}
	code, response = searchService.Search()

	codeType := _struct.CheckCode(code)
	switch codeType {
	case _struct.UnknownError:
		s.tips = util.SetFgStyle(model.T(MsgSearchUnknownError), lipgloss.BrightRed)
		return s, tickSearch(time.Nanosecond)
	case _struct.NetworkError:
		s.tips = util.SetFgStyle(model.T(MsgSearchNetworkError), lipgloss.BrightRed)
		return s, tickSearch(time.Nanosecond)
	case _struct.Success:
		s.result = response
		switch s.searchType {
		case StSingleSong:
			s.result = _struct.GetSongsOfSearchResult(response)
		case StAlbum:
			s.result = _struct.GetAlbumsOfSearchResult(response)
		case StSinger:
			s.result = _struct.GetArtistsOfSearchResult(response)
		case StPlaylist:
			s.result = _struct.GetPlaylistsOfSearchResult(response)
		case StUser:
			s.result = _struct.GetUsersOfSearchResult(response)
		case StLyric:
			s.result = _struct.GetSongsOfSearchResult(response)
		case StRadio:
			s.result = _struct.GetDjRadiosOfSearchResult(response)
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

	lineCount := 0
	write := func(text string) {
		builder.WriteString(text)
		lineCount += strings.Count(text, "\n")
	}

	// title
	if configs.AppConfig.Theme.ShowTitle {
		write(pageTitleView(a, main, &top))
	} else {
		write("\n")
		top++
	}

	// menu title
	menuViews := main.CurMenu().MenuViews()
	if main.SelectedIndex() < len(menuViews) {
		typeMenu := menuViews[main.SelectedIndex()]
		s.menuTitle.Subtitle = typeMenu.Title
	}
	topBefore := top
	write(pageMenuTitleViewWithBack(a, main, &top, s.menuTitle, s.backBtnHovered))
	s.backBtnRowY = pageMenuTitleRow(a, main, topBefore)
	s.backBtnStartX = max(0, main.MenuStartColumn()-pageBackButtonWidth)
	write("\n\n")

	if main.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", main.MenuStartColumn()))
	}
	s.inputRowY = lineCount
	s.inputStartX = max(0, main.MenuStartColumn())
	s.inputEndX = max(s.inputStartX, a.WindowWidth()-1)
	s.wordsInput.SetWidth(max(1, a.WindowWidth()-s.inputStartX-runewidth.StringWidth(s.wordsInput.Prompt)))
	write(pageInputView(s.wordsInput, s.hoveredInput))

	write("\n\n")
	if main.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", main.MenuStartColumn()))
	}
	write(s.tips)
	write("\n\n")
	if main.MenuStartColumn() > 0 {
		write(strings.Repeat(" ", main.MenuStartColumn()))
	}
	s.submitRowY = lineCount
	s.submitStartX = max(0, main.MenuStartColumn())
	submitButtonView := s.submitButton
	if s.hoveredSubmit {
		submitButtonView = pageButtonHoverView(pageSubmitText())
	}
	s.submitEndX = s.submitStartX + lipgloss.Width(submitButtonView) - 1
	write(submitButtonView)
	spaceLen := a.WindowWidth() - main.MenuStartColumn() - lipgloss.Width(submitButtonView)
	if spaceLen > 0 {
		write(strings.Repeat(" ", spaceLen))
	}
	write("\n")

	fillPageHeight(&builder, a.WindowHeight())

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
	focusPageInput(&s.wordsInput)
	s.wordsInput.CharLimit = 32
	s.submitButton = pageSubmitButton(false)
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
