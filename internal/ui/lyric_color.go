package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Lyric color definitions (true color / RGB)
var (
	// LyricActiveColor is the color for played/active lyrics (soft cyan)
	LyricActiveColor = lipgloss.Color("#7EC8E3")
	// LyricTransitionColor is the color for transitioning lyrics (soft lavender)
	LyricTransitionColor = lipgloss.Color("#C9B1D4")
	// LyricInactiveColor is the color for unplayed/inactive lyrics (soft gray)
	LyricInactiveColor = lipgloss.Color("#6B6B6B")
	// LyricWhiteColor is white color for glow effects (soft white)
	LyricWhiteColor = lipgloss.Color("#E8E8E8")
)

// blendColor mixes two lipgloss.Color with a given ratio t (0.0 - 1.0).
func blendColor(c1, c2 lipgloss.Color, t float64) lipgloss.Color {
	r1, g1, b1 := hexToRGB(string(c1))
	r2, g2, b2 := hexToRGB(string(c2))

	r := uint8(float64(r1)*(1-t) + float64(r2)*t)
	g := uint8(float64(g1)*(1-t) + float64(g2)*t)
	b := uint8(float64(b1)*(1-t) + float64(b2)*t)

	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// hexToRGB converts a hex color string to RGB values.
func hexToRGB(hex string) (uint8, uint8, uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 128, 128, 128 // Default gray
	}

	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
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

// renderSimple renders lyrics with simple mode - word-by-word color change with linear interpolation.
// Uses direct linear color transition from inactive to active for smooth, continuous animation.
func renderSimple(words []wordWithTiming) string {
	var sb strings.Builder

	// Use defined colors
	activeColor := LyricActiveColor
	inactiveColor := LyricInactiveColor

	for _, w := range words {
		var color lipgloss.Color

		switch w.state {
		case wordStatePlayed:
			color = activeColor
		case wordStatePlaying:
			// Direct linear interpolation for smooth, continuous color transition
			// No easing applied - let the natural timing drive the animation
			color = blendColor(inactiveColor, activeColor, w.interpolation)
		case wordStateNotPlayed:
			color = inactiveColor
		default:
			color = inactiveColor
		}

		sb.WriteString(util.SetFgStyle(w.text, termenv.RGBColor(string(color))))
	}

	return sb.String()
}

// renderSmooth renders lyrics with smooth mode - uses color gradient transition with interpolation.
// All words participate in gradient calculation for continuous color transitions.
func renderSmooth(words []wordWithTiming, progress float64) string {
	var sb strings.Builder
	totalWords := len(words)
	if totalWords == 0 {
		return ""
	}

	// Use defined colors
	activeColor := LyricActiveColor
	inactiveColor := LyricInactiveColor

	// Transition area width (in words) - matching reference implementation
	fadeWidth := 2.0

	// Find the currently playing word index and its interpolation
	currentWordIdx := -1
	var currentWordInterpolation float64
	for i, w := range words {
		if w.state == wordStatePlaying {
			currentWordIdx = i
			currentWordInterpolation = w.interpolation
			break
		}
	}

	// Calculate precise position - this drives the continuous gradient
	var preciseCurrentPos float64
	if currentWordIdx >= 0 {
		// Use word index + interpolation within that word for sub-word precision
		preciseCurrentPos = float64(currentWordIdx) + currentWordInterpolation
	} else {
		// Fallback to progress-based calculation when no word is playing
		preciseCurrentPos = float64(totalWords) * progress
	}

	// All words participate in gradient calculation - no skipping
	for i, w := range words {
		pos := float64(i)

		// Calculate activation level (0.0 - 1.0) based on position
		var activation float64
		if pos < preciseCurrentPos-fadeWidth {
			activation = 1.0 // Fully activated (past the fade zone)
		} else if pos > preciseCurrentPos {
			activation = 0.0 // Not activated (ahead of current position)
		} else {
			// In transition area - linear interpolation for smooth gradient
			activation = (preciseCurrentPos - pos) / fadeWidth
			activation = clamp(activation, 0, 1)
		}

		// For the currently playing word, ensure color reflects actual progress
		// This prevents the word from appearing darker than its interpolation
		if w.state == wordStatePlaying && w.interpolation > activation {
			activation = w.interpolation
		}

		// For played words, ensure they stay fully active
		if w.state == wordStatePlayed {
			activation = 1.0
		}

		// Blend color based on activation level
		color := blendColor(inactiveColor, activeColor, activation)
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

	// Use defined colors
	activeColor := LyricActiveColor
	transitionColor := LyricTransitionColor
	inactiveColor := LyricInactiveColor

	// Find the currently playing word and use precise position
	var currentWordIdx int = -1
	var currentWordInterpolation float64
	for i, w := range words {
		if w.state == wordStatePlaying {
			currentWordIdx = i
			currentWordInterpolation = w.interpolation
			break
		}
	}

	// Calculate precise current position using linear interpolation
	var preciseCurrentPos float64
	if currentWordIdx >= 0 {
		preciseCurrentPos = float64(currentWordIdx) + currentWordInterpolation
	} else {
		preciseCurrentPos = float64(totalWords) * progress
	}

	for i, w := range words {
		pos := float64(i)

		// Calculate activation with linear transition
		var activation float64
		if pos < preciseCurrentPos-1 {
			activation = 1.0
		} else if pos > preciseCurrentPos+1 {
			activation = 0.0
		} else {
			activation = (preciseCurrentPos - pos + 1) / 2.0
			activation = clamp(activation, 0, 1)
		}

		// For the currently playing word, ensure color reflects actual progress
		if w.state == wordStatePlaying && w.interpolation > activation {
			activation = w.interpolation
		}

		// For played words, ensure they stay fully active
		if w.state == wordStatePlayed {
			activation = 1.0
		}

		var color lipgloss.Color
		if activation >= 1.0 {
			// Add wave effect for fully played words
			wave := math.Sin(animationTime*3.0 - float64(i)*0.5)
			wave = (wave + 1) / 2 // Normalize to 0-1

			// Oscillate between active color and transition color
			color = blendColor(transitionColor, activeColor, wave)
		} else if activation > 0 {
			// Transitioning word - direct linear blend
			color = blendColor(inactiveColor, activeColor, activation)
		} else {
			color = inactiveColor
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

	// Use defined colors
	activeColor := LyricActiveColor
	transitionColor := LyricTransitionColor
	inactiveColor := LyricInactiveColor
	whiteColor := LyricWhiteColor

	// Get current word's interpolation for transition calculations
	var currentInterpolation float64
	if currentWordIndex >= 0 && currentWordIndex < len(words) {
		currentInterpolation = words[currentWordIndex].interpolation
	}

	// Glow pulse - simple sin wave for smooth animation
	pulse := (math.Sin(animationTime*2.0) + 1) / 2

	// Preheat intensity - how much the next word lights up before playing
	const preheatMax = 0.4

	for i, w := range words {
		var color lipgloss.Color

		if i < currentWordIndex-1 {
			// Words played long ago - use active color
			color = activeColor
		} else if i == currentWordIndex-1 {
			// Just finished word - fade out glow smoothly
			// Use current word's interpolation to control fade out
			// When currentInterpolation = 0, still show glow (just finished)
			// When currentInterpolation approaches 1, glow fades to activeColor
			fadeOut := currentInterpolation
			glowStrength := (1 - fadeOut) * (0.3 + pulse*0.15)
			color = blendColor(activeColor, whiteColor, glowStrength)
		} else if i == currentWordIndex {
			// Current playing word - glow effect with smooth entry
			// interpolation = 0: color should be close to preheat end state
			// interpolation = 1: color should be full glow

			// Preheat end color (what the word looked like just before playing)
			preheatEndColor := blendColor(inactiveColor, transitionColor, preheatMax)

			// Base color transitions from preheat end to active
			baseColor := blendColor(preheatEndColor, activeColor, w.interpolation)

			// Glow strength increases with interpolation
			glowStrength := 0.15 + w.interpolation*0.5 + pulse*0.15
			glowStrength = clamp(glowStrength, 0, 0.8)

			color = blendColor(baseColor, whiteColor, glowStrength)
		} else if i == currentWordIndex+1 {
			// Next word - preheat based on current word's progress
			// Lights up gradually as current word plays
			preheatStrength := currentInterpolation * preheatMax
			color = blendColor(inactiveColor, transitionColor, preheatStrength)
		} else {
			// Words not yet played
			color = inactiveColor
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
