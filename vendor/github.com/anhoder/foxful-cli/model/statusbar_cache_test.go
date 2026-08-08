package model

import (
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/util"
)

// TestStatusBarCacheInvalidatesOnNavigation ensures the breadcrumb content
// (built from m.menuStack + m.menuTitle) participates in the status bar cache
// key. Navigation changes only those two fields — with the key missing them,
// the cached view keeps showing the stale breadcrumb until the next minute
// boundary, song change, or theme switch.
func TestStatusBarCacheInvalidatesOnNavigation(t *testing.T) {
	a := &App{windowWidth: 80, windowHeight: 24}
	m := &Main{
		menuTitle:           &MenuItem{Title: "Root Menu"},
		menuStack:           &util.Stack{},
		hoveredBreadcrumbIdx: -1,
	}
	sb := &DefaultStatusBar{}

	first := sb.View(a, m)
	if !strings.Contains(stripAnsiForTest(first), "Root Menu") {
		t.Fatalf("first render missing breadcrumb: %q", first)
	}

	// Enter a submenu: menuTitle is replaced while width, minute, style
	// generation, components and hover state all stay identical.
	m.menuTitle = &MenuItem{Title: "Sub Menu"}
	second := sb.View(a, m)
	if !strings.Contains(stripAnsiForTest(second), "Sub Menu") {
		t.Fatalf("stale breadcrumb after navigation: %q", second)
	}
	if strings.Contains(stripAnsiForTest(second), "Root Menu") {
		t.Fatalf("old breadcrumb still visible after navigation: %q", second)
	}

	// Push the parent onto the stack (returning-upwards shape): the path
	// becomes "Root Menu / Sub Menu".
	m.menuStack.Push(&menuStackItem{menuTitle: &MenuItem{Title: "Root Menu"}})
	third := sb.View(a, m)
	if !strings.Contains(stripAnsiForTest(third), "Root Menu") || !strings.Contains(stripAnsiForTest(third), "Sub Menu") {
		t.Fatalf("stacked breadcrumb missing after push: %q", third)
	}

	// Identical state must still hit the cache (byte-identical output).
	fourth := sb.View(a, m)
	if fourth != third {
		t.Fatalf("cache miss for identical state: %q != %q", fourth, third)
	}
}
