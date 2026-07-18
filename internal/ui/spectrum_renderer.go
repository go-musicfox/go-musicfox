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

// SpectrumStyle defines the visual rendering mode.
type SpectrumStyle string

const (
	SpectrumStyleBar          SpectrumStyle = "bar"
	SpectrumStyleLine         SpectrumStyle = "line"
	SpectrumStyleMirrorBar    SpectrumStyle = "mirror_bar"
	SpectrumStyleDot          SpectrumStyle = "dot"
	SpectrumStyleOscilloscope SpectrumStyle = "oscilloscope"
	SpectrumStyleVectorscope  SpectrumStyle = "vectorscope"
)

const (
	spectrumFullCharHalfBlock = '▌'
	spectrumFullCharFullBlock = '█'
	spectrumEmptyCharBlock    = ' '
)

// barOrientation controls the growth direction of bar-style spectrum bars.
type barOrientation int

const (
	barOrientBottom     barOrientation = iota // bars grow upward from bottom (default)
	barOrientTop                              // bars grow downward from top
	barOrientLeft                             // bars grow rightward from left edge
	barOrientRight                            // bars grow leftward from right edge
	barOrientHorizontal                       // mirror: center out, top+bottom
	barOrientVertical                         // mirror: center out, left+right
)

func (r *SpectrumRenderer) barOrientation() barOrientation {
	switch configs.AppConfig.Main.Visualizer.EffectiveBarOrientation() {
	case "top":
		return barOrientTop
	case "left":
		return barOrientLeft
	case "right":
		return barOrientRight
	case "horizontal":
		return barOrientHorizontal
	case "vertical":
		return barOrientVertical
	default:
		return barOrientBottom
	}
}

// --- Braille rendering helpers ---
//
// Each terminal cell can display one Unicode braille character (U+2800–U+28FF).
// A braille cell represents a 2-column × 4-row sub-pixel grid.
// Dot numbering (bit N = dot N+1):
//
//	(0,0) (1,0)      1 4
//	(0,1) (1,1)  →   2 5
//	(0,2) (1,2)      3 6
//	(0,3) (1,3)      7 8

var brailleDotBits = [2][4]byte{
	{0x01, 0x02, 0x04, 0x40}, // left column  (x=0): dots 1,2,3,7
	{0x08, 0x10, 0x20, 0x80}, // right column (x=1): dots 4,5,6,8
}

func brailleCell(dots byte) rune {
	return '\u2800' + rune(dots)
}

// setBrailleDot lights a dot at (subCol, subRow) on the braille grid.
// subCol is in [0, width*2-1], subRow in [0, height*4-1].
func setBrailleDot(grid [][]byte, subCol, subRow int) {
	if subCol < 0 || subRow < 0 {
		return
	}
	col := subCol / 2
	row := subRow / 4
	if row >= len(grid) || col >= len(grid[0]) {
		return
	}
	x := subCol % 2
	y := subRow % 4
	grid[row][col] |= brailleDotBits[x][y]
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// drawBrailleLine draws a Bresenham line segment on the braille grid.
func drawBrailleLine(grid [][]byte, x1, y1, x2, y2 int) {
	dx := absInt(x2 - x1)
	dy := -absInt(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx + dy

	maxCols := len(grid[0])*2 - 1
	maxRows := len(grid)*4 - 1
	for {
		if x1 >= 0 && x1 <= maxCols && y1 >= 0 && y1 <= maxRows {
			setBrailleDot(grid, x1, y1)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

// --- Layout ---

type spectrumLayout struct {
	topPadding    int
	barLines      int
	bottomPadding int
}

func (l spectrumLayout) lines() int {
	return l.topPadding + l.barLines + l.bottomPadding
}

// --- Renderer ---

// SpectrumRenderer draws an animated, height-responsive PCM spectrum.
type SpectrumRenderer struct {
	provider             player.SpectrumProvider
	progressLastWidth    float64
	progressRamp         []color.Color
	progressDimLastWidth float64
	progressDimRamp      []color.Color
	vertLastWidth        float64
	vertLastHeight       int
	vertRowRamps         [][]color.Color
	brailleGridLCache    [][]byte
	brailleGridRCache    [][]byte
	brailleGridMCache    [][]byte
	brailleCacheW        int
	brailleCacheH        int
	phaseMask            []float64 // per-column phase correlation, set per frame when SpectrumPhaseDiff enabled
}

func NewSpectrumRenderer(state *Player) *SpectrumRenderer {
	provider, _ := state.Player.(player.SpectrumProvider)
	return &SpectrumRenderer{provider: provider}
}

func (r *SpectrumRenderer) IsEnabled() bool {
	cfg := configs.AppConfig.Main.Visualizer
	return cfg.Enable && !cfg.IsSpectrogram() && r.provider != nil
}

func (r *SpectrumRenderer) LineCount(windowHeight, menuBottomRow int) int {
	return r.layout(windowHeight, menuBottomRow).lines()
}

func (r *SpectrumRenderer) layout(windowHeight, menuBottomRow int) spectrumLayout {
	if !r.IsEnabled() {
		return spectrumLayout{}
	}

	var barLines int
	space := windowHeight - FixedTopBottomRows - menuBottomRow

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
		bottomPadding: 0,
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

// render dispatches to the configured style.
func (r *SpectrumRenderer) render(frame player.SpectrumFrame, width, height int) string {
	style := SpectrumStyle(configs.AppConfig.Main.Visualizer.Style)
	isLineOrDot := style == SpectrumStyleLine || style == SpectrumStyleDot
	isRawSample := style == SpectrumStyleOscilloscope || style == SpectrumStyleVectorscope

	if !isLineOrDot && !isRawSample && !hasSignal(frame) {
		blank := strings.Repeat(" ", width) + "\n"
		return strings.Repeat(blank, height)
	}

	// Compute per-column phase correlation mask for dual-channel spectrum styles.
	if configs.AppConfig.Main.Visualizer.SpectrumPhaseDiff &&
		!configs.AppConfig.Main.Visualizer.IsMono() &&
		isLineOrDot {
		r.phaseMask = computePhaseMask(frame.PhasesL, frame.PhasesR, width)
	} else {
		r.phaseMask = nil
	}

	switch style {
	case SpectrumStyleLine:
		if configs.AppConfig.Main.Visualizer.LineMode == "block" {
			return r.renderLineBlock(r.lineFullChar(), r.lineHalfChar(), r.lineEmptyChar(), frame, width, height)
		}
		return r.renderLine(frame, width, height)
	case SpectrumStyleDot:
		if configs.AppConfig.Main.Visualizer.DotMode == "block" {
			return r.renderDotBlock(r.dotFullChar(), r.dotHalfChar(), r.dotEmptyChar(), frame, width, height)
		}
		return r.renderDot(frame, width, height)
	case SpectrumStyleOscilloscope:
		rawFrame := r.provider.RawSamples()
		if configs.AppConfig.Main.Visualizer.OscilloscopeMode == "block" {
			return r.renderOscilloscopeBlock(rawFrame, width, height)
		}
		return r.renderOscilloscopeBraille(rawFrame, width, height)
	case SpectrumStyleVectorscope:
		rawFrame := r.provider.RawSamples()
		if configs.AppConfig.Main.Visualizer.VectorscopeMode == "block" {
			return r.renderVectorscopeBlock(rawFrame, width, height)
		}
		return r.renderVectorscopeBraille(rawFrame, width, height)
	case SpectrumStyleMirrorBar:
		return r.renderMirror(frame, width, height)
	default:
		return r.renderBar(frame, width, height)
	}
}

func (r *SpectrumRenderer) lineFullChar() string {
	return firstCharOf(configs.AppConfig.Main.Visualizer.LineFullBlock)
}

func (r *SpectrumRenderer) lineHalfChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.LineHalfBlock)
	if ch == "" {
		ch = r.lineFullChar()
	}
	return ch
}

func (r *SpectrumRenderer) lineEmptyChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.LineEmptyBlock)
	if ch == "" {
		ch = " "
	}
	return ch
}

func (r *SpectrumRenderer) dotFullChar() string {
	return firstCharOf(configs.AppConfig.Main.Visualizer.DotFullBlock)
}

func (r *SpectrumRenderer) dotHalfChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.DotHalfBlock)
	if ch == "" {
		ch = r.dotFullChar()
	}
	return ch
}

