package configs

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
)

// ThemeFile represents a TOML-mappable theme definition.
type ThemeFile struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Dark        ThemeVariant      `toml:"dark"`
	Light       ThemeVariant      `toml:"light"`
}

// ThemeVariant holds palette, highlight, popup, notification, and app-specific colors.
type ThemeVariant struct {
	Primary      string `toml:"primary"`
	Secondary    string `toml:"secondary"`
	Accent       string `toml:"accent"`
	Success      string `toml:"success"`
	Warning      string `toml:"warning"`
	Error        string `toml:"error"`
	Info         string `toml:"info"`
	Muted        string `toml:"muted"`
	HintKey      string `toml:"hintKey"`
	Background   string `toml:"background"`
	Foreground   string `toml:"foreground"`
	Border       string `toml:"border"`
	Surface      string `toml:"surface"`

	Highlights   HighlightsConfig          `toml:"highlights"`
	Popup        PopupConfig               `toml:"popup"`
	Notification ThemeNotificationConfig   `toml:"notification"`
	App          AppColorConfig            `toml:"app"`
}

// HighlightsConfig allows overriding specific Highlight fields.
type HighlightsConfig struct {
	MenuTitle string `toml:"menuTitle"`
	// Further highlight overrides can be added here as needed.
}

// PopupConfig mirrors the foxful-cli PopupTheme.
type PopupConfig struct {
	Surface string `toml:"surface"`
	Border  string `toml:"border"`
}

// ThemeNotificationConfig mirrors the foxful-cli NotificationTheme.
type ThemeNotificationConfig struct {
	Surface       string `toml:"surface"`
	InfoBorder    string `toml:"infoBorder"`
	SuccessBorder string `toml:"successBorder"`
	WarningBorder string `toml:"warningBorder"`
	ErrorBorder   string `toml:"errorBorder"`
}

// AppColorConfig holds go-musicfox specific custom colors.
type AppColorConfig struct {
	// Lyric colors
	LyricActive      string `toml:"lyricActive"`
	LyricTransition  string `toml:"lyricTransition"`
	LyricInactive    string `toml:"lyricInactive"`
	LyricHighlight   string `toml:"lyricHighlight"`

	// Playbar colors
	PlaybarMode         string `toml:"playbarMode"`
	PlaybarVolume       string `toml:"playbarVolume"`
	PlaybarPlaying      string `toml:"playbarPlaying"`
	PlaybarPaused       string `toml:"playbarPaused"`
	PlaybarHeartLiked   string `toml:"playbarHeartLiked"`
	PlaybarHeartUnliked string `toml:"playbarHeartUnliked"`
	PlaybarArtist       string `toml:"playbarArtist"`

	// Visualizer colors: "random" or hex color
	VisualizerColorStart string `toml:"visualizerColorStart"`
	VisualizerColorEnd   string `toml:"visualizerColorEnd"`
}

// toTheme converts a ThemeVariant to a style.Theme.
func (v ThemeVariant) toTheme() style.Theme {
	t := style.DefaultDarkTheme()

	// Palette
	if v.Primary != "" {
		t.Primary = parseColor(v.Primary)
	}
	if v.Secondary != "" {
		t.Secondary = parseColor(v.Secondary)
	}
	if v.Accent != "" {
		t.Accent = parseColor(v.Accent)
	}
	if v.Success != "" {
		t.Success = parseColor(v.Success)
	}
	if v.Warning != "" {
		t.Warning = parseColor(v.Warning)
	}
	if v.Error != "" {
		t.Error = parseColor(v.Error)
	}
	if v.Info != "" {
		t.Info = parseColor(v.Info)
	}
	if v.Muted != "" {
		t.Muted = parseColor(v.Muted)
	}
	if v.HintKey != "" {
		t.HintKey = parseColor(v.HintKey)
	}

	// Terminal-level
	if v.Background != "" {
		t.Background = parseColor(v.Background)
	}
	if v.Foreground != "" {
		t.Foreground = parseColor(v.Foreground)
	}
	if v.Border != "" {
		t.Border = parseColor(v.Border)
	}
	if v.Surface != "" {
		t.Surface = parseColor(v.Surface)
	}

	// Highlights
	if v.Highlights.MenuTitle != "" {
		t.MenuTitle = style.Highlight{Fg: parseColor(v.Highlights.MenuTitle)}
	}

	// Popup
	if v.Popup.Surface != "" {
		t.Popup.Surface = parseColor(v.Popup.Surface)
	}
	if v.Popup.Border != "" {
		t.Popup.Border = parseColor(v.Popup.Border)
	}

	// Notification
	if v.Notification.Surface != "" {
		t.Notification.Surface = parseColor(v.Notification.Surface)
	}
	if v.Notification.InfoBorder != "" {
		t.Notification.InfoBorder = parseColor(v.Notification.InfoBorder)
	}
	if v.Notification.SuccessBorder != "" {
		t.Notification.SuccessBorder = parseColor(v.Notification.SuccessBorder)
	}
	if v.Notification.WarningBorder != "" {
		t.Notification.WarningBorder = parseColor(v.Notification.WarningBorder)
	}
	if v.Notification.ErrorBorder != "" {
		t.Notification.ErrorBorder = parseColor(v.Notification.ErrorBorder)
	}

	// Custom app styles (wrapped as Highlight for foxful-cli resolution)
	t.Custom = v.App.toCustomStyles()

	return t
}

