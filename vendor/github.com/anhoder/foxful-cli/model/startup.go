package model

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fogleman/ease"
	"github.com/muesli/termenv"
)

var (
	progressRamp       []string
	progressLastWidth  float64
	progressStartColor string
	progressEndColor   string
)

func GetProgressColor() (start, end string) {
	if progressStartColor == "" || progressEndColor == "" {
		progressStartColor, progressEndColor = util.GetRandomRgbColor(true)
	}
	return progressStartColor, progressEndColor
}

type tickStartupMsg struct{}

type StartupPage struct {
	options *StartupOptions

	loadedDuration time.Duration
	loadedPercent  float64
	loaded         bool
	nextPage       Page
}

func NewStartup(options *StartupOptions, nextPage Page) *StartupPage {
	return &StartupPage{
		options:  options,
		nextPage: nextPage,
	}
}

func (s *StartupPage) Init(a *App) tea.Cmd {
	return a.Tick(time.Nanosecond)
}

func (s *StartupPage) Msg() tea.Msg {
	return tickStartupMsg{}
}

func (s *StartupPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return false
}

func (s *StartupPage) Type() PageType {
	return PtStartup
}

func (s *StartupPage) Update(msg tea.Msg, a *App) (Page, tea.Cmd) {
	switch msg.(type) {
	case tickStartupMsg:
		if s.loadedDuration >= s.options.LoadingDuration {
			s.loaded = true
			return s.nextPage, a.RerenderCmd(true)
		}
		s.loadedDuration += s.options.TickDuration
		s.loadedPercent = float64(s.loadedDuration) / float64(s.options.LoadingDuration)
		if s.options.ProgressOutBounce {
			s.loadedPercent = ease.OutBounce(s.loadedPercent)
		}
		return s, a.Tick(s.options.TickDuration)
	case tea.WindowSizeMsg:
		s.nextPage.Update(msg, a)
	}
	return s, nil
}

func (s *StartupPage) View(a *App) string {
	var windowHeight, windowWidth = a.WindowHeight(), a.WindowWidth()
	if windowWidth <= 0 || windowHeight <= 0 {
		return ""
	}

	blankLine := 1
	tipsHeight := 1
	progressHeight := 1
	height := util.AsciiHeight + blankLine + tipsHeight + blankLine + progressHeight
	var top, bottom int
	if windowHeight-height > 0 {
		top = (windowHeight - height) / 2
	}
	if windowHeight-top-height > 0 {
		bottom = windowHeight - top - height
	}

	var uiBuilder strings.Builder
	if top > 1 {
		uiBuilder.WriteString(strings.Repeat("\n", top-1))
	}
	uiBuilder.WriteString(s.logoView(a))
	uiBuilder.WriteString("\n")
	if top != 0 && bottom != 0 {
		uiBuilder.WriteString("\n")
	}
	uiBuilder.WriteString(s.tipsView(a))
	uiBuilder.WriteString("\n")
	if top != 0 && bottom != 0 {
		uiBuilder.WriteString("\n")
	}
	uiBuilder.WriteString(s.progressView(a))
	uiBuilder.WriteString(strings.Repeat("\n", bottom))

	return uiBuilder.String()
}

func (s *StartupPage) logoView(a *App) string {
	var windowHeight, windowWidth = a.WindowHeight(), a.WindowWidth()
	if windowWidth <= 0 || windowHeight <= 0 {
		return ""
	}

	originLogo := util.GetAlphaAscii(s.options.Welcome)
	var logoWidth int
	if logoArr := strings.Split(originLogo, "\n"); len(logoArr) > 1 {
		logoWidth = utf8.RuneCountInString(logoArr[1])
	}

	var left int
	if windowWidth-logoWidth > 0 {
		left = (windowWidth - logoWidth) / 2
	}

	var logoBuilder strings.Builder
	leftSpace := strings.Repeat(" ", left)
	lines := strings.Split(originLogo, "\n")
	for _, line := range lines {
		logoBuilder.WriteString(leftSpace)
		logoBuilder.WriteString(line)
		logoBuilder.WriteString("\n")
	}
	return util.SetFgStyle(logoBuilder.String(), util.GetPrimaryColor())
}

func (s *StartupPage) tipsView(a *App) string {
	example := "Enter after 11.1 seconds..."
	var (
		left        int
		windowWidth = a.WindowWidth()
	)
	if windowWidth-len(example) > 0 {
		left = (windowWidth - len(example)) / 2
	}
	tips := fmt.Sprintf("%sEnter after %.1f seconds...",
		strings.Repeat(" ", left),
		float64(s.options.LoadingDuration-s.loadedDuration)/float64(time.Second))

	return util.SetFgStyle(tips, termenv.ANSIBrightBlack)
}

func (s *StartupPage) progressView(a *App) string {
	var width = float64(a.WindowWidth())

	start, end := GetProgressColor()
	if width != progressLastWidth {
		progressRamp = util.MakeRamp(start, end, width)
		progressLastWidth = width
	}

	return Progress(&a.options.ProgressOptions, int(width), int(math.Round(width*s.loadedPercent)), progressRamp)
}
