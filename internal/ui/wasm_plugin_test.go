package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/wasm"
)

// wasmTestRepoRoot returns the repository root, derived from this test file's
// path: internal/ui/wasm_plugin_test.go -> internal/ui -> internal -> repo root.
func wasmTestRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// buildWasmExample compiles the examples/wasm/hello plugin with Go's wasip1
// toolchain into dir (as main.wasm). The build is skipped when the toolchain is
// unavailable so CI stays robust.
func buildWasmExample(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", filepath.Join(dir, "main.wasm"), "./examples/wasm/hello")
	cmd.Dir = wasmTestRepoRoot(t)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("go wasm build unavailable: %v (%s)", err, out)
	}
}

// --- Menu surface ---

// TestWasmPluginMenuSurface checks the pure menu surface of WasmPluginMenu: it
// is an action menu (not playable / not locatable), renders its declared title,
// produces no submenu and reports its declared key. Run is not exercised here —
// the plugin stub has no backing instance, which is fine for surface checks.
func TestWasmPluginMenuSurface(t *testing.T) {
	menu := &WasmPluginMenu{
		BaseMenu: BaseMenu{},
		plugin:   &wasm.Plugin{ID: "x", Name: "y"},
		decl:     wasm.MenuDecl{Key: "wasm_x", Title: "X"},
	}

	if got := menu.GetMenuKey(); got != "wasm_x" {
		t.Fatalf("GetMenuKey() = %q, want %q", got, "wasm_x")
	}
	if menu.IsPlayable() {
		t.Fatal("IsPlayable() = true, want false")
	}
	if menu.IsLocatable() {
		t.Fatal("IsLocatable() = true, want false")
	}

	views := menu.MenuViews()
	if len(views) != 1 {
		t.Fatalf("MenuViews() returned %d items, want 1", len(views))
	}
	if views[0].Title != "X" {
		t.Fatalf("MenuViews()[0].Title = %q, want %q", views[0].Title, "X")
	}

	if sub := menu.SubMenu(nil, 0); sub != nil {
		t.Fatalf("SubMenu() = %v, want nil", sub)
	}
}

// --- Response helpers ---

// TestWasmLevelToModel proves the wasm.Response Level string maps to the model
// notification levels; unknown and empty levels default to info.
func TestWasmLevelToModel(t *testing.T) {
	cases := []struct {
		level string
		want  model.NotificationLevel
	}{
		{"", model.NotificationInfo},
		{"info", model.NotificationInfo},
		{"success", model.NotificationSuccess},
		{"warning", model.NotificationWarning},
		{"error", model.NotificationError},
		{"bogus", model.NotificationInfo},
	}
	for _, tc := range cases {
		if got := wasmLevelToModel(tc.level); got != tc.want {
			t.Fatalf("wasmLevelToModel(%q) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

// TestWasmResponseSpec proves toast/view responses map to a multi-line in-app
// notification spec with the declared title, message and mapped level.
func TestWasmResponseSpec(t *testing.T) {
	spec := wasmResponseSpec(wasm.Response{Action: "toast", Title: "T", Message: "M", Level: "warning"})
	if spec.Title != "T" || spec.Message != "M" || spec.Level != model.NotificationWarning {
		t.Fatalf("toast spec = %+v, want title T message M level warning", spec)
	}

	// view renders the same shape (MVP: content through the toast message).
	spec = wasmResponseSpec(wasm.Response{Action: "view", Title: "V", Message: "body", Level: ""})
	if spec.Title != "V" || spec.Message != "body" || spec.Level != model.NotificationInfo {
		t.Fatalf("view spec = %+v, want title V message body level info", spec)
	}
}

// TestRunWasmSideEffects proves the exec side effect reports success on a clean
// run and an error spec (with the error message) on failure. open_url is
// intentionally not unit-tested: open.Start would launch a real browser.
func TestRunWasmSideEffects(t *testing.T) {
	// exec success
	spec := runWasmSideEffects(wasm.Response{Action: "exec", Command: "true"})
	if spec.Level != model.NotificationSuccess {
		t.Fatalf("exec success spec = %+v, want level success", spec)
	}
	if spec.Message != "true" {
		t.Fatalf("exec success spec.Message = %q, want %q", spec.Message, "true")
	}

	// exec failure (the binary does not exist)
	spec = runWasmSideEffects(wasm.Response{Action: "exec", Command: "definitely-not-a-real-binary-xyz"})
	if spec.Level != model.NotificationError {
		t.Fatalf("exec failure spec = %+v, want level error", spec)
	}
	if spec.Message == "" {
		t.Fatal("exec failure spec.Message is empty, want the exec error")
	}

	// Unknown/empty action: ignored (empty spec; the caller skips Notify).
	spec = runWasmSideEffects(wasm.Response{Action: "bogus"})
	if spec.Level != model.NotificationInfo || spec.Title != "" || spec.Message != "" {
		t.Fatalf("unknown action spec = %+v, want zero spec", spec)
	}
}

// --- Registration wiring (the valuable one) ---

// TestRegisterWasmPlugin builds a real WASM plugin, loads it through a real
// wasm.Manager and runs registerWasmPlugin. It asserts the menu provider is
// buildable through the registry, the main-menu item is present and the plugin
// attribution records both keys. The registry is package-global and accumulates
// across a single test run, so the menu key is derived from the test name to
// stay unique.
func TestRegisterWasmPlugin(t *testing.T) {
	key := "wasm_test_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, t.Name())

	root := t.TempDir()
	pluginDir := filepath.Join(root, "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	buildWasmExample(t, pluginDir)

	manifest := fmt.Sprintf(`
id = "hello"
name = "Hello WASM"
version = "0.1.0"
author = "test"
description = "ui registration integration test"

[[menus]]
key = %q
title = "Hello"
after = ""
export = "run"
`, key)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := wasm.NewManager()
	if err != nil {
		t.Fatalf("wasm.NewManager: %v", err)
	}
	defer m.Close(context.Background())

	if errs := m.LoadDir(context.Background(), root); len(errs) != 0 {
		t.Fatalf("unexpected load errors: %v", errs)
	}
	plugins := m.Plugins()
	if len(plugins) != 1 {
		t.Fatalf("Plugins() returned %d plugins, want 1", len(plugins))
	}
	p := plugins[0]

	registerWasmPlugin(p)

	// The menu provider is buildable through the registry.
	menu, err := BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(%q): %v", key, err)
	}
	wm, ok := menu.(*WasmPluginMenu)
	if !ok {
		t.Fatalf("BuildMenu(%q) = %T, want *WasmPluginMenu", key, menu)
	}
	if got := wm.GetMenuKey(); got != key {
		t.Fatalf("GetMenuKey() = %q, want %q", got, key)
	}

	// The main-menu item is present.
	var found bool
	for _, item := range MainMenuPluginItems() {
		if item.Key == key {
			found = true
			if item.Title != "Hello" {
				t.Fatalf("main-menu item title = %q, want %q", item.Title, "Hello")
			}
		}
	}
	if !found {
		t.Fatalf("main-menu item %q not registered", key)
	}

	// The WithPlugin attribution records the menu + main-menu keys.
	info := pluginInfoSnapshot(t, "hello")
	if info == nil {
		t.Fatalf("plugin %q not declared", "hello")
	}
	if !containsString(info.MenuKeys, key) || !containsString(info.MainMenuItems, key) {
		t.Fatalf("plugin info = %+v, want MenuKeys+MainMenuItems to contain %q", info, key)
	}
}