// AppCustomStyles is the struct passed to Theme.Custom for foxful-cli reflection.
// All fields must be of type style.Highlight.
type AppCustomStyles struct {
	LyricActive      style.Highlight
	LyricTransition  style.Highlight
	LyricInactive    style.Highlight
	LyricHighlight   style.Highlight
	PlaybarMode      style.Highlight
	PlaybarVolume    style.Highlight
	PlaybarPlaying   style.Highlight
	PlaybarPaused    style.Highlight
	PlaybarHeartLiked  style.Highlight
	PlaybarHeartUnliked style.Highlight
	PlaybarArtist    style.Highlight
}

// toCustomStyles converts AppColorConfig to an AppCustomStyles for Theme.Custom.
func (c AppColorConfig) toCustomStyles() AppCustomStyles {
	return AppCustomStyles{
		LyricActive:       fgHighlight(c.LyricActive, LyricActiveColor),
		LyricTransition:   fgHighlight(c.LyricTransition, LyricTransitionColor),
		LyricInactive:     fgHighlight(c.LyricInactive, LyricInactiveColor),
		LyricHighlight:    fgHighlight(c.LyricHighlight, LyricWhiteColor),
		PlaybarMode:       fgHighlight(c.PlaybarMode, PlaybarModeColor),
		PlaybarVolume:     fgHighlight(c.PlaybarVolume, PlaybarVolumeColor),
		PlaybarPlaying:    fgHighlight(c.PlaybarPlaying, PlaybarPlayingColor),
		PlaybarPaused:     fgHighlight(c.PlaybarPaused, PlaybarPausedColor),
		PlaybarHeartLiked:   fgHighlight(c.PlaybarHeartLiked, PlaybarHeartLikedColor),
		PlaybarHeartUnliked: fgHighlight(c.PlaybarHeartUnliked, PlaybarHeartUnlikedColor),
		PlaybarArtist:     fgHighlight(c.PlaybarArtist, PlaybarArtistColor),
	}
}

// fgHighlight creates a Highlight with the given foreground color,
// falling back to defaultHex if empty or "random".
func fgHighlight(hex, defaultHex string) style.Highlight {
	c := parseAppColor(hex, defaultHex)
	return style.Highlight{Fg: c}
}

// parseAppColor resolves an app color: empty → default, "random" → random, otherwise hex.
func parseAppColor(hex, defaultHex string) color.Color {
	switch {
	case hex == "":
		return parseColor(defaultHex)
	case strings.EqualFold(hex, "random"):
		r, g, b := rand.Intn(156)+100, rand.Intn(156)+100, rand.Intn(156)+100
		return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
	default:
		return parseColor(hex)
	}
}

