package ui_test

// PROTO: Phase 3.6 plugin-boundary compile proof. This file lives in the
// external test package ui_test — OUTSIDE package ui — and proves that a
// plugin-shaped menu can now be written from an external package:
//
//   - it embeds the exported ui.BaseMenu (previously the unexported baseMenu
//     could not be named outside package ui);
//   - its RegisterMenu factory closure is typed `func(base ui.BaseMenu, ...)`,
//     which compiles because baseMenu is an alias of BaseMenu (interchangeable
//     with the internal factory signature);
//   - ui.BuildMenu accepts a ui.BaseMenu value as the base argument.
//
// This mirrors the external-plugin form documented in docs/plugin_development.md.
// It is a verification artifact, NOT a shipped plugin.

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// externalExampleMenu is a plugin-shaped menu embedding the exported BaseMenu.
// It implements ui.Menu by overriding GetMenuKey/MenuViews/SubMenu/BeforeEnterMenuHook
// (the embedded BaseMenu provides the rest, including IsPlayable/IsLocatable).
type externalExampleMenu struct {
	ui.BaseMenu
	menus []model.MenuItem
}

func (m *externalExampleMenu) GetMenuKey() string { return "external_example_menu" }

func (m *externalExampleMenu) MenuViews() []model.MenuItem { return m.menus }

func (m *externalExampleMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

func (m *externalExampleMenu) BeforeEnterMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		m.menus = []model.MenuItem{{Title: "external plugin menu"}}
		return true, nil
	}
}

// init registers the external menu at compile time (import + init()). The
// closure's base parameter is ui.BaseMenu — the internal signature uses the
// baseMenu alias, so this is the exact boundary being proven.
func init() {
	ui.RegisterMenu("external_example_menu", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return &externalExampleMenu{BaseMenu: base}, nil
	})
}

// TestExternalPluginBoundaryCompilesAndBuilds proves the external factory
// shape registers and builds through the real registry, and that the exported
// BaseMenu forwarding methods are reachable (nil-safe on a zero base).
func TestExternalPluginBoundaryCompilesAndBuilds(t *testing.T) {
	menu, err := ui.BuildMenu("external_example_menu", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(external_example_menu) error = %v", err)
	}
	ext, ok := menu.(*externalExampleMenu)
	if !ok {
		t.Fatalf("BuildMenu(external_example_menu) = %T, want *externalExampleMenu", menu)
	}
	if ext.GetMenuKey() != "external_example_menu" {
		t.Fatalf("GetMenuKey() = %q, want external_example_menu", ext.GetMenuKey())
	}
	if ok, _ := ext.BeforeEnterMenuHook()(nil); !ok {
		t.Fatal("external menu BeforeEnterMenuHook returned ok=false")
	}
	if len(ext.MenuViews()) != 1 {
		t.Fatalf("MenuViews() = %v, want 1 static item", ext.MenuViews())
	}

	// Exported forwarding methods compile and are nil-safe on a zero BaseMenu.
	zero := ui.BaseMenu{}
	if zero.Player() != nil || zero.User() != nil || zero.TrackManager() != nil ||
		zero.LyricService() != nil || zero.DesktopLyrics() != nil || zero.CoverRenderer() != nil ||
		zero.ShareSvc() != nil || zero.Lastfm() != nil || zero.Ctx() != nil ||
		zero.App() != nil || zero.MustMain() != nil || zero.Rerender() != nil ||
		zero.Search() != nil || zero.Netease() != nil {
		t.Fatal("zero BaseMenu forwarding methods must return zero values")
	}
	if page, cmd := zero.ToLoginPage(nil); page != nil || cmd != nil {
		t.Fatalf("zero BaseMenu.ToLoginPage = (%v, %v), want (nil, nil)", page, cmd)
	}
	if page, cmd := zero.ToSearchPage(ui.StSingleSong); page != nil || cmd != nil {
		t.Fatalf("zero BaseMenu.ToSearchPage = (%v, %v), want (nil, nil)", page, cmd)
	}
	// Services (Phase 3.9): the accessor getter is nil-safe on a zero base.
	if svc := zero.Services(); svc != nil {
		t.Fatalf("zero BaseMenu.Services() = %v, want nil", svc)
	}
}
