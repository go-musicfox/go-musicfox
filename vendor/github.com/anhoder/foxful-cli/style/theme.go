// Package style provides centralized theme definitions and reusable lipgloss styles
// for the foxful-cli TUI framework.
//
// This package consolidates style definitions that were previously scattered across
// model/*.go and util/ui.go, making it easier to customize the visual appearance.
// The public types (Theme, StyleSet) are designed to be configurable by downstream
// consumers like go-musicfox.
package style

import (
	"image/color"
	"reflect"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// Highlight defines a complete text style for a UI element, following Neovim's
// :highlight model where foreground, background, and text attributes are bundled
// together in a single configuration.
//
// Zero/nil fields mean "not set" and will fall back to element-specific defaults
// during StyleSet construction. This avoids the anti-pattern of splitting a
// single component's style across multiple Theme fields like FooFg + FooBg.
//
// The Preset field allows referencing a named highlight preset (e.g. "normal",
// "bold", "dim") as a base. When set, the preset's values provide defaults that
// can be overridden by explicit Fg/Bg/Bold/Italic/Underline fields. Presets are
// looked up first in Theme.HighlightPresets (user-defined), then in the global
// BuiltinHighlightPresets map.
type Highlight struct {
	Preset    string      // Preset name. Empty = no preset applied.
	Fg        color.Color // Foreground color. nil = fall back to default.
	Bg        color.Color // Background color. nil = fall back to default.
	Bold      *bool       // Bold attribute. nil = not set (use element default).
	Italic    *bool       // Italic attribute. nil = not set.
	Underline *bool       // Underline attribute. nil = not set.
}

// PopupTheme groups every visual token owned by a popup surface.
// Surface is the sole background for non-interactive popup content.
type PopupTheme struct {
	Surface color.Color // nil → Theme.Surface → Theme.Background → adaptive default
	Border  color.Color // nil → Theme.Border → Theme.Accent

	// Title and Content apply only foreground and text attributes. Their
	// Highlight.Bg values are deliberately ignored so they cannot fragment the
	// popup surface.
	Title   Highlight
	Content Highlight

	// Action backgrounds are intentional interactive-state affordances.
	Action        Highlight
	ActionFocused Highlight
	ActionHover   Highlight

	// Context menu rows share the popup surface while exposing their own frame
	// and interactive states.
	ContextMenuBorder       color.Color // nil → #5C5C5C on dark backgrounds, #A8A8A8 on light
	ContextMenuItem         Highlight   // fg→terminal default, bg→Popup.Surface
	ContextMenuItemFocused  Highlight   // fg/bg→SelectedItem
	ContextMenuItemHover    Highlight   // bg→SelectedItem.Bg, underline→false
	ContextMenuItemDisabled Highlight   // fg→Theme.Muted, bg→Popup.Surface
	ContextMenuHeader       Highlight   // fg→Primary, bold→true, bg→Popup.Surface (分组标题行)
	ContextMenuSeparator    color.Color // nil → ContextMenuBorder
}

// PopupStyleSet is the resolved, render-ready popup visual surface.
type PopupStyleSet struct {
	Surface color.Color

	Frame                   lipgloss.Style
	Title                   lipgloss.Style
	Content                 lipgloss.Style
	Action                  lipgloss.Style
	ActionFocused           lipgloss.Style
	ActionHover             lipgloss.Style
	ScrollTrack             lipgloss.Style
	ScrollThumb             lipgloss.Style
	ContextMenuFrame        lipgloss.Style
	ContextMenuItem         lipgloss.Style
	ContextMenuItemFocused  lipgloss.Style
	ContextMenuItemHover    lipgloss.Style
	ContextMenuItemDisabled lipgloss.Style
	ContextMenuHeader       lipgloss.Style
	ContextMenuSeparator    lipgloss.Style
}

// NotificationTheme groups the visual tokens for the notification system.
// Nil color fields fall back to semantic Theme colors.
type NotificationTheme struct {
	Surface color.Color // nil → Theme.Surface → Theme.Background → adaptive default

	// Per-level border colors. Nil falls back to the matching semantic color.
	InfoBorder    color.Color // nil → Theme.Info
	SuccessBorder color.Color // nil → Theme.Success
	WarningBorder color.Color // nil → Theme.Warning
	ErrorBorder   color.Color // nil → Theme.Error

	// Title and Message apply only foreground and text attributes; their
	// backgrounds are forced to Surface so they cannot fragment the surface.
	// Message foreground is unset by default and inherits the terminal color.
	Title   Highlight
	Message Highlight

	// Action backgrounds are intentional interactive-state affordances.
	Action      Highlight
	ActionHover Highlight

	// Per-level icon prefixes (optional; falls back to defaults).
	InfoIcon    string
	SuccessIcon string
	WarningIcon string
	ErrorIcon   string
}

// NotificationStyleSet is the resolved, render-ready notification visual surface.
type NotificationStyleSet struct {
	Surface color.Color

	InfoFrame    lipgloss.Style
	SuccessFrame lipgloss.Style
	WarningFrame lipgloss.Style
	ErrorFrame   lipgloss.Style

	Title       lipgloss.Style
	Message     lipgloss.Style
	Close       lipgloss.Style
	Action      lipgloss.Style
	ActionHover lipgloss.Style

	// Level icon prefixes (Unicode).
	InfoIcon    string
	SuccessIcon string
	WarningIcon string
	ErrorIcon   string
}

// BuiltinHighlightPresets defines named highlight presets that can be referenced
// via Highlight.Preset. Users can override or extend these via Theme.HighlightPresets.
// Preset values act as defaults — explicit fields on the Highlight take precedence.
var BuiltinHighlightPresets = map[string]Highlight{
	"normal":    {},                         // No overrides; uses element defaults
	"bold":      {Bold: BoolPtr(true)},      // Bold text
	"italic":    {Italic: BoolPtr(true)},    // Italic text
	"underline": {Underline: BoolPtr(true)}, // Underlined text
}

// Theme defines the color palette and visual tokens for the application.
// All fields use the standard image/color.Color interface for lipgloss v2 compatibility.
// Any field left as nil falls back to a sensible default (documented per field).
type Theme struct {
	// ---- Base palette (semantic colors used as fallbacks for Highlights) ----

	Primary                color.Color // Main accent: selections, highlights, active elements
	Secondary              color.Color // Subtitles, hints, inactive elements
	Accent                 color.Color // Highlighted borders, focus indicators, decorative accents
	Success                color.Color // Positive/success indicators (green)
	Warning                color.Color // Caution/warning indicators (yellow)
	Error                  color.Color // Error/negative indicators (red)
	Info                   color.Color // Informational messages (blue/cyan)
	Muted                  color.Color // Grayed-out/disabled text
	HintKey                color.Color // Keyboard shortcut labels in help hint bar
	StatusBarBreadcrumbSep color.Color // Separator color in breadcrumb trails ("/" or ">")

	// Terminal-level colors
	Background color.Color // Terminal background (for reverse/contrast computation)
	Foreground color.Color // Terminal foreground (default text)
	Border     color.Color // Default subtle border color. Falls back to Accent.
	Surface    color.Color // Background for elevated elements (cards, panels, popups)

	// ---- Component highlights (Neovim-style: fg, bg, attributes together) ----
	// Each field bundles the complete visual style for a UI element.
	// Nil Fg/Bg fall back to the defaults documented below.
	// Nil Bold/Italic/Underline means "use the element's default" (usually false).

	Title                Highlight         // fg→Primary, bold→true
	MenuTitle            Highlight         // fg→Primary
	MenuItem             Highlight         // fg→nil (terminal default), bg→transparent
	SelectedItem         Highlight         // fg→Primary, bg→computed from Primary blend
	Subtitle             Highlight         // fg→Secondary
	Prompt               Highlight         // fg→Primary
	BackButton           Highlight         // fg→Secondary, bold→true
	Button               Highlight         // fg→Primary
	ButtonBlurred        Highlight         // fg→Secondary
	ProgressEmpty        Highlight         // fg→Secondary
	ScrollTrack          Highlight         // fg→Secondary, scrollbar track (gutter lines)
	ScrollThumb          Highlight         // fg→Secondary, scrollbar thumb (position indicator)
	Popup                PopupTheme        // popup-owned surface and interactive states
	Notification         NotificationTheme // notification-owned surface, borders, icons
	StatusBar            Highlight         // fg→Secondary, bg→transparent
	StatusBarText        Highlight         // fg→Secondary
	StatusBarBreadcrumb  Highlight         // fg→Muted, bg→computed (falls back to StatusBarTime.Bg)
	StatusBarTime        Highlight         // fg→Muted, bg→computed from Surface blend
	StatusBarNugget      Highlight         // fg→Foreground (or white on dark), bg→transparent
	StatusBarNuggetLabel Highlight         // fg→same as StatusBarNugget.Fg, bg→Primary
	AppBackground        Highlight         // Bg→transparent (terminal shows through)
	SelectedItemBg       Highlight         // Bg→transparent (falls back to AppBackground.Bg)

	// ---- Hover/click highlights (interactive states) ----
	// Each field below controls the visual feedback when the mouse hovers or
	// clicks on an interactive element. Nil fields fall back to the documented
	// defaults, which are derived from the element's normal style + visual cues.
	MenuItemHover            Highlight // fg→SelectedItem.Fg, underline→true
	SelectedItemHover        Highlight // same as SelectedItem, underline→true
	StatusBarBreadcrumbHover Highlight // fg→computed (lighten/darken), bg→computed, bold→true, underline→true
	StatusBarBreadcrumbClick Highlight // fg→computed (stronger), bg→computed, bold→true, underline→true
	BackButtonHover          Highlight // fg→BackButton.Fg, bold→true

	// HighlightPresets allows users to define named highlight presets that can
	// be referenced via Highlight.Preset on any Highlight field. User-defined
	// presets take precedence over BuiltinHighlightPresets.
	//
	// Example:
	//   theme.HighlightPresets = map[string]Highlight{
	//     "accent": {Fg: lipgloss.Color("#FF5F87"), Bold: BoolPtr(true)},
	//   }
	//   theme.StatusBar.Preset = "accent"  // StatusBar inherits accent preset
	HighlightPresets map[string]Highlight

	// Custom holds any user-defined struct with Highlight fields for
	// application-specific styles. The struct's Highlight fields are
	// resolved through the full pipeline (presets → highlight → lipgloss.Style)
	// during StyleSet construction.
	Custom any
}

// BoolPtr returns a pointer to the given bool value.
// Used for setting Bold/Italic/Underline on Highlight structs.
func BoolPtr(b bool) *bool { return &b }

// orString returns a if it is non-empty, otherwise b.
func orString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ---- terminal background detection (runtime-updatable) ----

// detectedDarkBg caches whether the terminal has a dark background.
// Defaults to true (dark backgrounds are common in terminals).
// Updated on startup by App.Run() via SetDarkBackground, and at runtime via BackgroundColorMsg.
var detectedDarkBg = true

// SetDarkBackground updates the cached terminal background detection result.
// Called on startup from App.Run() to seed the initial value, and on every
// tea.BackgroundColorMsg to keep the theme in sync with system light/dark
// mode changes at runtime.
func SetDarkBackground(dark bool) {
	detectedDarkBg = dark
}

// ---- built-in theme presets ----

// DefaultDarkTheme returns the standard foxful-cli dark theme.
// This is the original DefaultTheme, suitable for dark-background terminals.
// All detail colors are nil (unset), so they fall back to the base palette colors.
func DefaultDarkTheme() Theme {
	return Theme{
		Primary:                lipgloss.BrightGreen,
		Secondary:              lipgloss.BrightBlack,
		Accent:                 lipgloss.BrightBlue,
		Success:                lipgloss.BrightGreen,
		Warning:                lipgloss.BrightYellow,
		Error:                  lipgloss.BrightRed,
		Info:                   lipgloss.BrightCyan,
		Muted:                  lipgloss.BrightBlack,
		HintKey:                lipgloss.Color("#6E6E6E"),
		StatusBarBreadcrumbSep: lipgloss.Color("#757575"),
		Background:             lipgloss.Color("#1A1A1A"),
		Foreground:             lipgloss.Color("#FFFFFF"),
		Border:                 lipgloss.Color("#333333"),
		Surface:                lipgloss.Color("#242424"),

		StatusBarBreadcrumb: Highlight{},
	}
}

// DefaultLightTheme returns the standard foxful-cli light theme.
// Designed for light-background terminals with a Material-inspired palette.
// All detail colors are nil (unset), so they fall back to the base palette colors.
func DefaultLightTheme() Theme {
	return Theme{
		Primary:                lipgloss.Color("#2E7D32"), // Green 700
		Secondary:              lipgloss.Color("#757575"), // Gray 600
		Accent:                 lipgloss.Color("#1565C0"), // Blue 800
		Success:                lipgloss.Color("#388E3C"), // Green 700
		Warning:                lipgloss.Color("#E65100"), // Orange 900
		Error:                  lipgloss.Color("#C62828"), // Red 800
		Info:                   lipgloss.Color("#0277BD"), // Light Blue 800
		Muted:                  lipgloss.Color("#9E9E9E"), // Gray 500
		HintKey:                lipgloss.Color("#BDBDBD"), // Gray 400
		StatusBarBreadcrumbSep: lipgloss.Color("#BDBDBD"), // Gray 400
		Background:             lipgloss.Color("#FFFFFF"),
		Foreground:             lipgloss.Color("#212121"), // Gray 900
		Border:                 lipgloss.Color("#E0E0E0"), // Gray 300
		Surface:                lipgloss.Color("#F5F5F5"), // Gray 100

		StatusBarBreadcrumb: Highlight{},
	}
}

// DefaultTheme returns an adaptive default theme based on terminal background
// detection. It automatically selects between the dark and light default themes
// by querying the terminal's color scheme via OSC escape sequences.
//
// This is the recommended theme for most applications. If you want to explicitly
// use a dark or light theme regardless of terminal settings, use DefaultDarkTheme()
// or DefaultLightTheme() instead.
func DefaultTheme() Theme {
	if detectedDarkBg {
		return DefaultDarkTheme()
	}
	return DefaultLightTheme()
}

// DarkTheme returns a dark-background theme preset (alias for DefaultDarkTheme).
func DarkTheme() Theme {
	return DefaultDarkTheme()
}

// LightTheme returns a light-background theme preset (alias for DefaultLightTheme).
func LightTheme() Theme {
	return DefaultLightTheme()
}

// ---- preset themes inspired by popular design systems ----

// GitHubDarkTheme returns a theme inspired by GitHub's dark mode.
// Background: #0d1117, Surface: #161b22, Accent: #58a6ff, Border: #30363d.
func GitHubDarkTheme() Theme {
	return Theme{
		Primary:                lipgloss.Color("#58A6FF"),
		Secondary:              lipgloss.Color("#8B949E"),
		Accent:                 lipgloss.Color("#58A6FF"),
		Success:                lipgloss.Color("#3FB950"),
		Warning:                lipgloss.Color("#D29922"),
		Error:                  lipgloss.Color("#F85149"),
		Info:                   lipgloss.Color("#79C0FF"),
		Muted:                  lipgloss.Color("#484F58"),
		HintKey:                lipgloss.Color("#6E7681"),
		StatusBarBreadcrumbSep: lipgloss.Color("#484F58"),
		Background:             lipgloss.Color("#0D1117"),
		Foreground:             lipgloss.Color("#C9D1D9"),
		Border:                 lipgloss.Color("#30363D"),
		Surface:                lipgloss.Color("#161B22"),
	}
}

// VSCodeDarkTheme returns a theme inspired by VS Code's Dark+ default.
// Background: #1e1e1e, Surface: #252526, Accent: #007acc, Border: #454545.
func VSCodeDarkTheme() Theme {
	return Theme{
		Primary:                lipgloss.Color("#007ACC"),
		Secondary:              lipgloss.Color("#858585"),
		Accent:                 lipgloss.Color("#007ACC"),
		Success:                lipgloss.Color("#6A9955"),
		Warning:                lipgloss.Color("#CE9178"),
		Error:                  lipgloss.Color("#F44747"),
		Info:                   lipgloss.Color("#9CDCFE"),
		Muted:                  lipgloss.Color("#585858"),
		HintKey:                lipgloss.Color("#7A7A7A"),
		StatusBarBreadcrumbSep: lipgloss.Color("#585858"),
		Background:             lipgloss.Color("#1E1E1E"),
		Foreground:             lipgloss.Color("#D4D4D4"),
		Border:                 lipgloss.Color("#454545"),
		Surface:                lipgloss.Color("#252526"),
	}
}

// LinearDarkTheme returns a theme inspired by Linear's design system.
// Background: #1a1a1a, Surface: #222222, Accent: #5e6ad2, Border: #2a2a2a.
func LinearDarkTheme() Theme {
	return Theme{
		Primary:                lipgloss.Color("#5E6AD2"),
		Secondary:              lipgloss.Color("#6E6E6E"),
		Accent:                 lipgloss.Color("#5E6AD2"),
		Success:                lipgloss.Color("#4CB782"),
		Warning:                lipgloss.Color("#F2A33F"),
		Error:                  lipgloss.Color("#E5484D"),
		Info:                   lipgloss.Color("#7B80E8"),
		Muted:                  lipgloss.Color("#4A4A4A"),
		HintKey:                lipgloss.Color("#6E6E6E"),
		StatusBarBreadcrumbSep: lipgloss.Color("#4A4A4A"),
		Background:             lipgloss.Color("#1A1A1A"),
		Foreground:             lipgloss.Color("#E6E6E6"),
		Border:                 lipgloss.Color("#2A2A2A"),
		Surface:                lipgloss.Color("#222222"),
	}
}

// HasDarkBackground detects whether the terminal has a dark background.
// Returns true for dark background, false for light background.
// Defaults to true (dark) if detection fails.
//
// This is a convenience wrapper around lipgloss.HasDarkBackground.
// The result is cached after the first call for efficiency.
func HasDarkBackground() bool {
	return detectedDarkBg
}

// AdaptiveTheme returns a function that picks the light or dark theme based
// on the terminal background. Usage:
//
//	isDark := style.HasDarkBackground()
//	theme := style.AdaptiveTheme(style.DefaultLightTheme(), style.DefaultDarkTheme())(isDark)
//
// Note: DefaultTheme() now handles auto-detection internally. This function
// is useful when you want to use custom light/dark theme pairs.
func AdaptiveTheme(light, dark Theme) func(isDark bool) Theme {
	return func(isDark bool) Theme {
		if isDark {
			return dark
		}
		return light
	}
}

// StyleSet holds pre-built lipgloss styles for common UI elements.
// These are designed to be composed and extended by callers.
type StyleSet struct {
	// Title is the base style for the application title bar.
	Title lipgloss.Style

	// MenuTitle is the style for the current menu title line.
	MenuTitle lipgloss.Style

	// MenuItem is the default style for unselected menu items.
	MenuItem lipgloss.Style

	// MenuItemHover is the style for unselected menu items on mouse hover.
	// Uses underline and primary color to indicate clickability.
	MenuItemHover lipgloss.Style

	// SelectedItem is the style for the currently selected menu item.
	SelectedItem lipgloss.Style

	// SelectedItemHover is the style for the selected menu item on mouse hover.
	// Adds underline on top of the selected style.
	SelectedItemHover lipgloss.Style

	// Subtitle is the style for menu item subtitles.
	Subtitle lipgloss.Style

	// Prompt is the style for the focused input prompt (e.g., "> ").
	Prompt lipgloss.Style

	// Button is the style for focused buttons.
	Button lipgloss.Style

	// ButtonBlurred is the style for blurred/inactive buttons.
	ButtonBlurred lipgloss.Style

	// ProgressEmpty is the style for the unfilled portion of progress bars.
	ProgressEmpty lipgloss.Style

	// Border is the style for decorative borders.
	Border lipgloss.Style

	// Popup is the complete resolved visual surface for popup dialogs.
	Popup PopupStyleSet

	// Notification is the complete resolved visual surface for notifications.
	Notification NotificationStyleSet

	// Success is the style for success/positive messages.
	Success lipgloss.Style

	// Warning is the style for warning/caution messages.
	Warning lipgloss.Style

	// Error is the style for error/negative messages.
	Error lipgloss.Style

	// Info is the style for informational messages.
	Info lipgloss.Style

	// Muted is the style for muted/disabled text.
	Muted lipgloss.Style

	// HintKey is the style for keyboard shortcut key labels in the help hint bar.
	// Uses Theme.HintKey — intentionally subtle.
	HintKey lipgloss.Style

	// StatusBar is the base style for the bottom status bar (full-width background).
	StatusBar lipgloss.Style

	// StatusBarText is the text style inside the status bar.
	StatusBarText lipgloss.Style

	// StatusBarBreadcrumb is the breadcrumb text style inside the status bar.
	StatusBarBreadcrumb lipgloss.Style

	// StatusBarBreadcrumbBg is the background style for the breadcrumb path area in the status bar.
	StatusBarBreadcrumbBg lipgloss.Style

	// StatusBarNugget is the base nugget style for status bar blocks (colored bg, white text, padding).
	StatusBarNugget lipgloss.Style

	// StatusBarNuggetLabel is a nugget for the breadcrumb label in the status bar.
	StatusBarNuggetLabel lipgloss.Style

	// StatusBarTime is a nugget for the time display in the status bar.
	StatusBarTime lipgloss.Style

	// StatusBarBreadcrumbHover is the style for breadcrumb ancestor segments on mouse hover.
	// Derived from Breadcrumb with underline and bold.
	StatusBarBreadcrumbHover lipgloss.Style

	// StatusBarBreadcrumbClick is the style for breadcrumb ancestor segments on mouse press.
	// Derived from Breadcrumb with stronger visual feedback.
	StatusBarBreadcrumbClick lipgloss.Style

	// StatusBarBreadcrumbSep is the style for the breadcrumb separator (">").
	StatusBarBreadcrumbSep lipgloss.Style

	// BackButton is the style for the back/return icon button shown before the
	// menu title when inside a submenu. Uses primary color to indicate clickability.
	BackButton lipgloss.Style

	// BackButtonHover is the style for the back button on mouse hover.
	// Adds visual feedback (highlight background, bold).
	BackButtonHover lipgloss.Style

	// AppBackground is the base style for the overall application window.
	// Defaults to transparent (terminal background shows through).
	AppBackground lipgloss.Style

	// Normal is the base style for unstyled normal text.
	// Just applies the normal Text color without any extra formatting.
	Normal lipgloss.Style

	// Custom holds application-specific styles built from Theme.Custom.
	// Internally stored as map[string]lipgloss.Style keyed by struct field name.
	// Use CustomStyles[T]() to get typed flat access via a user-defined struct
	// with lipgloss.Style fields:
	//   custom := style.CustomStyles[AppCustomStyles](style.CurrentStyleSet())
	//   playerColor := custom.PlayerColor
	Custom any

	theme Theme
}

// BlankStyle returns a lipgloss.Style for rendering blank/whitespace content
// that inherits the app background color. Priority:
//   - component's own background (via Inherit if caller passes parent)
//   - AppBackground color (from theme)
//   - transparent (no background set)
func (ss StyleSet) BlankStyle() lipgloss.Style {
	return ss.AppBackground
}

// NewStyleSet creates a pre-configured set of styles from a Theme.
// Use this as the base and customize individual styles as needed.
func NewStyleSet(theme Theme) StyleSet {
	base := StyleSet{theme: theme}

	or := func(v, d color.Color) color.Color {
		if v != nil {
			return v
		}
		return d
	}

	// Helper: apply Highlight to a lipgloss style
	applyHL := func(s lipgloss.Style, hl Highlight) lipgloss.Style {
		if hl.Fg != nil {
			s = s.Foreground(hl.Fg)
		}
		if hl.Bg != nil {
			s = s.Background(hl.Bg)
		}
		if hl.Bold != nil {
			s = s.Bold(*hl.Bold)
		}
		if hl.Italic != nil {
			s = s.Italic(*hl.Italic)
		}
		if hl.Underline != nil {
			s = s.Underline(*hl.Underline)
		}
		return s
	}

	// Popup title/body text must never be able to override PopupTheme.Surface.
	applyPopupText := func(s lipgloss.Style, hl Highlight, surface color.Color) lipgloss.Style {
		hl.Bg = nil
		return applyHL(s, hl).Background(surface)
	}

	// ---- Resolve Highlight defaults ----

	// resolvePreset merges a named preset (if set) into the Highlight.
	// User-defined presets in theme.HighlightPresets take precedence over
	// BuiltinHighlightPresets. Explicit fields on hl always override preset values.
	resolvePreset := func(hl Highlight) Highlight {
		if hl.Preset == "" {
			return hl
		}
		var preset Highlight
		var ok bool
		if theme.HighlightPresets != nil {
			preset, ok = theme.HighlightPresets[hl.Preset]
		}
		if !ok {
			preset, ok = BuiltinHighlightPresets[hl.Preset]
		}
		if !ok {
			return hl // Unknown preset, keep as-is
		}
		// Merge: preset provides defaults, hl explicit fields override
		if hl.Fg == nil {
			hl.Fg = preset.Fg
		}
		if hl.Bg == nil {
			hl.Bg = preset.Bg
		}
		if hl.Bold == nil {
			hl.Bold = preset.Bold
		}
		if hl.Italic == nil {
			hl.Italic = preset.Italic
		}
		if hl.Underline == nil {
			hl.Underline = preset.Underline
		}
		// Clear preset name after resolution to avoid re-resolving
		hl.Preset = ""
		return hl
	}

	// Helper to resolve a Highlight: first apply presets, then fill fg/bg fallbacks.
	resolveHL := func(hl Highlight, defaultFg, defaultBg color.Color) Highlight {
		out := resolvePreset(hl)
		if out.Fg == nil && defaultFg != nil {
			out.Fg = defaultFg
		}
		if out.Bg == nil && defaultBg != nil {
			out.Bg = defaultBg
		}
		return out
	}

	// Pre-resolve all highlights to apply presets. This ensures code paths that
	// access highlight fields directly (e.g., theme.StatusBar.Fg) also pick up
	// any preset values. resolveHL calls will also apply presets (idempotently
	// since Preset is cleared after first resolution).
	theme.Title = resolvePreset(theme.Title)
	theme.MenuTitle = resolvePreset(theme.MenuTitle)
	theme.MenuItem = resolvePreset(theme.MenuItem)
	theme.SelectedItem = resolvePreset(theme.SelectedItem)
	theme.Subtitle = resolvePreset(theme.Subtitle)
	theme.Prompt = resolvePreset(theme.Prompt)
	theme.BackButton = resolvePreset(theme.BackButton)
	theme.Button = resolvePreset(theme.Button)
	theme.ButtonBlurred = resolvePreset(theme.ButtonBlurred)
	theme.ProgressEmpty = resolvePreset(theme.ProgressEmpty)
	theme.Popup.Title = resolvePreset(theme.Popup.Title)
	theme.Popup.Content = resolvePreset(theme.Popup.Content)
	theme.Popup.Action = resolvePreset(theme.Popup.Action)
	theme.Popup.ActionFocused = resolvePreset(theme.Popup.ActionFocused)
	theme.Popup.ActionHover = resolvePreset(theme.Popup.ActionHover)
	theme.Popup.ContextMenuItem = resolvePreset(theme.Popup.ContextMenuItem)
	theme.Popup.ContextMenuItemFocused = resolvePreset(theme.Popup.ContextMenuItemFocused)
	theme.Popup.ContextMenuItemHover = resolvePreset(theme.Popup.ContextMenuItemHover)
	theme.Popup.ContextMenuItemDisabled = resolvePreset(theme.Popup.ContextMenuItemDisabled)
	theme.Popup.ContextMenuHeader = resolvePreset(theme.Popup.ContextMenuHeader)
	theme.Notification.Title = resolvePreset(theme.Notification.Title)
	theme.Notification.Message = resolvePreset(theme.Notification.Message)
	theme.Notification.Action = resolvePreset(theme.Notification.Action)
	theme.Notification.ActionHover = resolvePreset(theme.Notification.ActionHover)
	theme.StatusBar = resolvePreset(theme.StatusBar)
	theme.StatusBarText = resolvePreset(theme.StatusBarText)
	theme.StatusBarBreadcrumb = resolvePreset(theme.StatusBarBreadcrumb)
	theme.StatusBarTime = resolvePreset(theme.StatusBarTime)
	theme.StatusBarNugget = resolvePreset(theme.StatusBarNugget)
	theme.StatusBarNuggetLabel = resolvePreset(theme.StatusBarNuggetLabel)
	theme.AppBackground = resolvePreset(theme.AppBackground)
	theme.SelectedItemBg = resolvePreset(theme.SelectedItemBg)
	theme.MenuItemHover = resolvePreset(theme.MenuItemHover)
	theme.SelectedItemHover = resolvePreset(theme.SelectedItemHover)
	theme.StatusBarBreadcrumbHover = resolvePreset(theme.StatusBarBreadcrumbHover)
	theme.StatusBarBreadcrumbClick = resolvePreset(theme.StatusBarBreadcrumbClick)
	theme.BackButtonHover = resolvePreset(theme.BackButtonHover)

	// Transparent sentinel
	var noColor color.Color = lipgloss.NoColor{}

	// Detect dark/light background for computed values
	bgIsDark := false
	if theme.Background != nil && theme.Background != noColor {
		if bg, ok := colorful.MakeColor(theme.Background); ok {
			_, _, l := bg.Hsl()
			bgIsDark = l < 0.5
		}
	}

	// ---- Base palette ----
	titleHL := resolveHL(theme.Title, theme.Primary, nil)
	if titleHL.Bold == nil {
		titleHL.Bold = BoolPtr(true)
	}

	menuTitleHL := resolveHL(theme.MenuTitle, theme.Primary, nil)
	subtitleHL := resolveHL(theme.Subtitle, theme.Secondary, nil)
	promptHL := resolveHL(theme.Prompt, theme.Primary, nil)
	backButtonHL := resolveHL(theme.BackButton, theme.Secondary, nil)
	if backButtonHL.Bold == nil {
		backButtonHL.Bold = BoolPtr(true)
	}

	// Selected item: compute a low-contrast bg from Primary if not set.
	selectedItemHL := resolveHL(theme.SelectedItem, theme.Primary, nil)
	if selectedItemHL.Bg == nil {
		if primary, ok := colorful.MakeColor(theme.Primary); ok {
			if theme.Background != nil && theme.Background != noColor {
				if bg, ok := colorful.MakeColor(theme.Background); ok {
					_, _, l := bg.Hsl()
					if l > 0.5 {
						selectedItemHL.Bg = primary.BlendLab(bg, 0.94).Clamped()
					} else {
						highlighted := primary.BlendLab(colorful.Color{R: 1, G: 1, B: 1}, 0.8)
						selectedItemHL.Bg = highlighted.BlendLab(bg, 0.78).Clamped()
					}
				}
			}
		}
		if selectedItemHL.Bg == nil {
			selectedItemHL.Bg = noColor
		}
	}

	// Breadcrumb background: use StatusBarBreadcrumb.Bg → computed from Surface → fallback
	breadcrumbBg := theme.StatusBarBreadcrumb.Bg
	if breadcrumbBg == nil {
		if theme.Surface != nil && theme.Surface != noColor {
			if srf, ok := colorful.MakeColor(theme.Surface); ok {
				if bgIsDark {
					breadcrumbBg = srf.BlendLab(colorful.Color{R: 1, G: 1, B: 1}, 0.3).Clamped()
				} else {
					breadcrumbBg = srf.BlendLab(colorful.Color{}, 0.1).Clamped()
				}
			}
		}
		if breadcrumbBg == nil {
			breadcrumbBg = theme.AppBackground.Bg
		}
	}

	// Buttons / progress
	buttonHL := resolveHL(theme.Button, theme.Primary, nil)
	buttonBlurredHL := resolveHL(theme.ButtonBlurred, theme.Secondary, nil)
	progressEmptyHL := resolveHL(theme.ProgressEmpty, theme.Secondary, nil)

	// Popup
	popupSurface := or(theme.Popup.Surface, theme.Surface)
	if popupSurface == nil {
		popupSurface = theme.Background
	}
	if popupSurface == nil {
		popupSurface = theme.AppBackground.Bg
	}
	borderColor := or(theme.Border, theme.Accent)
	popupBorder := or(theme.Popup.Border, borderColor)
	popupTitleHL := resolveHL(theme.Popup.Title, theme.Primary, nil)
	if popupTitleHL.Bold == nil {
		popupTitleHL.Bold = BoolPtr(true)
	}
	popupContentHL := resolveHL(theme.Popup.Content, theme.Foreground, nil)
	popupActionHL := resolveHL(theme.Popup.Action, theme.Primary, theme.Muted)
	popupActionFocusedHL := resolveHL(theme.Popup.ActionFocused, lipgloss.Color("#FFFFFF"), theme.Primary)
	contextMenuBorderColor := color.Color(lipgloss.Color("#A8A8A8"))
	if bgIsDark {
		contextMenuBorderColor = lipgloss.Color("#5C5C5C")
	}
	contextMenuBorderColor = or(theme.Popup.ContextMenuBorder, contextMenuBorderColor)
	contextMenuItemHL := resolveHL(theme.Popup.ContextMenuItem, nil, popupSurface)
	contextMenuItemFocusedHL := resolveHL(theme.Popup.ContextMenuItemFocused, selectedItemHL.Fg, selectedItemHL.Bg)
	contextMenuItemHoverHL := resolveHL(theme.Popup.ContextMenuItemHover, contextMenuItemHL.Fg, selectedItemHL.Bg)
	if theme.Popup.ContextMenuItemHover.Fg == nil {
		if bg := theme.AppBackground.Bg; bg == nil {
			contextMenuItemHoverHL.Fg = theme.Primary
		} else if _, transparent := bg.(lipgloss.NoColor); transparent {
			contextMenuItemHoverHL.Fg = theme.Primary
		}
	}
	if contextMenuItemHoverHL.Underline == nil {
		contextMenuItemHoverHL.Underline = BoolPtr(false)
	}
	contextMenuItemDisabledHL := resolveHL(theme.Popup.ContextMenuItemDisabled, theme.Muted, popupSurface)
	contextMenuHeaderHL := resolveHL(theme.Popup.ContextMenuHeader, theme.Primary, popupSurface)
	if contextMenuHeaderHL.Bold == nil {
		contextMenuHeaderHL.Bold = BoolPtr(true)
	}
	contextMenuSeparatorColor := or(theme.Popup.ContextMenuSeparator, contextMenuBorderColor)

	// Scrollbar: global configurable elements for consistent look across components.
	scrollTrackHL := resolveHL(theme.ScrollTrack, theme.Secondary, nil)
	scrollThumbHL := resolveHL(theme.ScrollThumb, theme.Secondary, nil)

	// Status bar
	// Fall back to the app background (then transparent) so the filler cells
	// between the breadcrumb and time never reveal content drawn beneath the TUI.
	statusBarFg := or(theme.StatusBar.Fg, theme.Secondary)
	statusBarBg := or(theme.StatusBar.Bg, or(theme.AppBackground.Bg, noColor))
	statusBarTextFg := or(theme.StatusBarText.Fg, theme.Secondary)

	// Status bar time bg
	statusBarTimeFg := or(theme.StatusBarTime.Fg, nil)
	statusBarTimeBg := theme.StatusBarTime.Bg
	if statusBarTimeBg == nil {
		// Use the same computed value as breadcrumbBg for consistency
		statusBarTimeBg = breadcrumbBg
	}

	// Status bar nuggets
	statusBarNuggetFg := or(theme.StatusBarNugget.Fg, theme.Foreground)
	statusBarNuggetLabelFg := or(theme.StatusBarNuggetLabel.Fg, statusBarNuggetFg)
	statusBarNuggetLabelBg := or(theme.StatusBarNuggetLabel.Bg, theme.Primary)

	// Breadcrumb separator
	breadcrumbSepColor := theme.StatusBarBreadcrumbSep
	if breadcrumbSepColor == nil {
		breadcrumbSepColor = theme.Muted
	}

	// App / Menu backgrounds
	appBg := or(theme.AppBackground.Bg, noColor)
	selectedItemBg := or(theme.SelectedItemBg.Bg, appBg)

	// ---- Hover highlights (interactive states) ----

	// MenuItemHover: fg defaults to selectedItemHL.Fg, underline adds clickability cue
	menuItemHoverHL := resolveHL(theme.MenuItemHover, selectedItemHL.Fg, nil)
	if menuItemHoverHL.Bg == nil {
		menuItemHoverHL.Bg = selectedItemBg
	}
	if menuItemHoverHL.Underline == nil {
		menuItemHoverHL.Underline = BoolPtr(false)
	}

	// SelectedItemHover: same as SelectedItem, underline true by default
	selectedItemHoverHL := resolveHL(theme.SelectedItemHover, selectedItemHL.Fg, selectedItemHL.Bg)
	if selectedItemHoverHL.Underline == nil {
		selectedItemHoverHL.Underline = BoolPtr(false)
	}

	// StatusBarBreadcrumbHover: fg computed based on dark/light, bg→breadcrumbBg
	breadcrumbHoverHL := resolveHL(theme.StatusBarBreadcrumbHover, nil, breadcrumbBg)
	if breadcrumbHoverHL.Fg == nil {
		if bgIsDark {
			breadcrumbHoverHL.Fg = lipgloss.Color("#E0E0E0")
		} else {
			breadcrumbHoverHL.Fg = lipgloss.Color("#424242")
		}
	}
	if breadcrumbHoverHL.Bold == nil {
		breadcrumbHoverHL.Bold = BoolPtr(true)
	}
	if breadcrumbHoverHL.Underline == nil {
		breadcrumbHoverHL.Underline = BoolPtr(true)
	}

	// StatusBarBreadcrumbClick: fg computed (stronger), bg→breadcrumbBg
	breadcrumbClickHL := resolveHL(theme.StatusBarBreadcrumbClick, nil, breadcrumbBg)
	if breadcrumbClickHL.Fg == nil {
		if bgIsDark {
			breadcrumbClickHL.Fg = lipgloss.Color("#FFFFFF")
		} else {
			breadcrumbClickHL.Fg = lipgloss.Color("#000000")
		}
	}
	if breadcrumbClickHL.Bold == nil {
		breadcrumbClickHL.Bold = BoolPtr(true)
	}
	if breadcrumbClickHL.Underline == nil {
		breadcrumbClickHL.Underline = BoolPtr(true)
	}

	// BackButtonHover: fg defaults to backButtonHL.Fg, bold true
	backButtonHoverHL := resolveHL(theme.BackButtonHover, backButtonHL.Fg, nil)
	if backButtonHoverHL.Bold == nil {
		backButtonHoverHL.Bold = BoolPtr(true)
	}

	// Popup action hover: action fg/bg plus an underline cue by default.
	popupActionHoverHL := resolveHL(theme.Popup.ActionHover, popupActionHL.Fg, popupActionHL.Bg)
	if popupActionHoverHL.Underline == nil {
		popupActionHoverHL.Underline = BoolPtr(true)
	}

	// ---- Build StyleSet ----

	base.Title = applyHL(lipgloss.NewStyle(), titleHL)

	base.MenuTitle = applyHL(lipgloss.NewStyle().Background(appBg), Highlight{Fg: menuTitleHL.Fg})

	base.MenuItem = lipgloss.NewStyle().Background(selectedItemBg)

	base.MenuItemHover = applyHL(lipgloss.NewStyle(), menuItemHoverHL)

	base.SelectedItem = lipgloss.NewStyle().
		Foreground(selectedItemHL.Fg).
		Background(selectedItemHL.Bg)

	base.SelectedItemHover = applyHL(base.SelectedItem, selectedItemHoverHL)

	base.Subtitle = applyHL(lipgloss.NewStyle().Background(appBg), Highlight{Fg: subtitleHL.Fg})

	base.Prompt = applyHL(lipgloss.NewStyle(), Highlight{Fg: promptHL.Fg})

	base.Button = applyHL(lipgloss.NewStyle(), Highlight{Fg: buttonHL.Fg})

	base.ButtonBlurred = applyHL(lipgloss.NewStyle(), Highlight{Fg: buttonBlurredHL.Fg})

	base.ProgressEmpty = applyHL(lipgloss.NewStyle(), Highlight{Fg: progressEmptyHL.Fg})

	base.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	base.Popup = PopupStyleSet{
		Surface: popupSurface,
		Frame: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(popupBorder).
			BorderBackground(popupSurface).
			Background(popupSurface).
			Padding(0, 1),
		Title: applyPopupText(lipgloss.NewStyle(), popupTitleHL, popupSurface).
			MarginBottom(1),
		Content: applyPopupText(lipgloss.NewStyle(), popupContentHL, popupSurface),
		Action: applyHL(lipgloss.NewStyle(), popupActionHL).
			Padding(0, 2),
		ActionFocused: applyHL(lipgloss.NewStyle(), popupActionFocusedHL).
			Padding(0, 2),
		ActionHover: applyHL(lipgloss.NewStyle(), popupActionHoverHL).
			Padding(0, 2),
		ContextMenuFrame: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(contextMenuBorderColor).
			BorderBackground(popupSurface).
			Background(popupSurface),
		ContextMenuItem:         applyHL(lipgloss.NewStyle(), contextMenuItemHL),
		ContextMenuItemFocused:  applyHL(lipgloss.NewStyle(), contextMenuItemFocusedHL),
		ContextMenuItemHover:    applyHL(lipgloss.NewStyle(), contextMenuItemHoverHL),
		ContextMenuItemDisabled: applyHL(lipgloss.NewStyle(), contextMenuItemDisabledHL),
		ContextMenuHeader:       applyHL(lipgloss.NewStyle(), contextMenuHeaderHL),
		ContextMenuSeparator: lipgloss.NewStyle().
			Foreground(contextMenuSeparatorColor).
			Background(popupSurface),
		ScrollTrack: applyHL(lipgloss.NewStyle().
			Background(popupSurface).
			Faint(true), scrollTrackHL),
		ScrollThumb: applyHL(lipgloss.NewStyle().
			Background(popupSurface), scrollThumbHL),
	}

	// Notification surface: reuse popup-style resolution.
	notifSurface := or(theme.Notification.Surface, theme.Surface)
	if notifSurface == nil {
		notifSurface = theme.Background
	}
	if notifSurface == nil {
		notifSurface = theme.AppBackground.Bg
	}
	notifTitleHL := resolveHL(theme.Notification.Title, theme.Primary, nil)
	if notifTitleHL.Bold == nil {
		notifTitleHL.Bold = BoolPtr(true)
	}
	notifMessageHL := resolveHL(theme.Notification.Message, nil, nil)
	// Keep default actions visible without letting their background dominate the notification.
	notifActionBg := theme.Muted
	if muted, ok := colorful.MakeColor(theme.Muted); ok {
		if notifSurface != nil && notifSurface != noColor {
			if surface, ok := colorful.MakeColor(notifSurface); ok {
				notifActionBg = muted.BlendLab(surface, 0.6).Clamped()
			}
		}
	}
	notifActionHL := resolveHL(theme.Notification.Action, theme.Primary, notifActionBg)
	notifActionHoverHL := resolveHL(theme.Notification.ActionHover, notifActionHL.Fg, notifActionHL.Bg)
	if notifActionHoverHL.Underline == nil {
		notifActionHoverHL.Underline = BoolPtr(true)
	}

	notifFrame := func(border color.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			BorderBackground(notifSurface).
			Background(notifSurface).
			Padding(0, 1)
	}

	base.Notification = NotificationStyleSet{
		Surface:      notifSurface,
		InfoFrame:    notifFrame(or(theme.Notification.InfoBorder, theme.Info)),
		SuccessFrame: notifFrame(or(theme.Notification.SuccessBorder, theme.Success)),
		WarningFrame: notifFrame(or(theme.Notification.WarningBorder, theme.Warning)),
		ErrorFrame:   notifFrame(or(theme.Notification.ErrorBorder, theme.Error)),
		Title:        applyPopupText(lipgloss.NewStyle(), notifTitleHL, notifSurface),
		Message:      applyPopupText(lipgloss.NewStyle(), notifMessageHL, notifSurface),
		Close:        applyHL(lipgloss.NewStyle(), Highlight{Fg: theme.Muted, Bg: notifSurface}),
		Action: applyHL(lipgloss.NewStyle(), notifActionHL).
			Padding(0, 1),
		ActionHover: applyHL(lipgloss.NewStyle(), notifActionHoverHL).
			Padding(0, 1),
		InfoIcon:    orString(theme.Notification.InfoIcon, "ℹ "),
		SuccessIcon: orString(theme.Notification.SuccessIcon, "✓ "),
		WarningIcon: orString(theme.Notification.WarningIcon, "⚠ "),
		ErrorIcon:   orString(theme.Notification.ErrorIcon, "✗ "),
	}

	// Semantic colors
	base.Success = lipgloss.NewStyle().Foreground(theme.Success)
	base.Warning = lipgloss.NewStyle().Foreground(theme.Warning)
	base.Error = lipgloss.NewStyle().Foreground(theme.Error)
	base.Info = lipgloss.NewStyle().Foreground(theme.Info)
	base.Muted = lipgloss.NewStyle().Foreground(theme.Muted)

	// HintKey
	hintKeyColor := theme.HintKey
	if hintKeyColor == nil {
		hintKeyColor = theme.Muted
	}
	base.HintKey = lipgloss.NewStyle().Foreground(hintKeyColor)

	// Status bar
	base.StatusBar = lipgloss.NewStyle().
		Foreground(statusBarFg).
		Background(statusBarBg)
	base.StatusBarText = lipgloss.NewStyle().
		Foreground(statusBarTextFg).
		Background(statusBarBg)

	base.StatusBarBreadcrumb = lipgloss.NewStyle().
		Foreground(theme.StatusBarBreadcrumb.Fg).
		Background(breadcrumbBg)
	base.StatusBarBreadcrumbBg = lipgloss.NewStyle().
		Background(breadcrumbBg)

	// Status bar nuggets
	base.StatusBarNugget = applyHL(lipgloss.NewStyle(), Highlight{Fg: statusBarNuggetFg}).
		Padding(0, 1)
	base.StatusBarNuggetLabel = base.StatusBarNugget.
		Foreground(statusBarNuggetLabelFg).
		Background(statusBarNuggetLabelBg)
	base.StatusBarTime = lipgloss.NewStyle().
		Foreground(statusBarTimeFg).
		Background(statusBarTimeBg).
		Padding(0, 1)

	// Breadcrumb hover/click styles
	base.StatusBarBreadcrumbHover = applyHL(lipgloss.NewStyle(), breadcrumbHoverHL)
	base.StatusBarBreadcrumbClick = applyHL(lipgloss.NewStyle(), breadcrumbClickHL)
	base.StatusBarBreadcrumbSep = lipgloss.NewStyle().Inherit(base.StatusBarBreadcrumb).Foreground(breadcrumbSepColor)

	// App background
	base.AppBackground = lipgloss.NewStyle().Background(appBg)

	// Normal style
	base.Normal = lipgloss.NewStyle()

	// Back button
	base.BackButton = applyHL(lipgloss.NewStyle(), backButtonHL)
	base.BackButtonHover = applyHL(lipgloss.NewStyle(), backButtonHoverHL)

	// Build Custom styles via reflection
	if theme.Custom != nil {
		customVal := reflect.ValueOf(theme.Custom)
		// Dereference pointer types
		for customVal.Kind() == reflect.Ptr {
			customVal = customVal.Elem()
		}
		if customVal.Kind() == reflect.Struct {
			typ := customVal.Type()
			resolved := make(map[string]lipgloss.Style, typ.NumField())
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if !field.IsExported() {
					continue
				}
				if field.Type != reflect.TypeOf(Highlight{}) {
					continue
				}
				hl := customVal.Field(i).Interface().(Highlight)
				hl = resolvePreset(hl)
				resolved[field.Name] = applyHL(lipgloss.NewStyle(), hl)
			}
			base.Custom = resolved
		}
	}

	// In accessible mode, add color-independent emphasis (reverse/bold/underline)
	// so selection and focus stay visible when the terminal cannot render color.
	if AccessibleMode() {
		base = applyAccessibleEmphasis(base)
	}

	return base
}