// ansiColorNames maps lipgloss ANSI color names to their constants.
// When used in theme TOML, these are rendered by the terminal's color scheme,
// not as fixed 24-bit hex values.
var ansiColorNames = map[string]color.Color{
	"Black":        lipgloss.Black,
	"Red":          lipgloss.Red,
	"Green":        lipgloss.Green,
	"Yellow":       lipgloss.Yellow,
	"Blue":         lipgloss.Blue,
	"Magenta":      lipgloss.Magenta,
	"Cyan":         lipgloss.Cyan,
	"White":        lipgloss.White,
	"BrightBlack":  lipgloss.BrightBlack,
	"BrightRed":    lipgloss.BrightRed,
	"BrightGreen":  lipgloss.BrightGreen,
	"BrightYellow": lipgloss.BrightYellow,
	"BrightBlue":   lipgloss.BrightBlue,
	"BrightMagenta": lipgloss.BrightMagenta,
	"BrightCyan":   lipgloss.BrightCyan,
	"BrightWhite":  lipgloss.BrightWhite,
}

func parseColor(s string) color.Color {
	if c, ok := ansiColorNames[s]; ok {
		return c
	}
	return lipgloss.Color(s)
}

// ResolveVisualizerColors resolves the visualizer start/end colors.
// Returns the color hex strings, applying "random" logic when configured.
func (c AppColorConfig) ResolveVisualizerColors() (start, end string) {
	if c.VisualizerColorStart == "" || strings.EqualFold(c.VisualizerColorStart, "random") ||
		c.VisualizerColorEnd == "" || strings.EqualFold(c.VisualizerColorEnd, "random") {
		return util.GetRandomRgbColor(true)
	}
	return c.VisualizerColorStart, c.VisualizerColorEnd
}

// IsVisualizerRandom returns true when the visualizer should use random colors.
func (c AppColorConfig) IsVisualizerRandom() bool {
	return c.VisualizerColorStart == "" || strings.EqualFold(c.VisualizerColorStart, "random") ||
		c.VisualizerColorEnd == "" || strings.EqualFold(c.VisualizerColorEnd, "random")
}

// Default application color constants (fallback when no theme is configured).
// These match the visual appearance of the original hardcoded colors:
// - Lyric colors were already hex values
// - Playbar colors were lipgloss ANSI constants; hex values are terminal-typical approximations.
const (
	// Lyric (original hex values)
	LyricActiveColor     = "#7EC8E3"
	LyricTransitionColor = "#C9B1D4"
	LyricInactiveColor   = "#6B6B6B"
	LyricWhiteColor      = "#E8E8E8"

	// Playbar (ANSI color approximations matching typical terminal palettes)
	PlaybarModeColor         = "#FF55FF" // BrightMagenta
	PlaybarVolumeColor       = "#5555FF" // BrightBlue
	PlaybarPlayingColor      = "#FFFF55" // BrightYellow
	PlaybarPausedColor       = "#CCCC00" // Yellow
	PlaybarHeartLikedColor   = "#CC0000" // Red
	PlaybarHeartUnlikedColor = "#CCCCCC" // White (ANSI White = 7, not BrightWhite = 15)
	PlaybarArtistColor       = "#767676" // BrightBlack
)

// ResolvedAppStyles mirrors AppCustomStyles with lipgloss.Style fields,
// matching what foxful-cli's StyleSet.Custom map contains after resolution.
// Use style.CustomStyles[ResolvedAppStyles](ss) to obtain an instance.
type ResolvedAppStyles struct {
	LyricActive        lipgloss.Style
	LyricTransition    lipgloss.Style
	LyricInactive      lipgloss.Style
	LyricHighlight     lipgloss.Style
	PlaybarMode        lipgloss.Style
	PlaybarVolume      lipgloss.Style
	PlaybarPlaying     lipgloss.Style
	PlaybarPaused      lipgloss.Style
	PlaybarHeartLiked  lipgloss.Style
	PlaybarHeartUnliked lipgloss.Style
	PlaybarArtist      lipgloss.Style
}

// GetCurrentAppColors returns the resolved custom app styles from the current foxful-cli StyleSet.
func GetCurrentAppColors() ResolvedAppStyles {
	return style.CustomStyles[ResolvedAppStyles](style.CurrentStyleSet())
}

// SafeGetForeground extracts the foreground color from a lipgloss.Style,
// falling back to the provided default color string (ANSI name or hex).
func SafeGetForeground(s lipgloss.Style, defaultStr string) color.Color {
	if c := s.GetForeground(); c != nil {
		return c
	}
	return parseColor(defaultStr)
}