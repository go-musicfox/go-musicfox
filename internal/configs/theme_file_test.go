package configs

import (
	"image/color"
	"testing"

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
