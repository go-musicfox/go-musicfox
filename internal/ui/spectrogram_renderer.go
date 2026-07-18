package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
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
}

func NewSpectrogramRenderer(state *Player) *SpectrogramRenderer {
	provider, _ := state.Player.(player.SpectrumProvider)
	return &SpectrogramRenderer{provider: provider}
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

func (r *SpectrogramRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	width := a.WindowWidth()
	h := r.LineCount(a.WindowHeight(), main.MenuBottomRow())
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
	}

	speed := configs.AppConfig.Main.Visualizer.EffectiveSpectrogramSpeed()
	if speed <= 0 {
		speed = 1
	}
	if speed > width {
		speed = width
	}

	// Scroll history left by speed columns.
	if speed < width {
		for row := 0; row < height; row++ {
			copy(r.history[row][:width-speed], r.history[row][speed:])
		}
	}
	// Clear the rightmost speed columns.
	for col := width - speed; col < width; col++ {
		for row := 0; row < height; row++ {
			r.history[row][col] = 0
		}
	}

	// Draw new data on rightmost column(s).
	for col := width - speed; col < width; col++ {
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

	// Render to string with color ramp.
	start, end := model.GetProgressColor()
	ramp := util.MakeRamp(start, end, float64(width*2))

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