func (r *SpectrumRenderer) dotEmptyChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.DotEmptyBlock)
	if ch == "" {
		ch = " "
	}
	return ch
}

func firstCharOf(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

// --- Style: bar (original horizontal progress bars) ---

func (r *SpectrumRenderer) renderBar(frame player.SpectrumFrame, width, height int) string {
	orient := r.barOrientation()
	switch orient {
	case barOrientLeft:
		return r.renderBarVertical(frame, width, height, false)
	case barOrientRight:
		return r.renderBarVertical(frame, width, height, true)
	case barOrientHorizontal:
		return r.renderBarHorizontal(frame, width, height)
	case barOrientVertical:
		return r.renderBarVerticalCenter(frame, width, height)
	default:
		// bottom (default) and top
		content := r.renderBarBottom(frame, width, height)
		if orient == barOrientTop {
			lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
			for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
				lines[i], lines[j] = lines[j], lines[i]
			}
			content = strings.Join(lines, "\n") + "\n"
		}
		return content
	}
}

// renderBarBottom renders horizontal bars growing left-to-right from bottom row upward.
func (r *SpectrumRenderer) renderBarBottom(frame player.SpectrumFrame, width, height int) string {
	cfg := configs.AppConfig.Main.Visualizer
	halfBlock, fullBlock, emptyBlock := cfg.BarCharacters()
	progressRamp := r.ramp(width)
	vertEnabled := cfg.BarVerticalGradient
	horizEnabled := cfg.BarHorizontalGradient

	var rowRamps [][]color.Color
	if vertEnabled {
		rowRamps = r.rowRamps(width, height, true)
	}

	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		ramp := progressRamp
		if len(rowRamps) > 0 {
			ramp = rowRamps[row]
		}

		// Horizontal gradient: compute per-row color or blend with vertical ramp
		if horizEnabled {
			horizColor := r.horizontalColor(row, height)
			if !vertEnabled {
				// Only horizontal: solid-color ramp
				ramp = r.horizontalBarRamp(row, width, height)
			} else {
				// Both: blend horizontal into vertical ramp
				ramp = blendRamps(ramp, horizColor)
			}
		}

		level := spectrumRowLevel(frame, row, height)

		// Idle bar head: show tiny cap when bar is nearly silent
		if cfg.ShowIdleBarHeads && level > 0 && level <= 0.01 {
			c := ramp[0]
			builder.WriteString(util.SetFgStyle(string(halfBlock), c))
			builder.WriteString(strings.Repeat(string(emptyBlock), width-1))
			builder.WriteByte('\n')
			continue
		}

		builder.WriteString(renderSpectrumBar(level, width, ramp, halfBlock, fullBlock, emptyBlock))
		builder.WriteByte('\n')
	}
	return builder.String()
}

// --- Style: mirror_bar (bars mirrored left/right from center) ---

