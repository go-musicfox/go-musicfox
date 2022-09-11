package term

import (
	"runtime"
	"strings"
)

// Style codes.
const (
	Default   = "0"
	Bold      = "1"
	Dim       = "2"
	Underline = "4"
	Inverse   = "7"
	Hidden    = "8"
	Strikeout = "9"
)

// Foreground color codes.
const (
	FgBlack   = "30"
	FgRed     = "31"
	FgGreen   = "32"
	FgYellow  = "33"
	FgBlue    = "34"
	FgMagenta = "35"
	FgCyan    = "36"
	FgWhite   = "37"
)

// Background color codes.
const (
	BgBlack   = "40"
	BgRed     = "41"
	BgGreen   = "42"
	BgYellow  = "43"
	BgBlue    = "44"
	BgMagenta = "45"
	BgCyan    = "46"
	BgWhite   = "47"
)

// Color returns a colored text based on the specified style and color codes.
func Color(text string, colors ...string) string {
	// The windows terminal has no color support.
	if runtime.GOOS == "windows" {
		return text
	}
	return esc(strings.Join(colors, ";")) + text + esc(Default)
}

// Start and stop escape sequences.
const (
	escStart = "\x1b["
	escStop  = "m"
)

// esc appeneds prefix and suffix escape sequences to the provided string.
func esc(s string) string {
	return escStart + s + escStop
}

// Blue returns a blue text on a black background.
func Blue(text string) string {
	return Color(text, FgBlue)
}

// Cyan returns a cyan text on a black background.
func Cyan(text string) string {
	return Color(text, FgCyan)
}

// Green returns a green text on a black background.
func Green(text string) string {
	return Color(text, FgGreen)
}

// Magenta returns a magenta text on a black background.
func Magenta(text string) string {
	return Color(text, FgMagenta)
}

// Red returns a red text on a black background.
func Red(text string) string {
	return Color(text, FgRed)
}

// White returns a white text on a black background.
func White(text string) string {
	return Color(text, FgWhite)
}

// Yellow returns a yellow text on a black background.
func Yellow(text string) string {
	return Color(text, FgYellow)
}

// BlueBold returns a bold blue text on a black background.
func BlueBold(text string) string {
	return Color(text, FgBlue, Bold)
}

// CyanBold returns a bold cyan text on a black background.
func CyanBold(text string) string {
	return Color(text, FgCyan, Bold)
}

// GreenBold returns a bold green text on a black background.
func GreenBold(text string) string {
	return Color(text, FgGreen, Bold)
}

// MagentaBold returns a bold magenta text on a black background.
func MagentaBold(text string) string {
	return Color(text, FgMagenta, Bold)
}

// RedBold returns a bold red text on a black background.
func RedBold(text string) string {
	return Color(text, FgRed, Bold)
}

// WhiteBold returns a bold white text on a black background.
func WhiteBold(text string) string {
	return Color(text, FgWhite, Bold)
}

// YellowBold returns a bold yellow text on a black background.
func YellowBold(text string) string {
	return Color(text, FgYellow, Bold)
}
