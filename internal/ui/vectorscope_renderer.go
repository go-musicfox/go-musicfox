package ui

import (
	"math"
	"strings"

	"github.com/anhoder/foxful-cli/util"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

// --- Vectorscope character helpers ---

func (r *SpectrumRenderer) vectorscopeFullChar() string {
	return firstCharOf(configs.AppConfig.Main.Visualizer.VectorscopeFullBlock)
}

func (r *SpectrumRenderer) vectorscopeHalfChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.VectorscopeHalfBlock)
	if ch == "" {
		ch = r.vectorscopeFullChar()
	}
	return ch
}

func (r *SpectrumRenderer) vectorscopeEmptyChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.VectorscopeEmptyBlock)
	if ch == "" {
		ch = " "
	}
	return ch
}

// --- Vectorscope braille mode ---

func (r *SpectrumRenderer) renderVectorscopeBraille(frame player.RawSampleFrame, width, height int) string {
	if frame.Count < 2 || width < 1 || height < 1 {
		return strings.Repeat(strings.Repeat(" ", width)+"\n", height)
	}

	gridL := r.getBrailleGrid(&r.brailleGridLCache, width, height)
	gridR := r.getBrailleGrid(&r.brailleGridRCache, width, height)

	subCols := width * 2
	subRows := height * 4

	hasStereo := hasStereoSamples(frame)
	mono := configs.AppConfig.Main.Visualizer.IsMono() || !hasStereo

	half := frame.Count / 2
	if half < 1 {
		half = 1
	}

	for i := 0; i < frame.Count; i++ {
		sampleL := frame.SamplesL[i]
		sampleR := sampleL
		if !mono {
			sampleR = frame.SamplesR[i]
		}

		// Clamp to [-1, 1] range.
		sampleL = clamp(sampleL, -1, 1)
		sampleR = clamp(sampleR, -1, 1)

		// L amplitude → X (left: -1 → col 0, right: +1 → col subCols-1).
		subCol := int(math.Round((sampleL + 1) * 0.5 * float64(subCols-1)))
		// R amplitude → Y (inverted: +1 → bottom, -1 → top).
		subRow := int(math.Round((1 - sampleR) * 0.5 * float64(subRows-1)))
		subCol = clampInt(subCol, 0, subCols-1)
		subRow = clampInt(subRow, 0, subRows-1)

		if i < half {
			setBrailleDot(gridL, subCol, subRow)
		} else {
			setBrailleDot(gridR, subCol, subRow)
		}
	}

	return r.renderBrailleGridDual(gridL, gridR, width, height)
}

// --- Vectorscope block mode ---

func (r *SpectrumRenderer) renderVectorscopeBlock(frame player.RawSampleFrame, width, height int) string {
	if frame.Count < 2 || width < 1 || height < 1 {
		return strings.Repeat(strings.Repeat(" ", width)+"\n", height)
	}

	fullChar := r.vectorscopeFullChar()
	halfChar := r.vectorscopeHalfChar()
	emptyChar := r.vectorscopeEmptyChar()
	if halfChar == "" {
		halfChar = fullChar
	}

	ramp := r.ramp(width)
	rampDim := r.rampDim(width)
	vRamp := r.vertBlockRamps(width, height)
	vRampDim := r.vertBlockRampsDim(width, height, rampDim)

	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)

	w, h := width, height
	hasStereo := hasStereoSamples(frame)
	mono := configs.AppConfig.Main.Visualizer.IsMono() || !hasStereo

	half := frame.Count / 2
	if half < 1 {
		half = 1
	}

	for i := 0; i < frame.Count; i++ {
		sampleL := frame.SamplesL[i]
		sampleR := sampleL
		if !mono {
			sampleR = frame.SamplesR[i]
		}
		sampleL = clamp(sampleL, -1, 1)
		sampleR = clamp(sampleR, -1, 1)

		col := int(math.Round((sampleL + 1) * 0.5 * float64(w-1)))
		row := int(math.Round((1 - sampleR) * 0.5 * float64(h-1)))
		col = clampInt(col, 0, w-1)
		row = clampInt(row, 0, h-1)

		priority := byte(2) // primary (first half)
		if i >= half {
			priority = byte(1) // dim (second half)
		}

		if grid[row][col] == 0 {
			grid[row][col] = priority
		} else if grid[row][col] != priority {
			grid[row][col] = 3
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
			v := grid[row][col]
			switch {
			case v == 3:
				builder.WriteString(util.SetFgStyle(fullChar, rowRampDim[col*2]))
			case v == 2:
				builder.WriteString(util.SetFgStyle(fullChar, rowRamp[col*2]))
			case v == 1:
				builder.WriteString(util.SetFgStyle(halfChar, rowRampDim[col*2]))
			default:
				builder.WriteString(emptyChar)
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}
