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

var ramp []string

func updateStartup(msg tea.Msg, m NeteaseModel) (tea.Model, tea.Cmd) {
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

	case tickMsg:
		if m.loadedDuration >= m.TotalDuration {
			m.Quitting = true
			return m, tea.Quit
		}
		m.loadedDuration += constants.StartupTickDuration
		m.loadedPercent = ease.OutBounce(float64(m.loadedDuration) / float64(m.TotalDuration))
		return m, tick(constants.StartupTickDuration)
	}

	return m, nil
}

func startupView(m NeteaseModel) string {

	if WindowWidth <= 0 || WindowHeight <= 0 {
		return ""
	}
	
	blankLine := 1
	tipsHeight := 1
	progressHeight := 1
	height := utils.AsciiHeight + blankLine + tipsHeight + blankLine + progressHeight
	var top, buttom int
	if WindowHeight - height > 0 {
		top = (WindowHeight - height) / 2
	}
	if WindowHeight - top - height > 0 {
		buttom = WindowHeight - top - height
	}

	logo := logoView(m)
	tips := tipsView(m)
	progress := progressView(m)
	view := fmt.Sprintf("%s%s\n\n%s\n\n%s%s", strings.Repeat("\n", top-1), logo, tips, progress, strings.Repeat("\n", buttom))

	return view
}

func startupResize() {

}

// get logo
func logoView(m NeteaseModel) string {
	if WindowWidth <= 0 || WindowHeight <= 0 {
		return ""
	}

	originLogo := utils.GetAlphaAscii(constants.AppName)
	var logoWidth int
	if logoArr := strings.Split(originLogo, "\n"); len(logoArr) > 1 {
		logoWidth = utf8.RuneCountInString(logoArr[1])
	}

	var left int
	if WindowWidth - logoWidth > 0 {
		left = (WindowWidth - logoWidth) / 2
	}

	leftSpace := strings.Repeat(" ", left)
	lines := strings.Split(originLogo, "\n")
	for i := range lines {
		lines[i] = leftSpace + lines[i]
	}
	return SetFgStyle(strings.Join(lines, "\n"), primaryColor)
}

// get tips
func tipsView(m NeteaseModel) string {
	example := "Enter after 11.1 seconds..."
	var left int
	if WindowWidth - len(example) > 0 {
		left = (WindowWidth - len(example)) / 2
	}
	tips := fmt.Sprintf("%sEnter after %.1f seconds...",
		strings.Repeat(" ", left),
		float64(m.TotalDuration - m.loadedDuration)/float64(time.Second))

	return SetFgStyle(tips, termProfile.Color("8"))
}

// get progress
func progressView(m NeteaseModel) string {
	width := float64(WindowWidth - 2)

	startColor, endColor := GetRandomRgbColor(true)
	if len(ramp) == 0 {
		ramp = makeRamp(startColor, endColor, width)
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