package ui

import (
	"math"
	"strings"

	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Lyric color definitions (true color / RGB)
var (
	LyricActiveColor     = lipgloss.Color("#7EC8E3")
	LyricTransitionColor = lipgloss.Color("#C9B1D4")
	LyricInactiveColor   = lipgloss.Color("#6B6B6B")
	LyricWhiteColor      = lipgloss.Color("#E8E8E8")
)

// colorRGB holds pre-computed RGB values for fast blending
type colorRGB struct {
	R, G, B uint8
}

// Pre-computed RGB values - avoid parsing hex every frame
var (
	lyricActiveRGB     = colorRGB{0x7E, 0xC8, 0xE3}
	lyricTransitionRGB = colorRGB{0xC9, 0xB1, 0xD4}
	lyricInactiveRGB   = colorRGB{0x6B, 0x6B, 0x6B}
	lyricWhiteRGB      = colorRGB{0xE8, 0xE8, 0xE8}
)

// hexDigits for fast hex encoding
const hexDigits = "0123456789ABCDEF"

// blendColorFast mixes two pre-computed RGB colors with ratio t (0.0 - 1.0).
func blendColorFast(c1, c2 colorRGB, t float64) lipgloss.Color {
	t1 := 1 - t
	r := uint8(float64(c1.R)*t1 + float64(c2.R)*t)
	g := uint8(float64(c1.G)*t1 + float64(c2.G)*t)
	b := uint8(float64(c1.B)*t1 + float64(c2.B)*t)

	// Fast hex encoding without fmt.Sprintf
	var buf [7]byte
	buf[0] = '#'
	buf[1] = hexDigits[r>>4]
	buf[2] = hexDigits[r&0x0F]
	buf[3] = hexDigits[g>>4]
	buf[4] = hexDigits[g&0x0F]
	buf[5] = hexDigits[b>>4]
	buf[6] = hexDigits[b&0x0F]

	return lipgloss.Color(string(buf[:]))
}

// blendColor mixes two lipgloss.Color with a given ratio t (0.0 - 1.0).
// Legacy function for backward compatibility.
func blendColor(c1, c2 lipgloss.Color, t float64) lipgloss.Color {
	r1, g1, b1 := hexToRGB(string(c1))
	r2, g2, b2 := hexToRGB(string(c2))
	return blendColorFast(colorRGB{r1, g1, b1}, colorRGB{r2, g2, b2}, t)
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 128, 128, 128
	}

	hexToByte := func(c byte) uint8 {
		switch {
		case c >= '0' && c <= '9':
			return c - '0'
		case c >= 'a' && c <= 'f':
			return c - 'a' + 10
		case c >= 'A' && c <= 'F':
			return c - 'A' + 10
		default:
			return 0
		}
	}

	r := hexToByte(hex[0])<<4 | hexToByte(hex[1])
	g := hexToByte(hex[2])<<4 | hexToByte(hex[3])
	b := hexToByte(hex[4])<<4 | hexToByte(hex[5])
	return r, g, b
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

		color := blendColorFast(lyricInactiveRGB, lyricActiveRGB, activation)
		sb.WriteString(util.SetFgStyle(w.text, termenv.RGBColor(string(color))))
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

		var color lipgloss.Color
		if activation >= 1.0 {
			wave := math.Sin(animationTime*3.0 - float64(i)*0.5)
			wave = (wave + 1) / 2
			color = blendColorFast(lyricTransitionRGB, lyricActiveRGB, wave)
		} else if activation > 0 {
			color = blendColorFast(lyricInactiveRGB, lyricActiveRGB, activation)
		} else {
			color = LyricInactiveColor
		}

		sb.WriteString(util.SetFgStyle(w.text, termenv.RGBColor(string(color))))
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

	for i, w := range words {
		var color lipgloss.Color

		if i < currentWordIndex-1 {
			color = LyricActiveColor
		} else if i == currentWordIndex-1 {
			fadeOut := currentInterpolation
			glowStrength := (1 - fadeOut) * (0.3 + pulse*0.15)
			color = blendColorFast(lyricActiveRGB, lyricWhiteRGB, glowStrength)
		} else if i == currentWordIndex {
			preheatEndRGB := colorRGB{
				R: uint8(float64(lyricInactiveRGB.R)*(1-preheatMax) + float64(lyricTransitionRGB.R)*preheatMax),
				G: uint8(float64(lyricInactiveRGB.G)*(1-preheatMax) + float64(lyricTransitionRGB.G)*preheatMax),
				B: uint8(float64(lyricInactiveRGB.B)*(1-preheatMax) + float64(lyricTransitionRGB.B)*preheatMax),
			}

			baseRGB := colorRGB{
				R: uint8(float64(preheatEndRGB.R)*(1-w.interpolation) + float64(lyricActiveRGB.R)*w.interpolation),
				G: uint8(float64(preheatEndRGB.G)*(1-w.interpolation) + float64(lyricActiveRGB.G)*w.interpolation),
				B: uint8(float64(preheatEndRGB.B)*(1-w.interpolation) + float64(lyricActiveRGB.B)*w.interpolation),
			}

			glowStrength := 0.15 + w.interpolation*0.5 + pulse*0.15
			glowStrength = clamp(glowStrength, 0, 0.8)

			color = blendColorFast(baseRGB, lyricWhiteRGB, glowStrength)
		} else if i == currentWordIndex+1 {
			preheatStrength := currentInterpolation * preheatMax
			color = blendColorFast(lyricInactiveRGB, lyricTransitionRGB, preheatStrength)
		} else {
			color = LyricInactiveColor
		}

		sb.WriteString(util.SetFgStyle(w.text, termenv.RGBColor(string(color))))
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

// LRC渲染模式函数 - 整行颜色变化

// renderLRCLineSmooth renders LRC lyrics with smooth mode - smooth color gradient transition.
// Uses eased interpolation for more natural color transitions.
func renderLRCLineSmooth(line string, progress float64) string {
	easedProgress := easeInOutCubic(progress)
	color := blendColorFast(lyricInactiveRGB, lyricActiveRGB, easedProgress)
	return util.SetFgStyle(line, termenv.RGBColor(string(color)))
}

// renderLRCWave renders LRC lyrics with wave mode - dynamic wave effect across the line.
// Adds animated wave effect that pulses with the music.
func renderLRCWave(line string, progress float64, animationTime float64) string {
	baseRGB := colorRGB{
		R: uint8(float64(lyricInactiveRGB.R)*(1-progress) + float64(lyricActiveRGB.R)*progress),
		G: uint8(float64(lyricInactiveRGB.G)*(1-progress) + float64(lyricActiveRGB.G)*progress),
		B: uint8(float64(lyricInactiveRGB.B)*(1-progress) + float64(lyricActiveRGB.B)*progress),
	}

	wave := math.Sin(animationTime*2.0) * 0.1
	blendFactor := 0.5 + wave
	finalColor := blendColorFast(lyricTransitionRGB, baseRGB, blendFactor)

	return util.SetFgStyle(line, termenv.RGBColor(string(finalColor)))
}

// renderLRCGlow renders LRC lyrics with glow mode - soft glow effect.
// The line has a soft glow that pulses gently.
func renderLRCGlow(line string, progress float64, animationTime float64) string {
	baseRGB := colorRGB{
		R: uint8(float64(lyricInactiveRGB.R)*(1-progress) + float64(lyricActiveRGB.R)*progress),
		G: uint8(float64(lyricInactiveRGB.G)*(1-progress) + float64(lyricActiveRGB.G)*progress),
		B: uint8(float64(lyricInactiveRGB.B)*(1-progress) + float64(lyricActiveRGB.B)*progress),
	}

	pulse := (math.Sin(animationTime*2.0) + 1) / 2
	glowStrength := 0.1 + progress*0.2 + pulse*0.1
	glowStrength = clamp(glowStrength, 0, 0.4)

	finalColor := blendColorFast(baseRGB, lyricWhiteRGB, glowStrength)
	return util.SetFgStyle(line, termenv.RGBColor(string(finalColor)))
}

// easeInOutCubic provides smooth easing for color transitions.
// Creates natural acceleration and deceleration effect.
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}
