package ui

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/bubbles/textinput"
	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/netease-music/service"
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

type SearchModel struct {
	index        int
	wordsInput   textinput.Model
	submitButton string
	tips         string
	searchType   SearchType
	result       interface{}
}

func (s *SearchModel) Reset() {
	s.tips = ""
	s.wordsInput.SetValue("")
	s.wordsInput.Reset()
	s.index = 0
	s.wordsInput.Focus()
	s.wordsInput.Prompt = GetFocusedPrompt()
	s.wordsInput.TextColor = primaryColorStr
	s.wordsInput.CharLimit = 32
	s.submitButton = GetBlurredSubmitButton()
}

func NewSearch() (search *SearchModel) {
	search = new(SearchModel)
	search.wordsInput = textinput.NewModel()
	search.wordsInput.Placeholder = " 输入关键词"
	search.wordsInput.Focus()
	search.wordsInput.Prompt = GetFocusedPrompt()
	search.wordsInput.TextColor = primaryColorStr
	search.wordsInput.CharLimit = 32

	search.submitButton = GetBlurredSubmitButton()

	return
}

// update search
func (s *SearchModel) update(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tickSearchMsg:
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.pageType = PtMain
			m.searchModel.Reset()
			return m, tickMainUI(time.Nanosecond)

		// Cycle between inputs
		case "tab", "shift+tab", "enter", "up", "down":
			if m.searchModel.searchType == StNull {
				return m, nil
			}

			inputs := []textinput.Model{
				m.searchModel.wordsInput,
			}

			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.searchModel.index == len(inputs) {
				if len(m.searchModel.wordsInput.Value()) <= 0 {
					m.searchModel.tips = SetFgStyle("关键词不得为空", termenv.ANSIBrightRed)
					return m, nil
				}
				var (
					code     float64
					response []byte
				)
				searchService := service.SearchService{
					S:     m.searchModel.wordsInput.Value(),
					Type:  strconv.Itoa(int(m.searchModel.searchType)),
					Limit: strconv.Itoa(constants.SearchPageSize),
				}
				code, response = searchService.Search()

				codeType := utils.CheckCode(code)
				switch codeType {
				case utils.UnknownError:
					m.searchModel.tips = SetFgStyle("未知错误，请稍后再试~", termenv.ANSIBrightRed)
					return m, tickSearch(time.Nanosecond)
				case utils.NetworkError:
					m.searchModel.tips = SetFgStyle("网络异常，请稍后再试~", termenv.ANSIBrightRed)
					return m, tickSearch(time.Nanosecond)
				case utils.Success:
					m.searchModel.result = response

					switch m.searchModel.searchType {
					case StSingleSong:
						m.searchModel.result = utils.GetSongsOfSearchResult(response)
					case StAlbum:
						m.searchModel.result = utils.GetAlbumsOfSearchResult(response)
					case StSinger:
						m.searchModel.result = utils.GetArtistsOfSearchResult(response)
					case StPlaylist:
						m.searchModel.result = utils.GetPlaylistsOfSearchResult(response)
					case StUser:
						m.searchModel.result = utils.GetUsersOfSearchResult(response)
					case StLyric:
						m.searchModel.result = utils.GetSongsOfSearchResult(response)
					case StRadio:
						m.searchModel.result = utils.GetDjRadiosOfSearchResult(response)
					}
					enterMenu(m, nil, nil)
				}

				m.pageType = PtMain
				m.searchModel.Reset()
				return m, tickMainUI(time.Nanosecond)
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.searchModel.index--
			} else {
				m.searchModel.index++
			}

			if m.searchModel.index > len(inputs) {
				m.searchModel.index = 0
			} else if m.searchModel.index < 0 {
				m.searchModel.index = len(inputs)
			}

			for i := 0; i <= len(inputs)-1; i++ {
				if i == m.searchModel.index {
					// Set focused state
					inputs[i].Focus()
					inputs[i].Prompt = GetFocusedPrompt()
					inputs[i].TextColor = primaryColorStr
					continue
				}
				// Remove focused state
				inputs[i].Blur()
				inputs[i].Prompt = GetBlurredPrompt()
				inputs[i].TextColor = ""
			}

			m.searchModel.wordsInput = inputs[0]

			if m.searchModel.index == len(inputs) {
				m.searchModel.submitButton = GetFocusedSubmitButton()
			} else {
				m.searchModel.submitButton = GetBlurredSubmitButton()
			}

			return m, nil
		}
	}

	// Handle character input and blinks
	return s.updateSearchInputs(msg, m)
}

func (s *SearchModel) updateSearchInputs(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.searchModel.wordsInput, cmd = m.searchModel.wordsInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (s *SearchModel) view(m *NeteaseModel) string {
	var builder strings.Builder

	// 距离顶部的行数
	top := 0

	// title
	if configs.ConfigRegistry.MainShowTitle {

		builder.WriteString(m.titleView(m, &top))
	} else {
		top++
	}

	// menu title
	menuViews := m.menu.MenuViews()
	var menuTitle = &MenuItem{Title: "搜索"}
	if m.selectedIndex < len(menuViews) {
		typeMenu := menuViews[m.selectedIndex]
		menuTitle.Subtitle = typeMenu.Title
	}
	builder.WriteString(m.menuTitleView(m, &top, menuTitle))
	builder.WriteString("\n\n\n")
	top += 2

	inputs := []textinput.Model{
		m.searchModel.wordsInput,
	}

	for i, input := range inputs {
		if m.menuStartColumn > 0 {
			builder.WriteString(strings.Repeat(" ", m.menuStartColumn))
		}

		builder.WriteString(input.View())

		var valueLen int
		if input.Value() == "" {
			valueLen = runewidth.StringWidth(input.Placeholder)
		} else {
			valueLen = runewidth.StringWidth(input.Value())
		}
		if spaceLen := m.WindowWidth - m.menuStartColumn - valueLen - 3; spaceLen > 0 {
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
	if m.menuStartColumn > 0 {
		builder.WriteString(strings.Repeat(" ", m.menuStartColumn))
	}
	builder.WriteString(m.searchModel.tips)
	builder.WriteString("\n\n")
	top++
	if m.menuStartColumn > 0 {
		builder.WriteString(strings.Repeat(" ", m.menuStartColumn))
	}
	builder.WriteString(m.searchModel.submitButton)
	builder.WriteString("\n")

	if m.WindowHeight > top+3 {
		builder.WriteString(strings.Repeat("\n", m.WindowHeight-top-3))
	}

	return builder.String()
}

// SearchHandle 搜索
func (s *SearchModel) SearchHandle(model *NeteaseModel, searchType SearchType) {
	model.pageType = PtSearch
	model.searchModel.searchType = searchType
}
