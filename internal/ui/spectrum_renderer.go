package ui

import (
	"image/color"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"

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
	space := windowHeight - 5 - menuBottomRow
	if space < 4 {
		return spectrumLayout{}
	}
	lyricLines := min(5, max(0, space-4))
	barLines := space - lyricLines - 2
	if maxHeight := configs.AppConfig.Main.Visualizer.MaxBarHeight(); maxHeight > 0 {
		barLines = min(barLines, maxHeight)
	}
	return spectrumLayout{
		topPadding:    1,
		barLines:      barLines,
		bottomPadding: 1,
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
	for row := 0; row < height; row++ {
		level := spectrumRowLevel(frame, row, height)
		builder.WriteString(renderSpectrumBar(level, width, progressRamp, halfBlock, fullBlock, emptyBlock))
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
