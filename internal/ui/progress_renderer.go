package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// ProgressRenderer is a dedicated UI component for rendering the playback progress bar.
type ProgressRenderer struct {
	svc   *menuServices
	state playerRendererState

	progressLastWidth float64
	progressRamp      []color.Color

	cachedView      string
	cachedLines     int
	cachedPassedSec int // rounded to seconds
	cachedDuration  int // total seconds
	cachedWidth     int
	cachedStyleGen  uint64
}

// NewProgressRenderer creates a new progress bar renderer component.
func NewProgressRenderer(svc *menuServices, state playerRendererState) *ProgressRenderer {
	return &ProgressRenderer{
		svc:   svc,
		state: state,
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

	width := r.svc.App().WindowWidth() - ProgressTimeDisplayWidth
	if width < 0 {
		width = 0
	}

	// Output caching: skip rebuild when progress has not ticked to the next second.
	// styleGen guards against replaying stale-colored output after a theme switch.
	styleGen := style.StyleGeneration()
	if passedDuration == r.cachedPassedSec && allDuration == r.cachedDuration &&
		width == r.cachedWidth && styleGen == r.cachedStyleGen {
		return r.cachedView, r.cachedLines
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
		ramp := r.progressRamp
		progressView = model.Progress(&progressOptions, width, fullSize, ramp)
	}

	var times string
	if allDuration/60 >= ProgressLongDurationThreshold {
		times = fmt.Sprintf("%03d:%02d/%03d:%02d", displayDuration/60, displayDuration%60, allDuration/60, allDuration%60)
	} else {
		times = fmt.Sprintf("%02d:%02d/%02d:%02d", displayDuration/60, displayDuration%60, allDuration/60, allDuration%60)
	}
	// Paint the time text glyphs and every filler cell with the app background
	// resolved from the SAME (global) StyleSet that model.Progress uses for the
	// bar cells; mixing an app-scoped StyleSet here would leave the separator and
	// trailing fill transparent whenever the two stylesets diverge.
	appBg := style.CurrentStyleSet().AppBackground
	timesBg := configs.ResolveBackground(nil)
	var styledTimes string
	if timesBg != nil {
		styledTimes = style.FGBG(times, util.GetPrimaryColor(), timesBg)
	} else {
		styledTimes = style.FG(times, util.GetPrimaryColor())
	}

	view = progressView + appBg.Render(" ") + styledTimes
	if allDuration/60 < ProgressLongDurationThreshold {
		view += appBg.Render(" ")
	}
	viewWidth := runewidth.StringWidth(stripAnsiCodes(view))
	windowWidth := r.svc.App().WindowWidth()
	remainingWidth := windowWidth - viewWidth
	if remainingWidth > 0 {
		view += appBg.Render(strings.Repeat(" ", remainingWidth))
	}

	// Store output cache
	r.cachedView = view
	r.cachedLines = 1
	r.cachedPassedSec = passedDuration
	r.cachedDuration = allDuration
	r.cachedWidth = width
	r.cachedStyleGen = styleGen

	return r.cachedView, r.cachedLines
}
