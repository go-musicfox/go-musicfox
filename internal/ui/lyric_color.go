package ui

import (
	"image/color"
	"math"
	"strings"

	"github.com/anhoder/foxful-cli/util"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// Lyric color accessors — fetched from current theme on each call.
// Falls back to hardcoded defaults when theme system is not initialized.
func lyricActiveColor() color.Color {
	return configs.SafeGetForeground(
		configs.GetCurrentAppColors().LyricActive,
		configs.LyricActiveColor)
}
func lyricTransitionColor() color.Color {
	return configs.SafeGetForeground(
		configs.GetCurrentAppColors().LyricTransition,
		configs.LyricTransitionColor)
}
func lyricInactiveColor() color.Color {
	return configs.SafeGetForeground(
		configs.GetCurrentAppColors().LyricInactive,
		configs.LyricInactiveColor)
}
func lyricWhiteColor() color.Color {
	return configs.SafeGetForeground(
		configs.GetCurrentAppColors().LyricHighlight,
		configs.LyricWhiteColor)
}

// Pre-computed RGB accessors — converted from theme colors on each call.
func lyricActiveRGB() color.RGBA {
	return toRGBA(lyricActiveColor())
}
func lyricTransitionRGB() color.RGBA {
	return toRGBA(lyricTransitionColor())
}
func lyricInactiveRGB() color.RGBA {
	return toRGBA(lyricInactiveColor())
}
func lyricWhiteRGB() color.RGBA {
	return toRGBA(lyricWhiteColor())
}

// blendColorFast mixes two pre-computed RGB colors with ratio t (0.0 - 1.0).
func blendColorFast(c1, c2 color.RGBA, t float64) color.Color {
	t1 := 1 - t
	return color.RGBA{
		R: uint8(float64(c1.R)*t1 + float64(c2.R)*t),
		G: uint8(float64(c1.G)*t1 + float64(c2.G)*t),
		B: uint8(float64(c1.B)*t1 + float64(c2.B)*t),
		A: 255,
	}
}

// blendColor mixes two color.Color with a given ratio t (0.0 - 1.0).
// Legacy function for backward compatibility.
func blendColor(c1, c2 color.Color, t float64) color.Color {
	rgb1 := toRGBA(c1)
	rgb2 := toRGBA(c2)
	return blendColorFast(rgb1, rgb2, t)
}

func toRGBA(c color.Color) color.RGBA {
	if rgba, ok := c.(color.RGBA); ok {
		return rgba
	}
	r, g, b, _ := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: 255,
	}
}

// clamp restricts a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// renderSmooth renders lyrics with smooth mode - uses color gradient transition with interpolation.
// All words participate in gradient calculation for continuous color transitions.
func renderSmooth(words []wordWithTiming, progress float64) string {
	var sb strings.Builder
	totalWords := len(words)
	if totalWords == 0 {
		return ""
	}

	fadeWidth := 2.0

	currentWordIdx := -1
	var currentWordInterpolation float64
	for i, w := range words {
		if w.state == wordStatePlaying {
			currentWordIdx = i
			currentWordInterpolation = w.interpolation
			break
		}
	}

	var preciseCurrentPos float64
	if currentWordIdx >= 0 {
		preciseCurrentPos = float64(currentWordIdx) + currentWordInterpolation
	} else {
		preciseCurrentPos = float64(totalWords) * progress
	}

	inactiveRGB := lyricInactiveRGB()
	activeRGB := lyricActiveRGB()

	for i, w := range words {
		pos := float64(i)

		var activation float64
		if pos < preciseCurrentPos-fadeWidth {
			activation = 1.0
		} else if pos > preciseCurrentPos {
			activation = 0.0
		} else {
			activation = (preciseCurrentPos - pos) / fadeWidth
			activation = clamp(activation, 0, 1)
		}

		if w.state == wordStatePlaying && w.interpolation > activation {
			activation = w.interpolation
		}

		if w.state == wordStatePlayed {
			activation = 1.0
		}

		color := blendColorFast(inactiveRGB, activeRGB, activation)
		sb.WriteString(util.SetFgStyle(w.text, color))
	}

	return sb.String()
}