func (r *SpectrumRenderer) renderMirror(frame player.SpectrumFrame, width, height int) string {
	cfg := configs.AppConfig.Main.Visualizer
	halfBlock, fullBlock, emptyBlock := cfg.MirrorBarCharacters()
	progressRamp := r.ramp(width)
	vertEnabled := cfg.BarVerticalGradient
	horizEnabled := cfg.BarHorizontalGradient

	var rowRamps [][]color.Color
	if vertEnabled {
		rowRamps = r.rowRamps(width, height, true)
	}

	mono := cfg.IsMono()
	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		ramp := progressRamp
		if len(rowRamps) > 0 {
			ramp = rowRamps[row]
		}

		// Horizontal gradient: compute per-row color or blend with vertical ramp
		if horizEnabled {
			horizColor := r.horizontalColor(row, height)
			if !vertEnabled {
				ramp = r.horizontalBarRamp(row, width, height)
			} else {
				ramp = blendRamps(ramp, horizColor)
			}
		}

		var levelL, levelR float64
		if mono {
			level := spectrumRowLevel(frame, row, height)
			levelL, levelR = level, level
		} else {
			levelL = spectrumRowLevelFrom(frame.LevelsL, row, height)
			levelR = spectrumRowLevelFrom(frame.LevelsR, row, height)
		}

		// Idle bar heads for mirror: show halfBlock at center from each side
		bothIdle := cfg.ShowIdleBarHeads && levelL > 0 && levelL <= 0.01 && levelR > 0 && levelR <= 0.01
		leftIdle := cfg.ShowIdleBarHeads && levelL > 0 && levelL <= 0.01 && !(levelR > 0 && levelR <= 0.01)
		rightIdle := cfg.ShowIdleBarHeads && levelR > 0 && levelR <= 0.01 && !(levelL > 0 && levelL <= 0.01)

		if bothIdle {
			builder.WriteString(renderMirrorIdleLineDual(width, ramp, halfBlock, emptyBlock, true, true))
			builder.WriteByte('\n')
			continue
		}
		if leftIdle {
			builder.WriteString(renderMirrorIdleLineDual(width, ramp, halfBlock, emptyBlock, true, false))
			builder.WriteByte('\n')
			continue
		}
		if rightIdle {
			builder.WriteString(renderMirrorIdleLineDual(width, ramp, halfBlock, emptyBlock, false, true))
			builder.WriteByte('\n')
			continue
		}

		builder.WriteString(renderMirrorBarLineDual(levelL, levelR, width, ramp, halfBlock, fullBlock, emptyBlock))
		builder.WriteByte('\n')
	}
	return builder.String()
}

// --- Style: line (braille-connected frequency curve) ---

func (r *SpectrumRenderer) renderLine(frame player.SpectrumFrame, width, height int) string {
	if configs.AppConfig.Main.Visualizer.IsMono() {
		return r.renderLineMono(frame, width, height)
	}
	return r.renderLineDual(frame, width, height, true)
}

// --- Style: dot (braille scatter plot) ---

func (r *SpectrumRenderer) renderDot(frame player.SpectrumFrame, width, height int) string {
	if configs.AppConfig.Main.Visualizer.IsMono() {
		return r.renderDotMono(frame, width, height)
	}
	return r.renderLineDual(frame, width, height, false)
}

// renderLineDual draws L and R channels on the same braille grid.
func (r *SpectrumRenderer) renderLineDual(frame player.SpectrumFrame, width, height int, connect bool) string {
	gridL := r.getBrailleGrid(&r.brailleGridLCache, width, height)
	r.buildBrailleGrid(frame.LevelsL, width, height, connect, gridL)
	gridR := r.getBrailleGrid(&r.brailleGridRCache, width, height)
	r.buildBrailleGrid(frame.LevelsR, width, height, connect, gridR)
	return r.renderBrailleGridDual(gridL, gridR, width, height)
}

