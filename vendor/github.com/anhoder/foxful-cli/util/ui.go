package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

var (
	TermProfile      = termenv.ColorProfile()
	PrimaryColor     string
	_primaryColor    termenv.Color
	_primaryColorStr string
)

// GetPrimaryColor get random color
func GetPrimaryColor() termenv.Color {
	if _primaryColor != nil {
		return _primaryColor
	}
	initPrimaryColor()
	return _primaryColor
}

func GetPrimaryColorString() string {
	if _primaryColorStr != "" {
		return _primaryColorStr
	}
	initPrimaryColor()
	return _primaryColorStr
}

func initPrimaryColor() {
	if _primaryColorStr != "" && _primaryColor != nil {
		return
	}
	if PrimaryColor == "" || PrimaryColor == RandomColor {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		_primaryColorStr = strconv.Itoa(rand.Intn(228-17) + 17)
	} else {
		_primaryColorStr = PrimaryColor
	}
	_primaryColor = TermProfile.Color(GetPrimaryColorString())
}

// GetRandomRgbColor get random rgb color
func GetRandomRgbColor(isRange bool) (string, string) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	r := 255 - rand.Intn(100)
	rand.New(rand.NewSource(time.Now().UnixNano() / 2))
	g := 255 - rand.Intn(100)
	rand.New(rand.NewSource(time.Now().UnixNano() / 3))
	b := 255 - rand.Intn(100)

	startColor := fmt.Sprintf("#%x%x%x", r, g, b)
	if !isRange {
		return startColor, ""
	}

	rand.New(rand.NewSource(time.Now().UnixNano() / 5))
	rEnd := 50 + rand.Intn(100)
	rand.New(rand.NewSource(time.Now().UnixNano() / 7))
	gEnd := 50 + rand.Intn(100)
	rand.New(rand.NewSource(time.Now().UnixNano() / 11))
	bEnd := 50 + rand.Intn(100)
	endColor := fmt.Sprintf("#%x%x%x", rEnd, gEnd, bEnd)

	return startColor, endColor
}

// SetFgStyle Return a function that will colorize the foreground of a given string.
func SetFgStyle(content string, color termenv.Color) string {
	return termenv.Style{}.Foreground(color).Styled(content)
}

// SetFgBgStyle Color a string's foreground and background with the given value.
func SetFgBgStyle(content string, fg, bg termenv.Color) string {
	return termenv.Style{}.Foreground(fg).Background(bg).Styled(content)
}

// SetNormalStyle don't set any style
func SetNormalStyle(content string) string {
	seq := strings.Join([]string{"0"}, ";")
	return fmt.Sprintf("%s%sm%s%sm", termenv.CSI, seq, content, termenv.CSI+termenv.ResetSeq)
}

func GetPrimaryFontStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(GetPrimaryColorString()))
}

// MakeRamp Generate a blend of colors.
func MakeRamp(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}

// Convert a colorful.Color to a hexidecimal format compatible with termenv.
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and 1.
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}
