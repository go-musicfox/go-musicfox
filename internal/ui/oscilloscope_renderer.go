package ui

import (
	"math"
	"strings"

	"github.com/anhoder/foxful-cli/util"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

// --- Oscilloscope character helpers ---

func (r *SpectrumRenderer) oscilloscopeFullChar() string {
	return firstCharOf(configs.AppConfig.Main.Visualizer.OscilloscopeFullBlock)
}

func (r *SpectrumRenderer) oscilloscopeHalfChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.OscilloscopeHalfBlock)
	if ch == "" {
		ch = r.oscilloscopeFullChar()
	}
	return ch
}

func (r *SpectrumRenderer) oscilloscopeEmptyChar() string {
	ch := firstCharOf(configs.AppConfig.Main.Visualizer.OscilloscopeEmptyBlock)
	if ch == "" {
		ch = " "
	}
	return ch
}

// --- Shared helper ---

func hasStereoSamples(frame player.RawSampleFrame) bool {
	for i := 0; i < frame.Count; i++ {
		if math.Abs(frame.SamplesR[i]) > 1e-9 {
			return true
		}
	}
	return false
}

// --- Oscilloscope braille mode ---

func (r *SpectrumRenderer) renderOscilloscopeBraille(frame player.RawSampleFrame, width, height int) string {
	if frame.Count < 2 || width < 1 || height < 1 {
		return strings.Repeat(strings.Repeat(" ", width)+"\n", height)
	}
	if configs.AppConfig.Main.Visualizer.IsMono() || !hasStereoSamples(frame) {
		return r.renderOscilloscopeBrailleMono(frame, width, height)
	}
	return r.renderOscilloscopeBrailleDual(frame, width, height)
}

func (r *SpectrumRenderer) renderOscilloscopeBrailleDual(frame player.RawSampleFrame, width, height int) string {
	gridL := r.getBrailleGrid(&r.brailleGridLCache, width, height)
	gridR := r.getBrailleGrid(&r.brailleGridRCache, width, height)
	connect := !configs.AppConfig.Main.Visualizer.OscilloscopeScatter
	r.buildOscilloscopeBraille(frame.SamplesL, frame.Count, width, height, connect, gridL)
	r.buildOscilloscopeBraille(frame.SamplesR, frame.Count, width, height, connect, gridR)
	return r.renderBrailleGridDual(gridL, gridR, width, height)
}

func (r *SpectrumRenderer) renderOscilloscopeBrailleMono(frame player.RawSampleFrame, width, height int) string {
	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)
	connect := !configs.AppConfig.Main.Visualizer.OscilloscopeScatter
	r.buildOscilloscopeBraille(frame.SamplesL, frame.Count, width, height, connect, grid)
	return r.renderBrailleGridMono(grid, width, height)
}

func (r *SpectrumRenderer) buildOscilloscopeBraille(samples []float64, count, w, h int, connect bool, grid [][]byte) {
	if count < 2 || w < 1 || h < 1 {
		return
	}
	subCols := w * 2
	subRows := h * 4

	// Decimate dense samples using peak-detect per sub-column.
	if count > subCols {
		samplesPerSubCol := float64(count) / float64(subCols)
		var prevMidSubCol, prevMidSubRow int
		hasPrev := false
		for subCol := 0; subCol < subCols; subCol++ {
			start := int(float64(subCol) * samplesPerSubCol)
			end := int(float64(subCol+1) * samplesPerSubCol)
			if end > count {
				end = count
			}
			if start >= end {
				continue
			}
			minV, maxV := samples[start], samples[start]
			for i := start; i < end; i++ {
				v := samples[i]
				if v < minV {
					minV = v
				}
				if v > maxV {
					maxV = v
				}
			}
			minSR := subRows - 1 - int(math.Round((minV+1)*0.5*float64(subRows-1)))
			maxSR := subRows - 1 - int(math.Round((maxV+1)*0.5*float64(subRows-1)))
			minSR = clampInt(minSR, 0, subRows-1)
			maxSR = clampInt(maxSR, 0, subRows-1)

			for sr := maxSR; sr <= minSR; sr++ {
				setBrailleDot(grid, subCol, sr)
			}

			if connect {
				midSR := (maxSR + minSR) / 2
				if hasPrev {
					drawBrailleLine(grid, prevMidSubCol, prevMidSubRow, subCol, midSR)
				}
				prevMidSubCol = subCol
				prevMidSubRow = midSR
				hasPrev = true
			}
		}
		return
	}

	// Direct point mapping for sparse samples.
	var prevSubCol, prevSubRow int
	first := true
	for i := 0; i < count; i++ {
		subCol := i * (subCols - 1) / (count - 1)
		norm := clamp((samples[i]+1)*0.5, 0, 1)
		subRow := subRows - 1 - int(math.Round(norm*float64(subRows-1)))

		if connect && !first {
			drawBrailleLine(grid, prevSubCol, prevSubRow, subCol, subRow)
		}
		setBrailleDot(grid, subCol, subRow)
		prevSubCol, prevSubRow = subCol, subRow
		first = false
	}
}

