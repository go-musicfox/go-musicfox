package ui

import (
	"fmt"
	"math"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/go-musicfox/go-musicfox/internal/configs"
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

func (r *ProgressRenderer) getRenderMode() progressRenderMode {
	if configs.AppConfig.Theme.Progress.RenderMode != "" {
		switch configs.AppConfig.Theme.Progress.RenderMode {
		case "smooth":
			return progressRenderModeSmooth
		case "wave":
			return progressRenderModeWave
		case "glow":
			return progressRenderModeGlow
		}
	}

	return progressRenderModeSmooth
}

// View renders the progress bar component.
func (r *ProgressRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	song := r.state.CurSong()
	allDuration := int(song.Duration.Seconds())
	if allDuration == 0 {
		r.progressRamp = nil
		return "", 1
	}

	passedDuration := int(r.state.PassedTime().Seconds())
	displayDuration := passedDuration
	if displayDuration > allDuration {
		displayDuration = allDuration
	}

	var progressPct int
	if passedDuration > allDuration {
		progressPct = 100
	} else {
		progressPct = passedDuration * 100 / allDuration
	}
	progress := float64(progressPct) / 100.0

	width := r.netease.WindowWidth() - 14
	if width < 0 {
		width = 0
	}
	fullSize := int(math.Round(float64(width) * progress))

	progressOptions := configs.AppConfig.Theme.Progress.ToModel()
	mode := r.getRenderMode()
	animationTime := r.state.PassedTime().Seconds()

	var progressView string
	switch mode {
	case progressRenderModeWave, progressRenderModeGlow:
		ramp := progressRampForMode(width, fullSize, animationTime, mode)
		progressView = model.Progress(&progressOptions, width, fullSize, ramp)
	case progressRenderModeSmooth:
		fallthrough
	default:
		start, end := model.GetProgressColor()
		if float64(width) != r.progressLastWidth || len(r.progressRamp) == 0 {
			r.progressRamp = util.MakeRamp(start, end, float64(width))
			r.progressLastWidth = float64(width)
		}
		progressView = model.Progress(&progressOptions, width, fullSize, r.progressRamp)
	}

	var times string
	if allDuration/60 >= 100 {
		times = fmt.Sprintf("%03d:%02d/%03d:%02d", displayDuration/60, displayDuration%60, allDuration/60, allDuration%60)
	} else {
		times = fmt.Sprintf("%02d:%02d/%02d:%02d", displayDuration/60, displayDuration%60, allDuration/60, allDuration%60)
	}
	styledTimes := util.SetFgStyle(times, util.GetPrimaryColor())

	if allDuration/60 >= 100 {
		return progressView + " " + styledTimes, 1
	}
	return progressView + " " + styledTimes + " ", 1
}
