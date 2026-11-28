package configs

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

func TestThemeMenuItemsUseAppBackground(t *testing.T) {
	theme, err := parseThemeFile(`
name = "menu-background"

[dark.highlights]
selectedItemBg = "#121212"

[dark.app]
background = "#E1E2E3"
`)
	if err != nil {
		t.Fatalf("parse theme: %v", err)
	}

	styles := style.NewStyleSet(theme.Dark.toTheme())
	if !sameColor(styles.MenuItem.GetBackground(), styles.AppBackground.GetBackground()) {
		t.Fatalf("menu item background = %v, want app background %v", styles.MenuItem.GetBackground(), styles.AppBackground.GetBackground())
	}
}

func sameColor(left, right color.Color) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	leftR, leftG, leftB, leftA := left.RGBA()
	rightR, rightG, rightB, rightA := right.RGBA()
	return leftR == rightR && leftG == rightG && leftB == rightB && leftA == rightA
}

func TestThemePlaybarColorsFollowPalette(t *testing.T) {
	theme, err := parseThemeFile(`
name = "palette-playbar"

[dark]
primary = "#112233"
accent = "#445566"
success = "#778899"

[dark.app]
playbarSong = "#ABCDEF"
`)
	if err != nil {
		t.Fatalf("parse theme: %v", err)
	}

	styles := style.NewStyleSet(theme.Dark.toTheme())
	colors := style.CustomStyles[ResolvedAppStyles](styles)

	if got, want := colors.PlaybarSong.GetForeground(), lipgloss.Color("#ABCDEF"); !sameColor(got, want) {
		t.Fatalf("explicit playbarSong = %v, want %v", got, want)
	}

	checks := []struct {
		name string
		got  color.Color
		want color.Color
	}{
		{"playbarMode", colors.PlaybarMode.GetForeground(), lipgloss.Color("#445566")},
		{"playbarPlaying", colors.PlaybarPlaying.GetForeground(), lipgloss.Color("#778899")},
		{"playbarTime", colors.PlaybarTime.GetForeground(), lipgloss.Color("#112233")},
		{"playbarHover", colors.PlaybarHover.GetForeground(), lipgloss.Color("#112233")},
		{"statusBarLogo", colors.StatusBarLogo.GetForeground(), lipgloss.Color("#112233")},
		{"playbarProgressStart", colors.PlaybarProgressStart.GetForeground(), lipgloss.Color("#112233")},
		{"playbarProgressEnd", colors.PlaybarProgressEnd.GetForeground(), lipgloss.Color("#445566")},
	}
	for _, c := range checks {
		if !sameColor(c.got, c.want) {
			t.Errorf("%s = %v, want palette color %v", c.name, c.got, c.want)
		}
	}
}

func TestBuiltinThemesConfigurePlaybarColors(t *testing.T) {
	themes := loadBuiltinThemes()
	for _, name := range []string{"Default", "Transparent", "Dracula", "Nord", "Gruvbox", "GitHub"} {
		tf, ok := themes[name]
		if !ok {
			t.Fatalf("built-in theme %q not loaded", name)
		}
		for _, variant := range []struct {
			label string
			theme ThemeVariant
		}{
			{"dark", tf.Dark},
			{"light", tf.Light},
		} {
			colors := style.CustomStyles[ResolvedAppStyles](style.NewStyleSet(variant.theme.toTheme()))
			fields := []struct {
				name  string
				style lipgloss.Style
			}{
				{"playbarMode", colors.PlaybarMode},
				{"playbarVolume", colors.PlaybarVolume},
				{"playbarPlaying", colors.PlaybarPlaying},
				{"playbarPaused", colors.PlaybarPaused},
				{"playbarHeartLiked", colors.PlaybarHeartLiked},
				{"playbarHeartUnliked", colors.PlaybarHeartUnliked},
				{"playbarArtist", colors.PlaybarArtist},
				{"playbarSong", colors.PlaybarSong},
				{"playbarTime", colors.PlaybarTime},
				{"playbarHover", colors.PlaybarHover},
				{"statusBarLogo", colors.StatusBarLogo},
				{"playbarProgressStart", colors.PlaybarProgressStart},
				{"playbarProgressEnd", colors.PlaybarProgressEnd},
			}
			for _, f := range fields {
				if IsUnsetColor(f.style.GetForeground()) {
					t.Errorf("%s/%s: %s is unset", name, variant.label, f.name)
				}
			}
		}
	}
}
