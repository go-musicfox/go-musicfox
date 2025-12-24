package ui

import (
	"fmt"
	"math"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressRenderer is a dedicated UI component for rendering the playback progress bar.
type ProgressRenderer struct {
	netease *Netease
	state   playerRendererState

	progressLastWidth float64
	progressRamp      []string
}

// NewProgressRenderer creates a new progress bar renderer component.
func NewProgressRenderer(netease *Netease, state playerRendererState) *ProgressRenderer {
	return &ProgressRenderer{
		netease: netease,
		state:   state,
	}
}

// Update handles UI messages.
func (r *ProgressRenderer) Update(msg tea.Msg, a *model.App) {}

// View renders the progress bar component.
func (r *ProgressRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	song := r.state.CurSong()
	allDuration := int(song.Duration.Seconds())
	if allDuration == 0 {
		r.progressRamp = nil
		return "", 1 // Return 1 line to maintain layout consistency
	}
	passedDuration := int(r.state.PassedTime().Seconds())

	// Ensure progress does not exceed 100%
	var progress int
	if passedDuration > allDuration {
		progress = 100
	} else {
		progress = passedDuration * 100 / allDuration
	}

	width := float64(r.netease.WindowWidth() - 14)
	start, end := model.GetProgressColor()
	if width != r.progressLastWidth || len(r.progressRamp) == 0 {
		r.progressRamp = util.MakeRamp(start, end, width)
		r.progressLastWidth = width
	}

	progressView := model.Progress(&r.netease.Options().ProgressOptions, int(width), int(math.Round(width*float64(progress)/100)), r.progressRamp)

	//将passedDuration的值限制在allDuration的范围内
	displayDuration := passedDuration
	if displayDuration > allDuration {
		displayDuration = allDuration
	}

	var times string
	if allDuration/60 >= 100 {
		times = fmt.Sprintf("%03d:%02d/%03d:%02d", displayDuration/60, displayDuration%60, allDuration/60, allDuration%60)
	} else {
		times = fmt.Sprintf("%02d:%02d/%02d:%02d", displayDuration/60, displayDuration%60, allDuration/60, allDuration%60)
	}

	styledTimes := util.SetFgStyle(times, util.GetPrimaryColor())

	// Add an extra space for alignment if the duration is shorter.
	if allDuration/60 >= 100 {
		return progressView + " " + styledTimes, 1
	} else {
		return progressView + " " + styledTimes + " ", 1
	}
}