// renderBrailleGridDual merges L and R braille grids with vertical gradient.
func (r *SpectrumRenderer) renderBrailleGridDual(gridL, gridR [][]byte, width, height int) string {
	ramp := r.ramp(width)
	rampDim := r.rampDim(width)
	vRamp := r.vertBrailleRamps(width, height)
	vRampDim := r.vertBrailleRampsDim(width, height, rampDim)

	hasR := false
	for row := 0; row < height && !hasR; row++ {
		for col := 0; col < width && !hasR; col++ {
			if gridR[row][col] != 0 {
				hasR = true
			}
		}
	}

	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		rowRamp := ramp
		rowRampDim := rampDim
		if len(vRamp) > 0 {
			rowRamp = vRamp[row]
			rowRampDim = vRampDim[row]
		}
		for col := 0; col < width; col++ {
			dotsL := gridL[row][col]
			dotsR := gridR[row][col]
			dots := dotsL | dotsR
			if dots == 0 {
				builder.WriteByte(' ')
				continue
			}
			c := rowRamp[col*2]
			if hasR && dotsR != 0 {
				c = rowRampDim[col*2]
			}
			// Apply phase correlation coloring when enabled.
			if len(r.phaseMask) > 0 && hasR && col < len(r.phaseMask) {
				c = blendPhaseColor(c, r.phaseMask[col])
			}
			builder.WriteString(util.SetFgStyle(string(brailleCell(dots)), c))
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// vertBrailleRamps returns per-row ramps blending the horizontal gradient
// with a top→bottom brightness shift. Top rows are slightly darker.
func (r *SpectrumRenderer) vertBrailleRamps(width, height int) [][]color.Color {
	if height <= 1 {
		return nil
	}
	base := r.ramp(width)
	black, _ := colorful.Hex("#000000")

	ramps := make([][]color.Color, height)
	for row := 0; row < height; row++ {
		// 0.0=top(dark) → 1.0=bottom(bright)
		t := float64(row) / float64(height-1)
		blend := 0.35 * (1.0 - t) // top blends 35% black, bottom 0%
		rowRamp := make([]color.Color, len(base))
		for i, c := range base {
			h, _ := colorful.MakeColor(c)
			rowRamp[i] = h.BlendLuv(black, blend)
		}
		ramps[row] = rowRamp
	}
	return ramps
}

// vertBrailleRampsDim applies the same vertical gradient to the shifted ramp.
func (r *SpectrumRenderer) vertBrailleRampsDim(width, height int, base []color.Color) [][]color.Color {
	if height <= 1 {
		return nil
	}
	black, _ := colorful.Hex("#000000")

	ramps := make([][]color.Color, height)
	for row := 0; row < height; row++ {
		t := float64(row) / float64(height-1)
		blend := 0.35 * (1.0 - t)
		rowRamp := make([]color.Color, len(base))
		for i, c := range base {
			h, _ := colorful.MakeColor(c)
			rowRamp[i] = h.BlendLuv(black, blend)
		}
		ramps[row] = rowRamp
	}
	return ramps
}

// --- Style: line/dot braille mono ---

// renderLineMono draws a single-channel braille frequency curve from combined Levels.
func (r *SpectrumRenderer) renderLineMono(frame player.SpectrumFrame, width, height int) string {
	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)
	r.buildBrailleGrid(frame.Levels, width, height, true, grid)
	return r.renderBrailleGridMono(grid, width, height)
}

// renderDotMono draws a single-channel braille scatter plot from combined Levels.
func (r *SpectrumRenderer) renderDotMono(frame player.SpectrumFrame, width, height int) string {
	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)
	r.buildBrailleGrid(frame.Levels, width, height, false, grid)
	return r.renderBrailleGridMono(grid, width, height)
}