// DefaultStyleSet returns a StyleSet using DefaultTheme (auto-adaptive).
// It detects the terminal background and builds styles for the appropriate
// dark or light theme. Safe to call multiple times; detection is cached.
func DefaultStyleSet() StyleSet {
	return NewStyleSet(DefaultTheme())
}

// ---- global StyleSet ----

var currentStyleSet = DefaultStyleSet()

// styleGeneration is bumped every time the global StyleSet changes. Downstream
// renderers that cache styled output can fold this value into their cache keys
// so a theme switch invalidates any output still carrying the old theme's
// colors, avoiding stale-color residue on the screen.
var styleGeneration uint64

// CurrentStyleSet returns the active global StyleSet.
// By default this is built from DefaultTheme. Call SetStyleSet to override
// with a custom theme constructed programmatically.
//
// Usage in downstream apps:
//
//	theme := style.VSCodeDarkTheme()
//	style.SetStyleSet(style.NewStyleSet(theme))
func CurrentStyleSet() StyleSet {
	return currentStyleSet
}

// StyleGeneration returns a counter that increments on every SetStyleSet call.
// Cache keys keyed on this value are invalidated whenever the theme changes.
func StyleGeneration() uint64 {
	return styleGeneration
}

// SetStyleSet sets the global StyleSet. Call during application startup
// before any UI rendering. The framework's internal View methods all read
// from this global StyleSet, so setting it once at init is sufficient.
func SetStyleSet(s StyleSet) {
	currentStyleSet = s
	styleGeneration++
}

