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
	Name        string       `toml:"name"`
	Description string       `toml:"description"`
	Dark        ThemeVariant `toml:"dark"`
	Light       ThemeVariant `toml:"light"`
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

	Highlights   HighlightsConfig        `toml:"highlights"`
	Popup        PopupConfig             `toml:"popup"`
	Notification ThemeNotificationConfig `toml:"notification"`
	StatusBar    StatusBarConfig         `toml:"statusBar"`
	Markdown     MarkdownConfig          `toml:"markdown"`
	App          AppColorConfig          `toml:"app"`
}

// HighlightsConfig allows overriding specific Highlight fields.
type HighlightsConfig struct {
	MenuTitle         string `toml:"menuTitle"`
	MenuItem          string `toml:"menuItem"`
	SelectedItem      string `toml:"selectedItem"`
	SelectedItemBg    string `toml:"selectedItemBg"`
	MenuItemHover     string `toml:"menuItemHover"`
	SelectedItemHover string `toml:"selectedItemHover"`
	Title             string `toml:"title"`
	Prompt            string `toml:"prompt"`
	BackButton        string `toml:"backButton"`
	BackButtonHover   string `toml:"backButtonHover"`
	Subtitle          string `toml:"subtitle"`
	ProgressEmpty     string `toml:"progressEmpty"`
	ScrollTrack       string `toml:"scrollTrack"`
	ScrollThumb       string `toml:"scrollThumb"`
	Button            string `toml:"button"`
	ButtonBlurred     string `toml:"buttonBlurred"`
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

// StatusBarConfig maps to foxful-cli StyleSet status bar fields.
type StatusBarConfig struct {
	Bar            string `toml:"bar"`
	Text           string `toml:"text"`
	Breadcrumb     string `toml:"breadcrumb"`
	BreadcrumbSep  string `toml:"breadcrumbSep"`
	BreadcrumbHover string `toml:"breadcrumbHover"`
	BreadcrumbClick string `toml:"breadcrumbClick"`
	Time           string `toml:"time"`
	Nugget         string `toml:"nugget"`
	NuggetLabel    string `toml:"nuggetLabel"`
}

// MarkdownConfig holds glamour markdown theme presets.
// Each field accepts a glamour built-in theme name (e.g. "dracula", "dark") or a file path to a custom JSON theme.
type MarkdownConfig struct {
	Dark  string `toml:"dark"`
	Light string `toml:"light"`
}

// AppColorConfig holds go-musicfox specific custom colors.
type AppColorConfig struct {
	// Lyric colors
	LyricActive     string `toml:"lyricActive"`
	LyricTransition string `toml:"lyricTransition"`
	LyricInactive   string `toml:"lyricInactive"`
	LyricHighlight  string `toml:"lyricHighlight"`

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

	// App background
	Background string `toml:"background"`

	// Context menu (reuses foxful-cli Popup theme)
	ContextMenuBg     string `toml:"contextMenuBg"`
	ContextMenuBorder string `toml:"contextMenuBorder"`

	// Spectrum vertical gradient blend color
	SpectrumVerticalBg string `toml:"spectrumVerticalBg"`

	// Desktop lyrics background
	DesktopLyricsBg string `toml:"desktopLyricsBg"`

	// Login/auth status colors
	LoginError   string `toml:"loginError"`
	LoginSuccess string `toml:"loginSuccess"`
	AuthStatus   string `toml:"authStatus"`

	// CLI config table
	ConfigTableBorder     string `toml:"configTableBorder"`
	ConfigTableSelectedFg string `toml:"configTableSelectedFg"`
	ConfigTableSelectedBg string `toml:"configTableSelectedBg"`
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
	h := v.Highlights
	if h.MenuTitle != "" {
		t.MenuTitle = fgPtr(h.MenuTitle)
	}
	if h.MenuItem != "" {
		t.MenuItem = fgPtr(h.MenuItem)
	}
	if h.SelectedItem != "" {
		t.SelectedItem = fgPtr(h.SelectedItem)
	}
	if h.SelectedItemBg != "" {
		t.SelectedItemBg = style.Highlight{Bg: parseColor(h.SelectedItemBg)}
	}
	if h.MenuItemHover != "" {
		t.MenuItemHover = fgPtr(h.MenuItemHover)
	}
	if h.SelectedItemHover != "" {
		t.SelectedItemHover = fgPtr(h.SelectedItemHover)
	}
	if h.Title != "" {
		t.Title = fgPtr(h.Title)
	}
	if h.Prompt != "" {
		t.Prompt = fgPtr(h.Prompt)
	}
	if h.BackButton != "" {
		t.BackButton = fgPtr(h.BackButton)
	}
	if h.BackButtonHover != "" {
		t.BackButtonHover = fgPtr(h.BackButtonHover)
	}
	if h.Subtitle != "" {
		t.Subtitle = fgPtr(h.Subtitle)
	}
	if h.ProgressEmpty != "" {
		t.ProgressEmpty = fgPtr(h.ProgressEmpty)
	}
	if h.ScrollTrack != "" {
		t.ScrollTrack = fgPtr(h.ScrollTrack)
	}
	if h.ScrollThumb != "" {
		t.ScrollThumb = fgPtr(h.ScrollThumb)
	}
	if h.Button != "" {
		t.Button = fgPtr(h.Button)
	}
	if h.ButtonBlurred != "" {
		t.ButtonBlurred = fgPtr(h.ButtonBlurred)
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

	// Status bar
	sb := v.StatusBar
	if sb.Bar != "" {
		t.StatusBar = fgPtr(sb.Bar)
	}
	if sb.Text != "" {
		t.StatusBarText = fgPtr(sb.Text)
	}
	if sb.Breadcrumb != "" {
		t.StatusBarBreadcrumb = fgPtr(sb.Breadcrumb)
	}
	if sb.BreadcrumbSep != "" {
		t.StatusBarBreadcrumbSep = parseColor(sb.BreadcrumbSep)
	}
	if sb.BreadcrumbHover != "" {
		t.StatusBarBreadcrumbHover = fgPtr(sb.BreadcrumbHover)
	}
	if sb.BreadcrumbClick != "" {
		t.StatusBarBreadcrumbClick = fgPtr(sb.BreadcrumbClick)
	}
	if sb.Time != "" {
		t.StatusBarTime = fgPtr(sb.Time)
	}
	if sb.Nugget != "" {
		t.StatusBarNugget = fgPtr(sb.Nugget)
	}
	if sb.NuggetLabel != "" {
		t.StatusBarNuggetLabel = fgPtr(sb.NuggetLabel)
	}

	// App background
	if v.App.Background != "" {
		t.AppBackground = style.Highlight{Bg: parseColor(v.App.Background)}
	}

	// Custom app styles (wrapped as Highlight for foxful-cli resolution)
	t.Custom = v.App.toCustomStyles()

	return t
}

// fgPtr creates a Highlight with the given foreground color string.
func fgPtr(hex string) style.Highlight {
	return style.Highlight{Fg: parseColor(hex)}
}

// AppCustomStyles is the struct passed to Theme.Custom for foxful-cli reflection.
// All fields must be of type style.Highlight.
type AppCustomStyles struct {
	LyricActive          style.Highlight
	LyricTransition      style.Highlight
	LyricInactive        style.Highlight
	LyricHighlight       style.Highlight
	PlaybarMode          style.Highlight
	PlaybarVolume        style.Highlight
	PlaybarPlaying       style.Highlight
	PlaybarPaused        style.Highlight
	PlaybarHeartLiked    style.Highlight
	PlaybarHeartUnliked  style.Highlight
	PlaybarArtist        style.Highlight
	LoginError           style.Highlight
	LoginSuccess         style.Highlight
	AuthStatus           style.Highlight
	ConfigTableBorder     style.Highlight
	ConfigTableSelectedFg style.Highlight
	ConfigTableSelectedBg style.Highlight
}

// toCustomStyles converts AppColorConfig to an AppCustomStyles for Theme.Custom.
func (c AppColorConfig) toCustomStyles() AppCustomStyles {
	return AppCustomStyles{
		LyricActive:          fgHighlight(c.LyricActive, LyricActiveColor),
		LyricTransition:      fgHighlight(c.LyricTransition, LyricTransitionColor),
		LyricInactive:        fgHighlight(c.LyricInactive, LyricInactiveColor),
		LyricHighlight:       fgHighlight(c.LyricHighlight, LyricWhiteColor),
		PlaybarMode:          fgHighlight(c.PlaybarMode, PlaybarModeColor),
		PlaybarVolume:        fgHighlight(c.PlaybarVolume, PlaybarVolumeColor),
		PlaybarPlaying:       fgHighlight(c.PlaybarPlaying, PlaybarPlayingColor),
		PlaybarPaused:        fgHighlight(c.PlaybarPaused, PlaybarPausedColor),
		PlaybarHeartLiked:    fgHighlight(c.PlaybarHeartLiked, PlaybarHeartLikedColor),
		PlaybarHeartUnliked:  fgHighlight(c.PlaybarHeartUnliked, PlaybarHeartUnlikedColor),
		PlaybarArtist:        fgHighlight(c.PlaybarArtist, PlaybarArtistColor),
		LoginError:           fgHighlight(c.LoginError, LoginErrorColor),
		LoginSuccess:         fgHighlight(c.LoginSuccess, LoginSuccessColor),
		AuthStatus:           fgHighlight(c.AuthStatus, AuthStatusColor),
		ConfigTableBorder:     fgHighlight(c.ConfigTableBorder, ConfigTableBorderColor),
		ConfigTableSelectedFg: fgHighlight(c.ConfigTableSelectedFg, ConfigTableSelectedFgColor),
		ConfigTableSelectedBg: bgHighlight(c.ConfigTableSelectedBg, ConfigTableSelectedBgColor),
	}
}

// fgHighlight creates a Highlight with the given foreground color,
// falling back to defaultHex if empty or "random".
func fgHighlight(hex, defaultHex string) style.Highlight {
	c := parseAppColor(hex, defaultHex)
	return style.Highlight{Fg: c}
}

// bgHighlight creates a Highlight with the given background color,
// falling back to defaultHex if empty or "random".
func bgHighlight(hex, defaultHex string) style.Highlight {
	c := parseAppColor(hex, defaultHex)
	return style.Highlight{Bg: c}
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
	"Black":         lipgloss.Black,
	"Red":           lipgloss.Red,
	"Green":         lipgloss.Green,
	"Yellow":        lipgloss.Yellow,
	"Blue":          lipgloss.Blue,
	"Magenta":       lipgloss.Magenta,
	"Cyan":          lipgloss.Cyan,
	"White":         lipgloss.White,
	"BrightBlack":   lipgloss.BrightBlack,
	"BrightRed":     lipgloss.BrightRed,
	"BrightGreen":   lipgloss.BrightGreen,
	"BrightYellow":  lipgloss.BrightYellow,
	"BrightBlue":    lipgloss.BrightBlue,
	"BrightMagenta": lipgloss.BrightMagenta,
	"BrightCyan":    lipgloss.BrightCyan,
	"BrightWhite":   lipgloss.BrightWhite,
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

// ResolveSpectrumVerticalBg returns the spectrum vertical blend background color.
func (c AppColorConfig) ResolveSpectrumVerticalBg() string {
	if c.SpectrumVerticalBg == "" {
		return SpectrumVerticalBgDefault
	}
	return c.SpectrumVerticalBg
}

// GetSpectrumVerticalBg returns the resolved spectrum vertical blend background color
// for the current theme and brightness.
func GetSpectrumVerticalBg() string {
	isDark := style.HasDarkBackground()
	return CurrentThemeRegistry().CurrentAppColorConfig(isDark).ResolveSpectrumVerticalBg()
}

// ResolveContextMenuBg returns the context menu background color.
func (c AppColorConfig) ResolveContextMenuBg() string {
	if c.ContextMenuBg == "" {
		return ContextMenuBgDefault
	}
	return c.ContextMenuBg
}

// ResolveContextMenuBorder returns the context menu border color.
func (c AppColorConfig) ResolveContextMenuBorder() string {
	if c.ContextMenuBorder == "" {
		return ContextMenuBorderDefault
	}
	return c.ContextMenuBorder
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

	// Login / auth status colors (ANSI color approximations)
	LoginErrorColor   = "#FF0000" // BrightRed
	LoginSuccessColor = "#00FF00" // BrightGreen
	AuthStatusColor   = "#0000FF" // BrightBlue

	// Config table colors
	ConfigTableBorderColor     = "#585858" // ANSI 240
	ConfigTableSelectedFgColor = "#FFFF55" // ANSI 229
	ConfigTableSelectedBgColor = "#D70000" // ANSI 160

	// Context menu defaults
	ContextMenuBgDefault     = "#383838"
	ContextMenuBorderDefault = "#5C5C5C"

	// Spectrum vertical blend background
	SpectrumVerticalBgDefault = "#000000"

	// Desktop lyrics default background
	DesktopLyricsBgDefault = "#000000"
)

// ResolvedAppStyles mirrors AppCustomStyles with lipgloss.Style fields,
// matching what foxful-cli's StyleSet.Custom map contains after resolution.
// Use style.CustomStyles[ResolvedAppStyles](ss) to obtain an instance.
type ResolvedAppStyles struct {
	LyricActive          lipgloss.Style
	LyricTransition      lipgloss.Style
	LyricInactive        lipgloss.Style
	LyricHighlight       lipgloss.Style
	PlaybarMode          lipgloss.Style
	PlaybarVolume        lipgloss.Style
	PlaybarPlaying       lipgloss.Style
	PlaybarPaused        lipgloss.Style
	PlaybarHeartLiked    lipgloss.Style
	PlaybarHeartUnliked  lipgloss.Style
	PlaybarArtist        lipgloss.Style
	LoginError           lipgloss.Style
	LoginSuccess         lipgloss.Style
	AuthStatus           lipgloss.Style
	ConfigTableBorder     lipgloss.Style
	ConfigTableSelectedFg lipgloss.Style
	ConfigTableSelectedBg lipgloss.Style
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

// SafeGetBackground extracts the background color from a lipgloss.Style,
// falling back to the provided default color string (ANSI name or hex).
func SafeGetBackground(s lipgloss.Style, defaultStr string) color.Color {
	if c := s.GetBackground(); c != nil {
		return c
	}
	return parseColor(defaultStr)
}

// GetCurrentMarkdownTheme returns the configured markdown theme name for the current brightness.
func GetCurrentMarkdownTheme(darkBackground bool) string {
	registry := CurrentThemeRegistry()
	tf, ok := registry.ActiveThemeOrDefault("")
	if !ok {
		if darkBackground {
			return MarkdownDarkDefault
		}
		return MarkdownLightDefault
	}
	var cfg MarkdownConfig
	if darkBackground {
		cfg = tf.Dark.Markdown
	} else {
		cfg = tf.Light.Markdown
	}

	name := cfg.Dark
	if !darkBackground {
		name = cfg.Light
	}
	if name == "" {
		if darkBackground {
			return MarkdownDarkDefault
		}
		return MarkdownLightDefault
	}
	return name
}

// Markdown theme defaults
const (
	MarkdownDarkDefault  = "dark"
	MarkdownLightDefault = "light"
)
