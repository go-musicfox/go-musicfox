package ui

import (
	"fmt"
	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/go-musicfox/constants"
	"github.com/anhoder/go-musicfox/utils"
	"github.com/fogleman/ease"
	"github.com/muesli/termenv"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ramp []string
	lastWidth float64
)

type startupModel struct {
	TotalDuration  time.Duration
	loadedDuration time.Duration
	loadedPercent  float64
	loaded         bool
	quitting       bool
}

// startup func
func updateStartup(msg tea.Msg, m neteaseModel) (tea.Model, tea.Cmd) {
	switch msg.(type) {

	case tea.KeyMsg:
		//switch msg.String() {
		//case "j", "down":
		//	m.Choice += 1
		//	if m.Choice > 3 {
		//		m.Choice = 3
		//	}
		//case "k", "up":
		//	m.Choice -= 1
		//	if m.Choice < 0 {
		//		m.Choice = 0
		//	}
		//case "enter":
		//	m.Chosen = true
		//	return m, frame()
		//}

	case tickStartupMsg:
		if m.loadedDuration >= m.TotalDuration {
			m.loaded = true
			termenv.ClearScreen()
			return m, tickMainUI(time.Nanosecond)
		}
		m.loadedDuration += constants.StartupTickDuration
		m.loadedPercent = ease.OutBounce(float64(m.loadedDuration) / float64(m.TotalDuration))
		return m, tickStartup(constants.StartupTickDuration)
	}

	return m, nil
}

// startup view
func startupView(m neteaseModel) string {

	if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}
	
	blankLine := 1
	tipsHeight := 1
	progressHeight := 1
	height := utils.AsciiHeight + blankLine + tipsHeight + blankLine + progressHeight
	var top, buttom int
	if m.WindowHeight - height > 0 {
		top = (m.WindowHeight - height) / 2
	}
	if m.WindowHeight - top - height > 0 {
		buttom = m.WindowHeight - top - height
	}

	logo := logoView(m)
	tips := tipsView(m)
	progress := progressView(m)
	view := fmt.Sprintf("%s%s\n\n%s\n\n%s%s", strings.Repeat("\n", top-1), logo, tips, progress, strings.Repeat("\n", buttom))

	return view
}

// get logo
func logoView(m neteaseModel) string {
	if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	originLogo := utils.GetAlphaAscii(constants.AppName)
	var logoWidth int
	if logoArr := strings.Split(originLogo, "\n"); len(logoArr) > 1 {
		logoWidth = utf8.RuneCountInString(logoArr[1])
	}

	var left int
	if m.WindowWidth - logoWidth > 0 {
		left = (m.WindowWidth - logoWidth) / 2
	}

	leftSpace := strings.Repeat(" ", left)
	lines := strings.Split(originLogo, "\n")
	for i := range lines {
		lines[i] = leftSpace + lines[i]
	}
	return SetFgStyle(strings.Join(lines, "\n"), primaryColor)
}

// get tips
func tipsView(m neteaseModel) string {
	example := "Enter after 11.1 seconds..."
	var left int
	if m.WindowWidth - len(example) > 0 {
		left = (m.WindowWidth - len(example)) / 2
	}
	tips := fmt.Sprintf("%sEnter after %.1f seconds...",
		strings.Repeat(" ", left),
		float64(m.TotalDuration - m.loadedDuration)/float64(time.Second))

	return SetFgStyle(tips, termProfile.Color("8"))
}

// get progress
func progressView(m neteaseModel) string {
	width := float64(m.WindowWidth - 2)

	startColor, endColor := GetRandomRgbColor(true)
	if width != lastWidth {
		ramp = makeRamp(startColor, endColor, width)
		lastWidth = width
	}

	fullSize := int(math.Round(width*m.loadedPercent))
	var fullCells string
	for i := 0; i < fullSize; i++ {
		fullCells += termenv.String(string(constants.ProgressFullChar)).Foreground(termProfile.Color(ramp[i])).String()
	}

	emptySize := int(width) - fullSize
	emptyCells := strings.Repeat(string(constants.ProgressEmptyChar), emptySize)

	return fmt.Sprintf("%s%s", fullCells, emptyCells)
}