package ui

import (
	"math"
)

var (
	ProgressActiveColor     = LyricActiveColor
	ProgressTransitionColor = LyricTransitionColor
	ProgressWhiteColor      = LyricWhiteColor
)

type progressRenderMode string

const (
	progressRenderModeSmooth progressRenderMode = "smooth"
	progressRenderModeWave   progressRenderMode = "wave"
	progressRenderModeGlow   progressRenderMode = "glow"
)

func progressRampForMode(width int, fullSize int, animationTime float64, mode progressRenderMode) []string {
	if width <= 0 {
		return nil
	}

	switch mode {
	case progressRenderModeWave:
		return progressRampWave(width, animationTime)
	case progressRenderModeGlow:
		return progressRampGlow(width, fullSize, animationTime)
	case progressRenderModeSmooth:
		fallthrough
	default:
		return nil
	}
}

func progressRampWave(width int, animationTime float64) []string {
	ramp := make([]string, width)
	for i := 0; i < width; i++ {
		w := math.Sin(animationTime*2.0 - float64(i)*0.3)
		w = (w + 1) / 2
		c := blendColor(ProgressTransitionColor, ProgressActiveColor, w)
		ramp[i] = string(c)
	}
	return ramp
}

func progressRampGlow(width int, fullSize int, animationTime float64) []string {
	ramp := make([]string, width)
	for i := range ramp {
		ramp[i] = string(ProgressActiveColor)
	}
	if fullSize <= 0 {
		return ramp
	}
	if fullSize > width {
		fullSize = width
	}

	pulse := (math.Sin(animationTime*2.0) + 1) / 2

	idx := fullSize - 1
	strength := 0.25 + pulse*0.35
	strength = clamp(strength, 0, 0.8)
	ramp[idx] = string(blendColor(ProgressActiveColor, ProgressWhiteColor, strength))

	if idx-1 >= 0 {
		ramp[idx-1] = string(blendColor(ProgressActiveColor, ProgressWhiteColor, 0.08+pulse*0.12))
	}
	return ramp
}
