package ui

import (
    "fmt"
    "github.com/lucasb-eyer/go-colorful"
    "github.com/muesli/termenv"
    "math/rand"
    "strconv"
    "strings"
    "time"
)

var (
    termProfile =   termenv.ColorProfile()
    primaryColor    termenv.Color
    primaryColorStr string
)

// GetPrimaryColor get random color
func GetPrimaryColor() termenv.Color {
    if primaryColor != nil {
        return primaryColor
    }
    rand.Seed(time.Now().UnixNano())
    primaryColorStr = strconv.Itoa(rand.Intn(228 - 17) + 17)
    primaryColor = termProfile.Color(primaryColorStr)

    return primaryColor
}

// GetPrimaryColorStr get random color string
func GetPrimaryColorStr() string {
    if primaryColorStr != "" {
        return primaryColorStr
    }
    rand.Seed(time.Now().UnixNano())
    primaryColorStr = strconv.Itoa(rand.Intn(228 - 17) + 17)
    primaryColor = termProfile.Color(primaryColorStr)

    return primaryColorStr
}

// GetRandomRgbColor get random rgb color
func GetRandomRgbColor(isRange bool) (string, string) {
    rand.Seed(time.Now().UnixNano())
    r := 255 - rand.Intn(100)
    rand.Seed(time.Now().UnixNano()/2)
    g := 255 - rand.Intn(100)
    rand.Seed(time.Now().UnixNano()/3)
    b := 255 - rand.Intn(100)

    startColor := fmt.Sprintf("#%x%x%x", r, g, b)
    if !isRange {
        return startColor, ""
    }

    rand.Seed(time.Now().UnixNano()/5)
    rEnd := 50 + rand.Intn(100)
    rand.Seed(time.Now().UnixNano()/7)
    gEnd := 50 + rand.Intn(100)
    rand.Seed(time.Now().UnixNano()/11)
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

// SetNormalStyle dont set any style
func SetNormalStyle(content string) string {
    seq := strings.Join([]string{"0"}, ";")
    return fmt.Sprintf("%s%sm%s%sm", termenv.CSI, seq, content, termenv.CSI+termenv.ResetSeq)
}

// Generate a blend of colors.
func makeRamp(colorA, colorB string, steps float64) (s []string) {
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