// renderWave renders lyrics with wave mode - adds dynamic wave effect with interpolation.
// animationTime: current time in seconds for smooth animation
func renderWave(words []wordWithTiming, progress float64, animationTime float64) string {
	var sb strings.Builder
	totalWords := len(words)
	if totalWords == 0 {
		return ""
	}

	var currentWordIdx int = -1
	var currentWordInterpolation float64
	for i, w := range words {
		if w.state == wordStatePlaying {
			currentWordIdx = i
			currentWordInterpolation = w.interpolation
			break
		}
	}

	var preciseCurrentPos float64
	if currentWordIdx >= 0 {
		preciseCurrentPos = float64(currentWordIdx) + currentWordInterpolation
	} else {
		preciseCurrentPos = float64(totalWords) * progress
	}

	inactiveRGB := lyricInactiveRGB()
	activeRGB := lyricActiveRGB()
	transitionRGB := lyricTransitionRGB()

	for i, w := range words {
		pos := float64(i)

		var activation float64
		if pos < preciseCurrentPos-1 {
			activation = 1.0
		} else if pos > preciseCurrentPos+1 {
			activation = 0.0
		} else {
			activation = (preciseCurrentPos - pos + 1) / 2.0
			activation = clamp(activation, 0, 1)
		}

		if w.state == wordStatePlaying && w.interpolation > activation {
			activation = w.interpolation
		}

		if w.state == wordStatePlayed {
			activation = 1.0
		}

		var lcolor color.Color
		if activation >= 1.0 {
			wave := math.Sin(animationTime*3.0 - float64(i)*0.5)
			wave = (wave + 1) / 2
			lcolor = blendColorFast(transitionRGB, activeRGB, wave)
		} else if activation > 0 {
			lcolor = blendColorFast(inactiveRGB, activeRGB, activation)
		} else {
			lcolor = lyricInactiveColor()
		}

		sb.WriteString(util.SetFgStyle(w.text, lcolor))
	}

	return sb.String()
}

// renderGlow renders lyrics with glow mode - current word glows with interpolation.
// Uses smooth transitions to avoid color jumps when words change.
// animationTime: current time in seconds for smooth animation
func renderGlow(words []wordWithTiming, currentWordIndex int, animationTime float64) string {
	var sb strings.Builder

	var currentInterpolation float64
	if currentWordIndex >= 0 && currentWordIndex < len(words) {
		currentInterpolation = words[currentWordIndex].interpolation
	}

	pulse := (math.Sin(animationTime*2.0) + 1) / 2
	const preheatMax = 0.4

	inactiveRGB := lyricInactiveRGB()
	activeRGB := lyricActiveRGB()
	transitionRGB := lyricTransitionRGB()
	whiteRGB := lyricWhiteRGB()

	for i, w := range words {
		var lcolor color.Color

		if i < currentWordIndex-1 {
			lcolor = lyricActiveColor()
		} else if i == currentWordIndex-1 {
			fadeOut := currentInterpolation
			glowStrength := (1 - fadeOut) * (0.3 + pulse*0.15)
			lcolor = blendColorFast(activeRGB, whiteRGB, glowStrength)
		} else if i == currentWordIndex {
			preheatEndRGB := color.RGBA{
				R: uint8(float64(inactiveRGB.R)*(1-preheatMax) + float64(transitionRGB.R)*preheatMax),
				G: uint8(float64(inactiveRGB.G)*(1-preheatMax) + float64(transitionRGB.G)*preheatMax),
				B: uint8(float64(inactiveRGB.B)*(1-preheatMax) + float64(transitionRGB.B)*preheatMax),
				A: 255,
			}

			baseRGB := color.RGBA{
				R: uint8(float64(preheatEndRGB.R)*(1-w.interpolation) + float64(activeRGB.R)*w.interpolation),
				G: uint8(float64(preheatEndRGB.G)*(1-w.interpolation) + float64(activeRGB.G)*w.interpolation),
				B: uint8(float64(preheatEndRGB.B)*(1-w.interpolation) + float64(activeRGB.B)*w.interpolation),
				A: 255,
			}

			glowStrength := 0.15 + w.interpolation*0.5 + pulse*0.15
			glowStrength = clamp(glowStrength, 0, 0.8)

			lcolor = blendColorFast(baseRGB, whiteRGB, glowStrength)
		} else if i == currentWordIndex+1 {
			preheatStrength := currentInterpolation * preheatMax
			lcolor = blendColorFast(inactiveRGB, transitionRGB, preheatStrength)
		} else {
			lcolor = lyricInactiveColor()
		}

		sb.WriteString(util.SetFgStyle(w.text, lcolor))
	}

	return sb.String()
}

