package util

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	TermProfile      = colorprofile.Detect(os.Stdout, os.Environ())
	PrimaryColor     string
	_primaryColor    color.Color
	_primaryColorStr string
	primaryColorOnce sync.Once

	// HasDarkBackground indicates whether the terminal has a dark background.
	// Defaults to true (dark backgrounds are common in terminals).
	// Callers can override this before calling LightDark() to control adaptive color selection.
	HasDarkBackground = true
)

// GetPrimaryColor get random color
func GetPrimaryColor() color.Color {
	ensurePrimaryColorInit()
	return _primaryColor
}

func GetPrimaryColorString() string {
	ensurePrimaryColorInit()
	return _primaryColorStr
}

func ensurePrimaryColorInit() {
	primaryColorOnce.Do(func() {
		if PrimaryColor == "" || PrimaryColor == RandomColor {
			rand.New(rand.NewSource(time.Now().UnixNano()))
			_primaryColorStr = strconv.Itoa(rand.Intn(228-17) + 17)
		} else {
			_primaryColorStr = PrimaryColor
		}
		_primaryColor = lipgloss.Color(_primaryColorStr)
	})
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

// LightDark returns darkColor when HasDarkBackground is true (terminal has a dark background),
// and lightColor when HasDarkBackground is false (terminal has a light background).
func LightDark(lightColor, darkColor color.Color) color.Color {
	if HasDarkBackground {
		return darkColor
	}
	return lightColor
}

// SetFgStyle Return a function that will colorize the foreground of a given string.
//
// Deprecated: Use style.FG() instead.
func SetFgStyle(content string, fg color.Color) string {
	return lipgloss.NewStyle().Foreground(fg).Render(content)
}

// SetFgBgStyle Color a string's foreground and background with the given value.
//
// Deprecated: Use style.FGBG() instead.
func SetFgBgStyle(content string, fg, bg color.Color) string {
	return lipgloss.NewStyle().Foreground(fg).Background(bg).Render(content)
}

// SetNormalStyle don't set any style
//
// Deprecated: Use style.Normal() instead.
func SetNormalStyle(content string) string {
	return fmt.Sprintf("\x1b[0m%s\x1b[0m", content)
}

// GetPrimaryFontStyle returns a lipgloss style with the primary color as the foreground.
// Pass true to make the text bold.
//
// Deprecated: The no-arg call GetPrimaryFontStyle() defaults to non-bold.
// Prefer passing the bold parameter explicitly: GetPrimaryFontStyle(true) or GetPrimaryFontStyle(false).
func GetPrimaryFontStyle(bold ...bool) lipgloss.Style {
	isBold := false
	if len(bold) > 0 {
		isBold = bold[0]
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(GetPrimaryColorString()))
	if isBold {
		style = style.Bold(true)
	}
	return style
}

// MakeRamp Generate a blend of colors.
func MakeRamp(colorA, colorB string, steps float64) (s []color.Color) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, c)
	}
	return
}
