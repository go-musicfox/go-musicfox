package ui

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	foxfulStyle "github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

// spectrogramChars is a character ramp from empty (space) to full block,
// with 8 intermediate levels for smooth gradient display.
var spectrogramChars = []rune(" ▁▂▃▄▅▆▇█")

// spectrogramChar maps a byte level [0,255] to a display character.
func spectrogramChar(level byte) rune {
	idx := int(level) * (len(spectrogramChars) - 1) / 255
	if idx < 0 {
		idx = 0
	}
	if idx >= len(spectrogramChars) {
		idx = len(spectrogramChars) - 1
	}
	return spectrogramChars[idx]
}

// SpectrogramRenderer draws a scrolling spectrogram (cava-inspired).
// New spectrum data appears on the right and scrolls leftward over time.
type SpectrogramRenderer struct {
	provider player.SpectrumProvider
	history  [][]byte // [row][col]: row=frequency band, col=time

	lastViewTime time.Time        // last View() render timestamp for time-based scrolling
	scrollFrac   float64          // fractional columns accumulated between renders
	nowFn        func() time.Time // injectable clock for tests

	rampLastW    int
	rampStyleGen uint64
	rampCache    []color.Color
}

func NewSpectrogramRenderer(state *Player) *SpectrogramRenderer {
	provider, _ := state.Player.(player.SpectrumProvider)
	return &SpectrogramRenderer{provider: provider, nowFn: time.Now}
}

func (r *SpectrogramRenderer) IsEnabled() bool {
	return configs.AppConfig.Main.Visualizer.Enable &&
		configs.AppConfig.Main.Visualizer.Style == "spectrogram" &&
		r.provider != nil
}

// LineCount returns the number of terminal rows consumed by the spectrogram.
// Reuses the same layout logic as SpectrumRenderer.layout().
func (r *SpectrogramRenderer) LineCount(windowHeight, menuBottomRow int) int {
	if !r.IsEnabled() {
		return 0
	}
	space := windowHeight - FixedTopBottomRows - menuBottomRow
	neededLyricLines := 0
	if space >= FullLyricLines {
		neededLyricLines = FullLyricLines
	} else if space >= CompactLyricLines {
		neededLyricLines = CompactLyricLines
	}
	barLines := max(0, space-neededLyricLines-SpectrumReservedLines)
	if maxHeight := configs.AppConfig.Main.Visualizer.MaxBarHeight(); maxHeight > 0 {
		barLines = min(barLines, maxHeight)
	}
	if barLines == 0 {
		return 0
	}
	return SpectrumVerticalPadding + barLines
}

func (*SpectrogramRenderer) Update(tea.Msg, *model.App) {}

// advanceScroll converts elapsed wall time into whole columns to scroll, at
// speed*5 columns per second (speed is the historical per-frame advance at
// 5 FPS). Fractional columns accumulate across calls, so the flow rate is
// independent of the render frame rate. Returns 0 on the first frame or after
// a long gap (playback paused), keeping the history frozen.
func (r *SpectrogramRenderer) advanceScroll(elapsed, interval time.Duration, speed, width int) int {
	if elapsed <= 0 || elapsed > 2*interval {
		return 0
	}
	r.scrollFrac += float64(speed) * 5 * elapsed.Seconds()
	cols := int(r.scrollFrac)
	if cols > width {
		cols = width
	}
	r.scrollFrac -= float64(cols)
	return cols
}

// scrollAndDraw shifts the history left by scrollCols columns and stamps the
// current frame into the rightmost scrollCols columns. Extracted from View so
// it can be unit-tested with a fixed frame and elapsed time.
func (r *SpectrogramRenderer) scrollAndDraw(frame player.SpectrumFrame, height, width, scrollCols int) {
	if height <= 0 || width <= 0 {
		return
	}
	if scrollCols < width {
		for row := 0; row < height; row++ {
			copy(r.history[row][:width-scrollCols], r.history[row][scrollCols:])
		}
	}
	// Clear the rightmost scrollCols columns.
	for col := width - scrollCols; col < width; col++ {
		for row := 0; row < height; row++ {
			r.history[row][col] = 0
		}
	}

	// Draw new data on the rightmost column(s).
	for col := width - scrollCols; col < width; col++ {
		for row := 0; row < height; row++ {
			startBand := row * player.SpectrumBandCount / height
			endBand := (row + 1) * player.SpectrumBandCount / height
			if endBand > player.SpectrumBandCount {
				endBand = player.SpectrumBandCount
			}
			level := 0.0
			for band := startBand; band < endBand; band++ {
				if frame.Levels[band] > level {
					level = frame.Levels[band]
				}
			}
			r.history[row][col] = byte(clamp(level, 0, 1) * 255)
		}
	}
}

func (r *SpectrogramRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	width := a.WindowWidth()
	h := r.LineCount(main.EffectiveWindowHeight(a), main.MenuBottomRow())
	if h <= SpectrumVerticalPadding || width <= 0 {
		return "", 0
	}
	height := h - SpectrumVerticalPadding

	frame := r.provider.Spectrum()

	// Ensure history buffer matches current dimensions.
	if len(r.history) != height || (len(r.history) > 0 && len(r.history[0]) != width) {
		r.history = make([][]byte, height)
		for i := range r.history {
			r.history[i] = make([]byte, width)
		}
		r.scrollFrac = 0
	}

	speed := configs.AppConfig.Main.Visualizer.EffectiveSpectrogramSpeed()
	if speed <= 0 {
		speed = 1
	}
	if speed > width {
		speed = width
	}

	// Scroll is time-based: `speed` means columns per 200ms (the historical
	// 5 FPS per-frame advance), so the flow rate stays constant regardless of
	// the configured frame rate. Fractional columns accumulate between frames.
	now := r.nowFn()
	elapsed := time.Duration(0)
	if !r.lastViewTime.IsZero() {
		elapsed = now.Sub(r.lastViewTime)
	} else {
		// First frame: advance one nominal frame interval so spectrum data
		// appears immediately, matching the historical per-frame behavior.
		elapsed = configs.AppConfig.Main.FrameRate.Interval()
	}
	r.lastViewTime = now

	scrollCols := r.advanceScroll(elapsed, configs.AppConfig.Main.FrameRate.Interval(), speed, width)
	r.scrollAndDraw(frame, height, width, scrollCols)

	// Render to string with color ramp.
	// The ramp depends only on width and theme colors, so cache it per
	// (width, style generation) instead of rebuilding it every frame.
	gen := foxfulStyle.StyleGeneration()
	if r.rampLastW != width || r.rampStyleGen != gen || len(r.rampCache) == 0 {
		start, end := model.GetProgressColor()
		r.rampCache = util.MakeRamp(start, end, float64(width*2))
		r.rampLastW = width
		r.rampStyleGen = gen
	}
	ramp := r.rampCache

	var builder strings.Builder
	builder.WriteString(strings.Repeat("\n", SpectrumVerticalPadding))

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			lvl := r.history[row][col]
			if lvl == 0 {
				builder.WriteByte(' ')
			} else {
				ch := spectrogramChar(lvl)
				builder.WriteString(util.SetFgStyle(string(ch), ramp[col*2]))
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String(), h
}
