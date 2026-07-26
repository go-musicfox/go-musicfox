package style

import (
	"os"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// ---- color-independent emphasis (accessible mode) ----
//
// Many UI states in a TUI are communicated purely through color: a selected
// row gets a colored background, an active tab gets a colored foreground, a
// focused button changes hue. When the terminal cannot render color, those
// distinctions vanish and the interface becomes ambiguous.
//
// A terminal reports "no color" in several common situations:
//   - NO_COLOR is set (see https://no-color.org)
//   - CLICOLOR=0
//   - TERM=dumb
//   - output is piped to a file or another program (not a TTY)
//
// The underlying color writer (github.com/charmbracelet/colorprofile, used by
// lipgloss) already strips color escape sequences in these cases, so we must
// NOT strip color ourselves. Instead, accessible mode adds emphasis that
// survives color stripping — reverse video, bold, and underline — so that
// selection and focus remain visible in a monochrome terminal.

var (
	accessibleMu   sync.RWMutex
	accessibleMode = detectAccessibleMode()
)

// detectAccessibleMode reports whether the active output cannot render color,
// which means color-based emphasis would be invisible.
func detectAccessibleMode() bool {
	// ASCII and NoTTY profiles carry no color. Anything at or below ASCII
	// (ASCII, NoTTY) renders monochrome.
	return colorprofile.Detect(os.Stdout, os.Environ()) <= colorprofile.Ascii
}

// AccessibleMode reports whether color-independent emphasis is currently
// active. It is auto-detected at startup from the terminal color profile and
// can be overridden with SetAccessibleMode.
func AccessibleMode() bool {
	accessibleMu.RLock()
	defer accessibleMu.RUnlock()
	return accessibleMode
}

// SetAccessibleMode explicitly enables or disables color-independent emphasis,
// overriding auto-detection. When enabled, selection and focus states gain
// reverse-video and bold emphasis so they remain distinguishable in terminals
// that do not render color.
//
// It rebuilds the global StyleSet so the change takes effect immediately for
// callers that read style.CurrentStyleSet(). Apps using an app-scoped StyleSet
// should rebuild it (via NewStyleSet) after calling this.
func SetAccessibleMode(on bool) {
	accessibleMu.Lock()
	changed := accessibleMode != on
	accessibleMode = on
	accessibleMu.Unlock()

	if changed {
		currentStyleSet = NewStyleSet(currentStyleSet.theme)
	}
}

// applyAccessibleEmphasis augments a StyleSet with color-independent emphasis.
// It is applied by NewStyleSet when AccessibleMode is active. The added
// attributes (reverse, bold, underline) render even after color is stripped,
// keeping the following states visible in monochrome terminals:
//
//   - the selected menu/list item
//   - hovered menu items and notification actions
//   - focused buttons and popup actions
//   - the hovered back button
func applyAccessibleEmphasis(s StyleSet) StyleSet {
	s.SelectedItem = s.SelectedItem.Reverse(true).Bold(true)
	s.SelectedItemHover = s.SelectedItemHover.Reverse(true).Bold(true).Underline(true)
	s.MenuItemHover = s.MenuItemHover.Underline(true).Bold(true)
	s.Button = s.Button.Reverse(true).Bold(true)
	s.ButtonBlurred = s.ButtonBlurred.Underline(true)
	s.BackButtonHover = s.BackButtonHover.Reverse(true).Bold(true)
	s.Popup.ActionFocused = s.Popup.ActionFocused.Reverse(true).Bold(true)
	s.Popup.ActionHover = s.Popup.ActionHover.Underline(true)
	s.Notification.Action = s.Notification.Action.Reverse(true).Bold(true)
	s.Notification.ActionHover = s.Notification.ActionHover.Reverse(true).Bold(true).Underline(true)
	return s
}

// ---- high-contrast theme presets ----
//
// These presets maximize luminance contrast between foreground and background
// (near-pure black and white) and use saturated, well-separated status colors.
// They are intended for low-vision users and environments that require a
// high-contrast appearance. Pair them via WithThemePair, or select one
// explicitly with SetStyleSet(NewStyleSet(style.HighContrastDarkTheme())).

// HighContrastDarkTheme returns a maximum-contrast theme for dark terminals:
// pure white text on pure black, with saturated status colors and a white
// border for strong element separation.
func HighContrastDarkTheme() Theme {
	return Theme{
		Primary:                lipgloss.Color("#FFFFFF"),
		Secondary:              lipgloss.Color("#FFFF00"),
		Accent:                 lipgloss.Color("#00FFFF"),
		Success:                lipgloss.Color("#00FF00"),
		Warning:                lipgloss.Color("#FFFF00"),
		Error:                  lipgloss.Color("#FF5555"),
		Info:                   lipgloss.Color("#00FFFF"),
		Muted:                  lipgloss.Color("#C6C6C6"),
		HintKey:                lipgloss.Color("#FFFFFF"),
		StatusBarBreadcrumbSep: lipgloss.Color("#FFFFFF"),
		Background:             lipgloss.Color("#000000"),
		Foreground:             lipgloss.Color("#FFFFFF"),
		Border:                 lipgloss.Color("#FFFFFF"),
		Surface:                lipgloss.Color("#000000"),
	}
}

// HighContrastLightTheme returns a maximum-contrast theme for light terminals:
// pure black text on pure white, with dark saturated status colors and a black
// border for strong element separation.
func HighContrastLightTheme() Theme {
	return Theme{
		Primary:                lipgloss.Color("#000000"),
		Secondary:              lipgloss.Color("#0000CC"),
		Accent:                 lipgloss.Color("#0000CC"),
		Success:                lipgloss.Color("#006600"),
		Warning:                lipgloss.Color("#8A4B00"),
		Error:                  lipgloss.Color("#CC0000"),
		Info:                   lipgloss.Color("#0000CC"),
		Muted:                  lipgloss.Color("#3A3A3A"),
		HintKey:                lipgloss.Color("#000000"),
		StatusBarBreadcrumbSep: lipgloss.Color("#000000"),
		Background:             lipgloss.Color("#FFFFFF"),
		Foreground:             lipgloss.Color("#000000"),
		Border:                 lipgloss.Color("#000000"),
		Surface:                lipgloss.Color("#FFFFFF"),
	}
}

// HighContrastTheme returns an adaptive high-contrast theme, selecting the dark
// or light variant based on detected terminal background.
func HighContrastTheme() Theme {
	if detectedDarkBg {
		return HighContrastDarkTheme()
	}
	return HighContrastLightTheme()
}