// --- Oscilloscope block mode ---

func (r *SpectrumRenderer) renderOscilloscopeBlock(frame player.RawSampleFrame, width, height int) string {
	if frame.Count < 2 || width < 1 || height < 1 {
		return strings.Repeat(strings.Repeat(" ", width)+"\n", height)
	}
	fullChar := r.oscilloscopeFullChar()
	halfChar := r.oscilloscopeHalfChar()
	emptyChar := r.oscilloscopeEmptyChar()
	if configs.AppConfig.Main.Visualizer.IsMono() || !hasStereoSamples(frame) {
		return r.renderOscilloscopeBlockMono(fullChar, halfChar, emptyChar, frame, width, height)
	}
	return r.renderOscilloscopeBlockDual(fullChar, halfChar, emptyChar, frame, width, height)
}

func (r *SpectrumRenderer) renderOscilloscopeBlockDual(fullChar, halfChar, emptyChar string, frame player.RawSampleFrame, width, height int) string {
	if halfChar == "" {
		halfChar = fullChar
	}
	ramp := r.ramp(width)
	rampDim := r.rampDim(width)
	vRamp := r.vertBlockRamps(width, height)
	vRampDim := r.vertBlockRampsDim(width, height, rampDim)

	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)
	connect := !configs.AppConfig.Main.Visualizer.OscilloscopeScatter
	r.drawOscilloscopeBlockChannel(grid, frame.SamplesL, frame.Count, width, height, connect, 2)
	r.drawOscilloscopeBlockChannel(grid, frame.SamplesR, frame.Count, width, height, connect, 1)

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

func (r *SpectrumRenderer) renderOscilloscopeBlockMono(fullChar, halfChar, emptyChar string, frame player.RawSampleFrame, width, height int) string {
	if halfChar == "" {
		halfChar = fullChar
	}
	ramp := r.ramp(width)
	vRamp := r.vertBlockRamps(width, height)

	grid := r.getBrailleGrid(&r.brailleGridMCache, width, height)
	connect := !configs.AppConfig.Main.Visualizer.OscilloscopeScatter
	r.drawOscilloscopeBlockChannel(grid, frame.SamplesL, frame.Count, width, height, connect, 2)

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

func (r *SpectrumRenderer) drawOscilloscopeBlockChannel(grid [][]byte, samples []float64, count, width, height int, connect bool, priority byte) {
	if count == 0 || width < 1 || height < 1 {
		return
	}
	h := height

	var prevMidCol, prevMidRow int
	hasPrev := false

	for col := 0; col < width; col++ {
		start := col * count / width
		end := (col + 1) * count / width
		if end > count {
			end = count
		}
		if start >= end {
			continue
		}

		minV, maxV := samples[start], samples[start]
		for i := start; i < end; i++ {
			v := samples[i]
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}

		minRow := h - 1 - int(math.Round((minV+1)*0.5*float64(h-1)))
		maxRow := h - 1 - int(math.Round((maxV+1)*0.5*float64(h-1)))
		minRow = clampInt(minRow, 0, h-1)
		maxRow = clampInt(maxRow, 0, h-1)

		for row := maxRow; row <= minRow; row++ {
			if grid[row][col] == 0 {
				grid[row][col] = priority
			} else if grid[row][col] != priority {
				grid[row][col] = 3
			}
		}

		if connect {
			midRow := (maxRow + minRow) / 2
			if hasPrev {
				drawGridLineMaskAt(grid, prevMidCol, prevMidRow, col, midRow, priority)
			}
			prevMidCol = col
			prevMidRow = midRow
			hasPrev = true
		}
	}
}