// wordState represents the state of a word.
type wordState int

const (
	wordStateNotPlayed wordState = iota
	wordStatePlaying
	wordStatePlayed
)

// wordWithTiming represents a word with its state and interpolation progress.
type wordWithTiming struct {
	text          string
	state         wordState
	interpolation float64 // 0.0-1.0, 词内的插值进度，用于平滑动画
}

// LRC line state for rendering
type lrcLineState int

const (
	lrcLineStateFuture lrcLineState = iota
	lrcLineStatePlaying
	lrcLineStatePlayed
)

// LRC行渲染模式函数 - 整行颜色变化

// renderLRCLineSmooth renders LRC lyrics with smooth mode - smooth color gradient transition.
// Uses eased interpolation for more natural color transitions.
func renderLRCLineSmooth(line string, progress float64) string {
	easedProgress := easeInOutCubic(progress)
	color := blendColorFast(lyricInactiveRGB(), lyricActiveRGB(), easedProgress)
	return util.SetFgStyle(line, color)
}

// renderLRCWave renders LRC lyrics with wave mode - dynamic wave effect across the line.
// Adds animated wave effect that pulses with the music.
func renderLRCWave(line string, progress float64, animationTime float64) string {
	inactiveRGB := lyricInactiveRGB()
	activeRGB := lyricActiveRGB()
	transitionRGB := lyricTransitionRGB()

	baseRGB := color.RGBA{
		R: uint8(float64(inactiveRGB.R)*(1-progress) + float64(activeRGB.R)*progress),
		G: uint8(float64(inactiveRGB.G)*(1-progress) + float64(activeRGB.G)*progress),
		B: uint8(float64(inactiveRGB.B)*(1-progress) + float64(activeRGB.B)*progress),
		A: 255,
	}

	wave := math.Sin(animationTime*2.0) * 0.1
	blendFactor := 0.5 + wave
	finalColor := blendColorFast(transitionRGB, baseRGB, blendFactor)

	return util.SetFgStyle(line, finalColor)
}

// renderLRCGlow renders LRC lyrics with glow mode - soft glow effect.
// The line has a soft glow that pulses gently.
func renderLRCGlow(line string, progress float64, animationTime float64) string {
	inactiveRGB := lyricInactiveRGB()
	activeRGB := lyricActiveRGB()
	whiteRGB := lyricWhiteRGB()

	baseRGB := color.RGBA{
		R: uint8(float64(inactiveRGB.R)*(1-progress) + float64(activeRGB.R)*progress),
		G: uint8(float64(inactiveRGB.G)*(1-progress) + float64(activeRGB.G)*progress),
		B: uint8(float64(inactiveRGB.B)*(1-progress) + float64(activeRGB.B)*progress),
		A: 255,
	}

	pulse := (math.Sin(animationTime*2.0) + 1) / 2
	glowStrength := 0.1 + progress*0.2 + pulse*0.1
	glowStrength = clamp(glowStrength, 0, 0.4)

	finalColor := blendColorFast(baseRGB, whiteRGB, glowStrength)
	return util.SetFgStyle(line, finalColor)
}

// easeInOutCubic provides smooth easing for color transitions.
// Creates natural acceleration and deceleration effect.
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}