// FG applies a foreground color to a style and renders the content.
// This is the lipgloss v2 equivalent of SetFgStyle.
func FG(content string, c color.Color) string {
	return lipgloss.NewStyle().Foreground(c).Render(content)
}

// FGBG applies foreground and background colors to a style and renders the content.
func FGBG(content string, fg, bg color.Color) string {
	return lipgloss.NewStyle().Foreground(fg).Background(bg).Render(content)
}

// Normal renders content without any ANSI styling.
// Unlike SetNormalStyle, this does NOT emit raw \x1b[0m reset sequences.
func Normal(content string) string {
	return lipgloss.NewStyle().Render(content)
}

// Dim applies a dimming style to content, suitable for background text behind popups.
func Dim(content string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightBlack).Render(content)
}

// ApplyHL applies a Highlight configuration to a lipgloss.Style and returns it.
// This is the public version of the internal applyHL closure used during StyleSet
// construction. Fields set to nil are treated as "not set" and are ignored.
func ApplyHL(s lipgloss.Style, hl Highlight) lipgloss.Style {
	if hl.Fg != nil {
		s = s.Foreground(hl.Fg)
	}
	if hl.Bg != nil {
		s = s.Background(hl.Bg)
	}
	if hl.Bold != nil {
		s = s.Bold(*hl.Bold)
	}
	if hl.Italic != nil {
		s = s.Italic(*hl.Italic)
	}
	if hl.Underline != nil {
		s = s.Underline(*hl.Underline)
	}
	return s
}

// CustomStyles converts a StyleSet's internal Custom map into a typed struct
// with lipgloss.Style fields. Define a struct matching field names with the
// Theme.Custom struct, but using lipgloss.Style instead of Highlight fields:
//
//	type AppCustomStyles struct {
//	    PlayerColor lipgloss.Style
//	}
//	custom := style.CustomStyles[AppCustomStyles](style.CurrentStyleSet())
//	playerColor := custom.PlayerColor
//
// Use this for type-safe, flat struct access to resolved custom styles.
func CustomStyles[C any](ss StyleSet) C {
	var result C
	m, ok := ss.Custom.(map[string]lipgloss.Style)
	if !ok || m == nil {
		return result
	}
	rv := reflect.ValueOf(&result).Elem()
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return result
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		if s, found := m[f.Name]; found {
			rv.Field(i).Set(reflect.ValueOf(s))
		}
	}
	return result
}
