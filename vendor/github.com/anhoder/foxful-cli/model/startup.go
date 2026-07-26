package model

import (
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/util"
	"github.com/fogleman/ease"
)

var (
	progressRamp       []color.Color
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

	startedAt      time.Time
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
	s.startedAt = time.Now()
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
		if s.options.LoadingDuration <= 0 {
			s.loaded = true
			return s.nextPage, a.RerenderCmd(true)
		}
		if s.startedAt.IsZero() {
			s.startedAt = time.Now()
		}

		elapsed := time.Since(s.startedAt)
		if elapsed >= s.options.LoadingDuration {
			s.loadedDuration = s.options.LoadingDuration
			s.loadedPercent = 1
			s.loaded = true
			return s.nextPage, a.RerenderCmd(true)
		}

		s.loadedDuration = elapsed
		s.loadedPercent = float64(elapsed) / float64(s.options.LoadingDuration)
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
	windowWidth, windowHeight := a.WindowWidth(), a.WindowHeight()
	if windowWidth <= 0 || windowHeight <= 0 {
		return ""
	}
	if special, ok := s.startupSpecialView(a); ok {
		return special
	}

	content := layout.JoinVertical(
		lipgloss.Center,
		s.animatedLogoView(a),
		"",
		s.startupStatusView(a),
		"",
		s.progressView(a),
	)

	return layout.Place(
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (s *StartupPage) progressView(a *App) string {
	var width = float64(a.WindowWidth())

	start, end := GetProgressColor()
	if width != progressLastWidth {
		progressRamp = util.MakeRamp(start, end, width)
		progressLastWidth = width
	}

	semanticPercent := s.animationProgress()
	visualPercent := s.loadedPercent
	if s.options.ReducedMotion {
		visualPercent = semanticPercent
	}
	return Progress(&a.options.ProgressOptions, int(width), int(math.Round(width*visualPercent)), progressRamp)
}
