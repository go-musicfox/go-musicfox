package ui

import (
	"image/color"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/lucasb-eyer/go-colorful"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

const (
	spectrumFullCharHalfBlock = '▌'
	spectrumFullCharFullBlock = '█'
	spectrumEmptyCharBlock    = ' '
)

type spectrumLayout struct {
	topPadding    int
	barLines      int
	bottomPadding int
}

func (l spectrumLayout) lines() int {
	return l.topPadding + l.barLines + l.bottomPadding
}

// SpectrumRenderer draws an animated, height-responsive PCM spectrum.
type SpectrumRenderer struct {
	provider          player.SpectrumProvider
	progressLastWidth float64
	progressRamp      []color.Color
	vertLastWidth     float64
	vertLastHeight    int
	vertRowRamps      [][]color.Color // per-row ramps for vertical gradient, nil when disabled
}

func NewSpectrumRenderer(state *Player) *SpectrumRenderer {
	provider, _ := state.Player.(player.SpectrumProvider)
	return &SpectrumRenderer{provider: provider}
}

func (r *SpectrumRenderer) IsEnabled() bool {
	return configs.AppConfig.Main.Visualizer.Enable && r.provider != nil
}

// LineCount reserves one blank row above and below the spectrum bars.
func (r *SpectrumRenderer) LineCount(windowHeight, menuBottomRow int) int {
	return r.layout(windowHeight, menuBottomRow).lines()
}

func (r *SpectrumRenderer) layout(windowHeight, menuBottomRow int) spectrumLayout {
	if !r.IsEnabled() {
		return spectrumLayout{}
	}

	var barLines int
	space := windowHeight - FixedTopBottomRows - menuBottomRow

	// Lyrics/cover take priority over spectrum; spectrum shrinks first.
	// Works for both static and dynamic menu modes — when the window is
	// small and DynamicMenuRows gives the menu fewer rows, spectrum
	// auto-shrinks or hides to free up space.
	neededLyricLines := 0
	if space >= FullLyricLines {
		neededLyricLines = FullLyricLines
	} else if space >= CompactLyricLines {
		neededLyricLines = CompactLyricLines
	}
	barLines = max(0, space-neededLyricLines-SpectrumReservedLines)
	if maxHeight := configs.AppConfig.Main.Visualizer.MaxBarHeight(); maxHeight > 0 {
		barLines = min(barLines, maxHeight)
	}

	if barLines == 0 {
		return spectrumLayout{}
	}
	return spectrumLayout{
		topPadding:    SpectrumVerticalPadding,
		barLines:      barLines,
		bottomPadding: SpectrumVerticalPadding,
	}
}

func (*SpectrumRenderer) Update(tea.Msg, *model.App) {}

func (r *SpectrumRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	width := a.WindowWidth()
	layout := r.layout(a.WindowHeight(), main.MenuBottomRow())
	if layout.barLines == 0 || width <= 0 {
		return "", 0
	}

	frame := r.provider.Spectrum()
	view = r.renderWithLayout(frame, width, layout)
	return view, layout.lines()
}

func (r *SpectrumRenderer) renderWithLayout(frame player.SpectrumFrame, width int, layout spectrumLayout) string {
	return strings.Repeat("\n", layout.topPadding) +
		r.render(frame, width, layout.barLines) +
		strings.Repeat("\n", layout.bottomPadding)
}

func (r *SpectrumRenderer) render(frame player.SpectrumFrame, width, height int) string {
	var builder strings.Builder
	builder.Grow((width + 1) * height)
	hasSignal := false
	for _, level := range frame.Levels {
		if level != 0 {
			hasSignal = true
			break
		}
	}
	if !hasSignal {
		blank := strings.Repeat(" ", width) + "\n"
		return strings.Repeat(blank, height)
	}

	halfBlock, fullBlock, emptyBlock := configs.AppConfig.Main.Visualizer.Characters()
	progressRamp := r.ramp(width)
	rowRamps := r.rowRamps(width, height)
	for row := 0; row < height; row++ {
		ramp := progressRamp
		if len(rowRamps) > 0 {
			ramp = rowRamps[row]
		}
		level := spectrumRowLevel(frame, row, height)
		builder.WriteString(renderSpectrumBar(level, width, ramp, halfBlock, fullBlock, emptyBlock))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (r *SpectrumRenderer) ramp(width int) []color.Color {
	if r.progressLastWidth != float64(width) || len(r.progressRamp) == 0 {
		start, end := model.GetProgressColor()
		r.progressRamp = util.MakeRamp(start, end, float64(width*2))
		r.progressLastWidth = float64(width)
	}
	return r.progressRamp
}

// rowRamps returns precomputed per-row horizontal ramps blended with a vertical
// gradient. When vertical gradient is disabled it returns nil and render() falls
// back to the plain horizontal ramp.
func (r *SpectrumRenderer) rowRamps(width, height int) [][]color.Color {
	if !configs.AppConfig.Main.Visualizer.VerticalGradient {
		return nil
	}
	if r.vertLastWidth == float64(width) && r.vertLastHeight == height && len(r.vertRowRamps) > 0 {
		return r.vertRowRamps
	}
	if height <= 1 {
		return nil
	}

	baseRamp := r.ramp(width)

	// Generate independent color pair for vertical gradient with wider color span.
	vertStart, vertEnd := util.GetRandomRgbColor(true)
	cStart, _ := colorful.Hex(vertStart)
	cEnd, _ := colorful.Hex(vertEnd)

	r.vertRowRamps = make([][]color.Color, height)
	for row := 0; row < height; row++ {
		t := float64(row) / float64(height-1) // 0=top, 1=bottom
		vColor := cStart.BlendLuv(cEnd, t)    // top→start, bottom→end

		rowRamp := make([]color.Color, len(baseRamp))
		for i, c := range baseRamp {
			hColor, _ := colorful.MakeColor(c)
			rowRamp[i] = hColor.BlendLuv(vColor, 0.65)
		}
		r.vertRowRamps[row] = rowRamp
	}
	r.vertLastWidth = float64(width)
	r.vertLastHeight = height
	return r.vertRowRamps
}

func renderSpectrumBar(level float64, width int, progressRamp []color.Color, halfBlock, fullBlock, emptyBlock rune) string {
	fillUnits := spectrumFillUnits(level, width)
	fullChars := fillUnits / 2
	hasHalfChar := fillUnits%2 == 1

	var builder strings.Builder
	builder.Grow(width)
	for column := 0; column < fullChars; column++ {
		builder.WriteString(util.SetFgStyle(string(fullBlock), progressRamp[column*2]))
	}
	if hasHalfChar {
		builder.WriteString(spectrumHalfBlockStyle(
			progressRamp[fullChars*2],
			progressRamp[fullChars*2+1],
		).Render(string(halfBlock)))
	}
	emptyChars := width - fullChars
	if hasHalfChar {
		emptyChars--
	}
	// Leave empty cells unstyled to preserve the terminal background.
	builder.WriteString(strings.Repeat(string(emptyBlock), emptyChars))
	return builder.String()
}

func spectrumHalfBlockStyle(foreground, background color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(foreground).Background(background)
}

func spectrumRowLevel(frame player.SpectrumFrame, row, height int) float64 {
	group := height - row - 1
	start := group * player.SpectrumBandCount / height
	end := (group + 1) * player.SpectrumBandCount / height
	level := 0.0
	for band := start; band < end; band++ {
		level = max(level, frame.Levels[band])
	}
	return clamp(level, 0, 1)
}

func spectrumFillUnits(level float64, width int) int {
	return int(math.Round(clamp(level, 0, 1) * float64(width*2)))
}