// renderBrailleGridMono renders a single braille grid with the primary color ramp.
func (r *SpectrumRenderer) renderBrailleGridMono(grid [][]byte, width, height int) string {
	ramp := r.ramp(width)
	vRamp := r.vertBrailleRamps(width, height)

	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		rowRamp := ramp
		if len(vRamp) > 0 {
			rowRamp = vRamp[row]
		}
		for col := 0; col < width; col++ {
			dots := grid[row][col]
			if dots == 0 {
				builder.WriteByte(' ')
				continue
			}
			c := rowRamp[col*2]
			builder.WriteString(util.SetFgStyle(string(brailleCell(dots)), c))
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// vertBlockRamps / vertBlockRampsDim reuse the same vertical gradient logic.
func (r *SpectrumRenderer) vertBlockRamps(width, height int) [][]color.Color {
	return r.vertBrailleRamps(width, height)
}

func (r *SpectrumRenderer) vertBlockRampsDim(width, height int, base []color.Color) [][]color.Color {
	return r.vertBrailleRampsDim(width, height, base)
}

// --- Style: line/dot block mode ---

// renderLineBlock draws L+R frequency curves on a character grid.
func (r *SpectrumRenderer) renderLineBlock(fullChar, halfChar, emptyChar string, frame player.SpectrumFrame, width, height int) string {
	if configs.AppConfig.Main.Visualizer.IsMono() {
		return r.renderBlockMono(fullChar, halfChar, emptyChar, frame.Levels, width, height, true)
	}
	return r.renderBlockDual(fullChar, halfChar, emptyChar, frame.LevelsL, frame.LevelsR, width, height, true)
}

// renderDotBlock draws L+R scatter dots on a character grid.
func (r *SpectrumRenderer) renderDotBlock(fullChar, halfChar, emptyChar string, frame player.SpectrumFrame, width, height int) string {
	if configs.AppConfig.Main.Visualizer.IsMono() {
		return r.renderBlockMono(fullChar, halfChar, emptyChar, frame.Levels, width, height, false)
	}
	return r.renderBlockDual(fullChar, halfChar, emptyChar, frame.LevelsL, frame.LevelsR, width, height, false)
}

func (r *SpectrumRenderer) renderBlockDual(fullChar, halfChar, emptyChar string, levelsL, levelsR [player.SpectrumBandCount]float64, width, height int, connect bool) string {
	if halfChar == "" {
		halfChar = fullChar
	}
	ramp := r.ramp(width)
	rampDim := r.rampDim(width)
	vRamp := r.vertBlockRamps(width, height)
	vRampDim := r.vertBlockRampsDim(width, height, rampDim)

	hasR := false
	for _, v := range levelsR {
		if v > 1e-9 {
			hasR = true
			break
		}
	}

	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)

	drawBlockChannel(grid, levelsL, width, height, connect, 2)
	if hasR {
		drawBlockChannel(grid, levelsR, width, height, connect, 1)
	}

	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		rowRamp := ramp
		rowRampDim := rampDim
		if len(vRamp) > 0 {
			rowRamp = vRamp[row]
			rowRampDim = vRampDim[row]
		}
		for col := 0; col < width; col++ {
			v := grid[row][col]
			switch {
			case v == 3:
				c := rowRampDim[col*2]
				if len(r.phaseMask) > 0 && hasR && col < len(r.phaseMask) {
					c = blendPhaseColor(c, r.phaseMask[col])
				}
				builder.WriteString(util.SetFgStyle(fullChar, c))
			case v == 2:
				builder.WriteString(util.SetFgStyle(fullChar, rowRamp[col*2]))
			case v == 1:
				c := rowRampDim[col*2]
				if len(r.phaseMask) > 0 && hasR && col < len(r.phaseMask) {
					c = blendPhaseColor(c, r.phaseMask[col])
				}
				builder.WriteString(util.SetFgStyle(halfChar, c))
			default:
				builder.WriteString(emptyChar)
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// renderBlockMono draws a single channel on a character grid using the primary color ramp.
func (r *SpectrumRenderer) renderBlockMono(fullChar, halfChar, emptyChar string, levels [player.SpectrumBandCount]float64, width, height int, connect bool) string {
	if halfChar == "" {
		halfChar = fullChar
	}
	ramp := r.ramp(width)
	vRamp := r.vertBlockRamps(width, height)

	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)
	drawBlockChannel(grid, levels, width, height, connect, 2)

	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		rowRamp := ramp
		if len(vRamp) > 0 {
			rowRamp = vRamp[row]
		}
		for col := 0; col < width; col++ {
			v := grid[row][col]
			switch {
			case v >= 2:
				builder.WriteString(util.SetFgStyle(fullChar, rowRamp[col*2]))
			case v == 1:
				builder.WriteString(util.SetFgStyle(halfChar, rowRamp[col*2]))
			default:
				builder.WriteString(emptyChar)
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// drawBlockChannel draws one channel onto the block grid.
// priority: 2=L, 1=R. When both hit the same cell, value becomes 3.
func drawBlockChannel(grid [][]byte, levels [player.SpectrumBandCount]float64, width, height int, connect bool, priority byte) {
	var points [player.SpectrumBandCount][2]int
	for band := 0; band < player.SpectrumBandCount; band++ {
		if player.SpectrumBandCount > 1 {
			points[band][0] = band * (width - 1) / (player.SpectrumBandCount - 1)
		}
		points[band][1] = height - 1 - int(math.Round(levels[band]*float64(height-1)))
	}

	// Mark band positions.
	for band := 0; band < player.SpectrumBandCount; band++ {
		r, c := points[band][1], points[band][0]
		if r >= 0 && r < height && c >= 0 && c < width {
			if grid[r][c] == 0 {
				grid[r][c] = priority
			} else if grid[r][c] != priority {
				grid[r][c] = 3 // both
			}
		}
	}

	if connect {
		for band := 0; band < player.SpectrumBandCount-1; band++ {
			drawGridLineMaskAt(grid, points[band][0], points[band][1], points[band+1][0], points[band+1][1], priority)
		}
	}

	// Fill below the line: for each column, fill from topmost lit cell down to bottom.
	for col := 0; col < width; col++ {
		topRow := height // track the highest occupied row in this column
		for row := 0; row < height; row++ {
			if grid[row][col] != 0 {
				topRow = row
				break
			}
		}
		for row := topRow + 1; row < height; row++ {
			if grid[row][col] == 0 {
				grid[row][col] = priority
			} else if grid[row][col] != priority {
				grid[row][col] = 3
			}
		}
	}
}

// drawGridLineMaskAt draws a Bresenham line, with additive overlap detection.
func drawGridLineMaskAt(grid [][]byte, x1, y1, x2, y2 int, priority byte) {
	height := len(grid)
	width := len(grid[0])
	dx := absInt(x2 - x1)
	dy := -absInt(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx + dy
	for {
		if x1 >= 0 && x1 < width && y1 >= 0 && y1 < height {
			if grid[y1][x1] == 0 {
				grid[y1][x1] = priority
			} else if grid[y1][x1] != priority {
				grid[y1][x1] = 3
			}
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

// --- Shared utilities ---

// rampDim returns a complementary-hue copy of the progress ramp for the R channel.
// Uses a ~150° hue shift (near-complementary) with full saturation to create
// a vibrant, distinct color that contrasts clearly with the L channel.
// Cached: when width matches the previous call, the cached ramp is returned.
func (r *SpectrumRenderer) rampDim(width int) []color.Color {
	if r.progressDimLastWidth == float64(width) && len(r.progressDimRamp) > 0 {
		return r.progressDimRamp
	}
	base := r.ramp(width)
	shifted := make([]color.Color, len(base))
	for i, c := range base {
		hColor, _ := colorful.MakeColor(c)
		h, s, l := hColor.Hsl()
		// ~150° hue shift for strong complementary contrast, keep vibrant.
		h = math.Mod(h+150, 360)
		shifted[i] = colorful.Hsl(h, s, l)
	}
	r.progressDimRamp = shifted
	r.progressDimLastWidth = float64(width)
	return shifted
}

// --- Braille grid construction ---

// getBrailleGrid returns a zeroed [][]byte grid of the given dimensions,
// reusing a pool entry when width and height match the previous call.
func (r *SpectrumRenderer) getBrailleGrid(cache *[][]byte, width, height int) [][]byte {
	if *cache != nil && r.brailleCacheW == width && r.brailleCacheH == height {
		grid := *cache
		for _, row := range grid {
			for j := range row {
				row[j] = 0
			}
		}
		return grid
	}
	grid := make([][]byte, height)
	for i := range grid {
		grid[i] = make([]byte, width)
	}
	*cache = grid
	r.brailleCacheW = width
	r.brailleCacheH = height
	return grid
}

func (r *SpectrumRenderer) buildBrailleGrid(levels [player.SpectrumBandCount]float64, width, height int, connect bool, grid [][]byte) {
	subCols := width * 2
	subRows := height * 4

	var points [player.SpectrumBandCount][2]int
	for band := 0; band < player.SpectrumBandCount; band++ {
		if player.SpectrumBandCount > 1 {
			points[band][0] = band * (subCols - 1) / (player.SpectrumBandCount - 1)
		}
		points[band][1] = subRows - 1 - int(math.Round(levels[band]*float64(subRows-1)))
	}

	if connect {
		for band := 0; band < player.SpectrumBandCount-1; band++ {
			drawBrailleLine(grid,
				points[band][0], points[band][1],
				points[band+1][0], points[band+1][1],
			)
		}
	}

	for band := 0; band < player.SpectrumBandCount; band++ {
		setBrailleDot(grid, points[band][0], points[band][1])
	}
}

// --- Bar rendering helpers ---

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
	builder.WriteString(strings.Repeat(string(emptyBlock), emptyChars))
	return builder.String()
}

func renderMirrorBarLine(level float64, width int, ramp []color.Color, halfBlock, fullBlock, emptyBlock rune) string {
	if width <= 1 {
		return renderSpectrumBar(level, width, ramp, halfBlock, fullBlock, emptyBlock)
	}

	halfWidth := width / 2
	rightWidth := width - halfWidth
	fillUnitsL := spectrumFillUnits(level, halfWidth)
	fullCharsL := fillUnitsL / 2
	hasHalfL := fillUnitsL%2 == 1
	fillUnitsR := spectrumFillUnits(level, rightWidth)
	fullCharsR := fillUnitsR / 2
	hasHalfR := fillUnitsR%2 == 1

	var builder strings.Builder
	builder.Grow(width)

	// Left half.
	for col := 0; col < halfWidth; col++ {
		distFromCenter := halfWidth - 1 - col
		switch {
		case distFromCenter < fullCharsL:
			builder.WriteString(util.SetFgStyle(string(fullBlock), ramp[col*2]))
		case hasHalfL && distFromCenter == fullCharsL:
			builder.WriteString(spectrumHalfBlockStyle(ramp[col*2], ramp[col*2+1]).Render(string(halfBlock)))
		default:
			builder.WriteString(string(emptyBlock))
		}
	}

	// Right half.
	for col := 0; col < rightWidth; col++ {
		distFromCenter := col
		rightCol := halfWidth + col
		switch {
		case distFromCenter < fullCharsR:
			builder.WriteString(util.SetFgStyle(string(fullBlock), ramp[rightCol*2]))
		case hasHalfR && distFromCenter == fullCharsR:
			builder.WriteString(spectrumHalfBlockStyle(ramp[rightCol*2], ramp[rightCol*2+1]).Render(string(halfBlock)))
		default:
			builder.WriteString(string(emptyBlock))
		}
	}
	return builder.String()
}

// renderMirrorBarLineDual builds a mirror bar with left half driven by levelL
// and right half driven by levelR.
func renderMirrorBarLineDual(levelL, levelR float64, width int, ramp []color.Color, halfBlock, fullBlock, emptyBlock rune) string {
	if width <= 1 {
		return renderSpectrumBar(levelL, width, ramp, halfBlock, fullBlock, emptyBlock)
	}

	halfWidth := width / 2
	rightWidth := width - halfWidth
	fillUnitsL := spectrumFillUnits(levelL, halfWidth)
	fullCharsL := fillUnitsL / 2
	hasHalfL := fillUnitsL%2 == 1
	fillUnitsR := spectrumFillUnits(levelR, rightWidth)
	fullCharsR := fillUnitsR / 2
	hasHalfR := fillUnitsR%2 == 1

	var builder strings.Builder
	builder.Grow(width)

	// Left half.
	for col := 0; col < halfWidth; col++ {
		distFromCenter := halfWidth - 1 - col
		switch {
		case distFromCenter < fullCharsL:
			builder.WriteString(util.SetFgStyle(string(fullBlock), ramp[col*2]))
		case hasHalfL && distFromCenter == fullCharsL:
			builder.WriteString(spectrumHalfBlockStyle(ramp[col*2], ramp[col*2+1]).Render(string(halfBlock)))
		default:
			builder.WriteString(string(emptyBlock))
		}
	}

	// Right half.
	for col := 0; col < rightWidth; col++ {
		distFromCenter := col
		rightCol := halfWidth + col
		switch {
		case distFromCenter < fullCharsR:
			builder.WriteString(util.SetFgStyle(string(fullBlock), ramp[rightCol*2]))
		case hasHalfR && distFromCenter == fullCharsR:
			builder.WriteString(spectrumHalfBlockStyle(ramp[rightCol*2], ramp[rightCol*2+1]).Render(string(halfBlock)))
		default:
			builder.WriteString(string(emptyBlock))
		}
	}
	return builder.String()
}

// renderMirrorIdleLineDual renders a mirror bar line showing only idle heads at the center.
func renderMirrorIdleLineDual(width int, ramp []color.Color, halfBlock, emptyBlock rune, leftIdle, rightIdle bool) string {
	halfWidth := width / 2
	rightWidth := width - halfWidth

	var builder strings.Builder
	builder.Grow(width)

	// Left half: empty except rightmost halfBlock if leftIdle
	for col := 0; col < halfWidth; col++ {
		if leftIdle && col == halfWidth-1 {
			builder.WriteString(util.SetFgStyle(string(halfBlock), ramp[(halfWidth-1)*2]))
		} else {
			builder.WriteString(string(emptyBlock))
		}
	}

	// Right half: empty except leftmost halfBlock if rightIdle
	for col := 0; col < rightWidth; col++ {
		if rightIdle && col == 0 {
			builder.WriteString(util.SetFgStyle(string(halfBlock), ramp[halfWidth*2]))
		} else {
			builder.WriteString(string(emptyBlock))
		}
	}
	return builder.String()
}

// renderBarVertical renders bars growing upward from bottom, one column per frequency group.
// When reversed is true (right orientation), columns are placed right-to-left.
func (r *SpectrumRenderer) renderBarVertical(frame player.SpectrumFrame, width, height int, reversed bool) string {
	cfg := configs.AppConfig.Main.Visualizer
	halfBlock, fullBlock, emptyBlock := cfg.BarCharacters()
	vertEnabled := cfg.BarVerticalGradient
	horizEnabled := cfg.BarHorizontalGradient

	// Per-column ramp: left-to-right color gradient (width*2 entries)
	colRamp := r.ramp(width)

	// Per-row vertical gradient ramps
	// For vertical bars, rowRamps maps rows (vertical axis) to colors.
	// We build rowRamps with width as horizontal dimension and height as rows.
	var rowRamps [][]color.Color
	if vertEnabled {
		rowRamps = r.rowRamps(width, height, true)
	}

	// Compute levels: one per column (width frequency groups)
	levels := make([]float64, width)
	for col := 0; col < width; col++ {
		levels[col] = spectrumRowLevel(frame, col, width)
	}

	// Build grid of styled strings
	grid := make([][]string, height)
	for row := range grid {
		grid[row] = make([]string, width)
		for col := range grid[row] {
			grid[row][col] = string(emptyBlock)
		}
	}

	for col := 0; col < width; col++ {
		level := levels[col]
		fillUnits := spectrumFillUnits(level, height)
		fullChars := fillUnits / 2
		hasHalf := fillUnits%2 == 1

		// Idle bar head: show halfBlock at top of column
		if cfg.ShowIdleBarHeads && level > 0 && level <= 0.01 {
			row := 0
			c := colRamp[col*2]
			if horizEnabled {
				c = r.horizontalColor(row, height)
			}
			grid[row][col] = util.SetFgStyle(string(halfBlock), c)
			continue
		}

		for i := 0; i < fullChars; i++ {
			row := height - 1 - i
			if row < 0 {
				break
			}
			c := colRamp[col*2]
			if vertEnabled && len(rowRamps) > 0 {
				c = rowRamps[row][col*2]
			}
			if horizEnabled {
				horizColor := r.horizontalColor(row, height)
				if vertEnabled && len(rowRamps) > 0 {
					hColor, _ := colorful.MakeColor(rowRamps[row][col*2])
					horizColorful, _ := colorful.MakeColor(horizColor)
					c = hColor.BlendLuv(horizColorful, 0.5)
				} else {
					c = horizColor
				}
			}
			grid[row][col] = util.SetFgStyle(string(fullBlock), c)
		}
		if hasHalf && fullChars < height {
			row := height - 1 - fullChars
			if row >= 0 {
				c := colRamp[col*2]
				if vertEnabled && len(rowRamps) > 0 {
					c = rowRamps[row][col*2]
				}
				if horizEnabled {
					horizColor := r.horizontalColor(row, height)
					if vertEnabled && len(rowRamps) > 0 {
						hColor, _ := colorful.MakeColor(rowRamps[row][col*2])
						horizColorful, _ := colorful.MakeColor(horizColor)
						c = hColor.BlendLuv(horizColorful, 0.5)
					} else {
						c = horizColor
					}
				}
				grid[row][col] = util.SetFgStyle(string(halfBlock), c)
			}
		}
	}

	// Render grid to string
	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		if reversed {
			for i := width - 1; i >= 0; i-- {
				builder.WriteString(grid[row][i])
			}
		} else {
			for _, s := range grid[row] {
				builder.WriteString(s)
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// renderBarHorizontal renders bars mirrored vertically: top half grows upward, bottom half downward.
func (r *SpectrumRenderer) renderBarHorizontal(frame player.SpectrumFrame, width, height int) string {
	topHalf := height / 2
	bottomHalf := height - topHalf

	var builder strings.Builder
	builder.Grow((width + 1) * height)

	// Top half: bars grow upward from center. Render normally then reverse rows.
	// Use bottomHalf groups for the top half (same frequency mapping, reversed rendering).
	if topHalf > 0 {
		// Render using the same frequency groups as bottom half for symmetry
		topContent := r.renderBarBottom(frame, width, topHalf)
		// Reverse the rows so the baseline is at the bottom of this section
		lines := strings.Split(strings.TrimSuffix(topContent, "\n"), "\n")
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
		builder.WriteString(strings.Join(lines, "\n"))
		builder.WriteByte('\n')
	}

	// Bottom half: bars grow downward from center (normal behavior)
	if bottomHalf > 0 {
		builder.WriteString(r.renderBarBottom(frame, width, bottomHalf))
	}

	return builder.String()
}

// renderBarVerticalCenter renders bars mirrored horizontally: left half grows leftward, right half rightward.
func (r *SpectrumRenderer) renderBarVerticalCenter(frame player.SpectrumFrame, width, height int) string {
	halfWidth := width / 2
	rightWidth := width - halfWidth

	// Left half: bars grow leftward (origin at right edge of left half)
	// Use renderBarVertical with reversed=true on a sub-width
	leftContent := r.renderBarVertical(frame, halfWidth, height, true)
	// Right half: bars grow rightward (origin at left edge of right half, i.e. normal)
	rightContent := r.renderBarVertical(frame, rightWidth, height, false)

	// Merge left and right halves line by line
	leftLines := strings.Split(strings.TrimSuffix(leftContent, "\n"), "\n")
	rightLines := strings.Split(strings.TrimSuffix(rightContent, "\n"), "\n")

	var builder strings.Builder
	builder.Grow((width + 1) * height)
	for row := 0; row < height; row++ {
		if row < len(leftLines) {
			builder.WriteString(leftLines[row])
		}
		if row < len(rightLines) {
			builder.WriteString(rightLines[row])
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// --- Shared utilities ---

func hasSignal(frame player.SpectrumFrame) bool {
	for _, level := range frame.LevelsL {
		if level != 0 {
			return true
		}
	}
	for _, level := range frame.LevelsR {
		if level != 0 {
			return true
		}
	}
	for _, level := range frame.Levels {
		if level != 0 {
			return true
		}
	}
	return false
}

func (r *SpectrumRenderer) ramp(width int) []color.Color {
	if r.progressLastWidth != float64(width) || len(r.progressRamp) == 0 {
		start, end := model.GetProgressColor()
		r.progressRamp = util.MakeRamp(start, end, float64(width*2))
		r.progressLastWidth = float64(width)
	}
	return r.progressRamp
}

func (r *SpectrumRenderer) rowRamps(width, height int, enabled bool) [][]color.Color {
	if !enabled {
		return nil
	}
	if r.vertLastWidth == float64(width) && r.vertLastHeight == height && len(r.vertRowRamps) > 0 {
		return r.vertRowRamps
	}
	if height <= 1 {
		return nil
	}

	baseRamp := r.ramp(width)

	vertStart, vertEnd := util.GetRandomRgbColor(true)
	cStart, _ := colorful.Hex(vertStart)
	cEnd, _ := colorful.Hex(vertEnd)

	r.vertRowRamps = make([][]color.Color, height)
	for row := 0; row < height; row++ {
		t := float64(row) / float64(height-1)
		vColor := cStart.BlendLuv(cEnd, t)

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

// horizontalColor returns a single color interpolated between theme start/end
// based on the row position within the total height.
func (r *SpectrumRenderer) horizontalColor(row, height int) color.Color {
	start, end := model.GetProgressColor()
	cStart, _ := colorful.Hex(start)
	cEnd, _ := colorful.Hex(end)
	t := float64(row) / float64(max(height-1, 1))
	return cStart.BlendLuv(cEnd, t)
}

// horizontalBarRamp returns a uniform-color ramp (all entries the same color)
// for a given row, derived from the horizontal gradient position.
func (r *SpectrumRenderer) horizontalBarRamp(row, width, height int) []color.Color {
	c := r.horizontalColor(row, height)
	rampLen := width * 2
	result := make([]color.Color, rampLen)
	for i := range result {
		result[i] = c
	}
	return result
}

// blendRamps blends a horizontal color into each entry of a row ramp at 50%.
func blendRamps(rowRamp []color.Color, horizColor color.Color) []color.Color {
	result := make([]color.Color, len(rowRamp))
	horizColorful, _ := colorful.MakeColor(horizColor)
	for i, c := range rowRamp {
		hColor, _ := colorful.MakeColor(c)
		result[i] = hColor.BlendLuv(horizColorful, 0.5)
	}
	return result
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

func spectrumRowLevelFrom(levels [player.SpectrumBandCount]float64, row, height int) float64 {
	group := height - row - 1
	start := group * player.SpectrumBandCount / height
	end := (group + 1) * player.SpectrumBandCount / height
	level := 0.0
	for band := start; band < end; band++ {
		level = max(level, levels[band])
	}
	return clamp(level, 0, 1)
}

func spectrumFillUnits(level float64, width int) int {
	return int(math.Round(clamp(level, 0, 1) * float64(width*2)))
}

// --- Phase difference visualization ---

// computePhaseMask returns a per-column phase correlation array [0,1] where
// 1 = channels perfectly in phase, 0 = completely out of phase.
func computePhaseMask(phasesL, phasesR [player.SpectrumBandCount]float64, width int) []float64 {
	mask := make([]float64, width)
	if width <= 0 {
		return mask
	}
	for col := 0; col < width; col++ {
		startBand := col * player.SpectrumBandCount / width
		endBand := (col + 1) * player.SpectrumBandCount / width
		if endBand > player.SpectrumBandCount {
			endBand = player.SpectrumBandCount
		}
		if startBand >= endBand {
			startBand = endBand - 1
			if startBand < 0 {
				startBand = 0
			}
		}
		sum := 0.0
		for band := startBand; band < endBand; band++ {
			// Compute absolute phase difference in [0, 2π), then map to [0, π].
			diff := math.Abs(math.Mod(phasesL[band]-phasesR[band]+math.Pi, 2*math.Pi) - math.Pi)
			sum += 1.0 - diff/math.Pi
		}
		mask[col] = clamp(sum/float64(endBand-startBand), 0, 1)
	}
	return mask
}

// blendPhaseColor blends the given color toward an orange warning hue
// based on phase correlation (1.0 = in phase = no change; 0.0 = out of phase = full shift).
func blendPhaseColor(c color.Color, correlation float64) color.Color {
	if correlation >= 1.0 {
		return c
	}
	hColor, _ := colorful.MakeColor(c)
	h, s, l := hColor.Hsl()
	// Shift hue toward orange/warning (30°) when correlation is low.
	blend := 1.0 - correlation
	targetH := 30.0
	// Wrap hue difference.
	hDiff := math.Mod(targetH-h+540, 360) - 180
	h = math.Mod(h+hDiff*blend+360, 360)
	// Boost saturation and reduce lightness for out-of-phase warning.
	s = clamp(s+0.15*blend, 0, 1)
	l = clamp(l-0.1*blend, 0, 1)
	return colorful.Hsl(h, s, l)
}
