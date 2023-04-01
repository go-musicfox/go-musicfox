package ui

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/utils"

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

type StartupModel struct {
	TotalDuration  time.Duration
	loadedDuration time.Duration
	loadedPercent  float64
	loaded         bool
	quitting       bool
}

func NewStartup() (m *StartupModel) {
	m = new(StartupModel)
	m.TotalDuration = time.Second * 2 // 默认
	return
}

// startup func
func (s *StartupModel) update(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickStartupMsg:
		if s.loadedDuration >= s.TotalDuration {
			s.loaded = true
			m.isListeningKey = true
			return m, m.rerenderTicker(true)
		}
		s.loadedDuration += constants.StartupTickDuration
		s.loadedPercent = float64(s.loadedDuration) / float64(s.TotalDuration)
		if configs.ConfigRegistry.StartupProgressOutBounce {
			s.loadedPercent = ease.OutBounce(s.loadedPercent)
		}
		return m, tickStartup(constants.StartupTickDuration)
	}

	return m, nil
}

// startup view
func (s *StartupModel) view(m *NeteaseModel) string {

	if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	blankLine := 1
	tipsHeight := 1
	progressHeight := 1
	height := utils.AsciiHeight + blankLine + tipsHeight + blankLine + progressHeight
	var top, bottom int
	if m.WindowHeight-height > 0 {
		top = (m.WindowHeight - height) / 2
	}
	if m.WindowHeight-top-height > 0 {
		bottom = m.WindowHeight - top - height
	}

	var uiBuilder strings.Builder
	if top > 1 {
		uiBuilder.WriteString(strings.Repeat("\n", top-1))
	}
	uiBuilder.WriteString(s.logoView(m))
	uiBuilder.WriteString("\n")
	if top != 0 && bottom != 0 {
		uiBuilder.WriteString("\n")
	}
	uiBuilder.WriteString(s.tipsView(m))
	uiBuilder.WriteString("\n")
	if top != 0 && bottom != 0 {
		uiBuilder.WriteString("\n")
	}
	uiBuilder.WriteString(s.progressView(m))
	uiBuilder.WriteString(strings.Repeat("\n", bottom))

	return uiBuilder.String()
}

// get logo
func (s *StartupModel) logoView(m *NeteaseModel) string {
	if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	originLogo := utils.GetAlphaAscii(configs.ConfigRegistry.StartupWelcome)
	var logoWidth int
	if logoArr := strings.Split(originLogo, "\n"); len(logoArr) > 1 {
		logoWidth = utf8.RuneCountInString(logoArr[1])
	}

	var left int
	if m.WindowWidth-logoWidth > 0 {
		left = (m.WindowWidth - logoWidth) / 2
	}

	var logoBuilder strings.Builder
	leftSpace := strings.Repeat(" ", left)
	lines := strings.Split(originLogo, "\n")
	for _, line := range lines {
		logoBuilder.WriteString(leftSpace)
		logoBuilder.WriteString(line)
		logoBuilder.WriteString("\n")
	}
	return SetFgStyle(logoBuilder.String(), GetPrimaryColor())
}

// get tips
func (s *StartupModel) tipsView(m *NeteaseModel) string {
	example := "Enter after 11.1 seconds..."
	var left int
	if m.WindowWidth-len(example) > 0 {
		left = (m.WindowWidth - len(example)) / 2
	}
	tips := fmt.Sprintf("%sEnter after %.1f seconds...",
		strings.Repeat(" ", left),
		float64(s.TotalDuration-s.loadedDuration)/float64(time.Second))

	return SetFgStyle(tips, termenv.ANSIBrightBlack)
}

// get progress
func (s *StartupModel) progressView(m *NeteaseModel) string {
	width := float64(m.WindowWidth)

	if progressStartColor == "" || progressEndColor == "" {
		progressStartColor, progressEndColor = GetRandomRgbColor(true)
	}
	if width != progressLastWidth {
		progressRamp = makeRamp(progressStartColor, progressEndColor, width)
		progressLastWidth = width
	}

	fullSize := int(math.Round(width * s.loadedPercent))
	var fullCells string
	for i := 0; i < fullSize && i < len(progressRamp); i++ {
		fullCells += termenv.String(string(configs.ConfigRegistry.ProgressFullChar)).Foreground(termProfile.Color(progressRamp[i])).String()
	}

	emptySize := 0
	if int(width)-fullSize > 0 {
		emptySize = int(width) - fullSize
	}
	emptyCells := SetFgStyle(strings.Repeat(string(configs.ConfigRegistry.ProgressEmptyChar), emptySize), termenv.ANSIBrightBlack)

	return fmt.Sprintf("%s%s", fullCells, emptyCells)
}
