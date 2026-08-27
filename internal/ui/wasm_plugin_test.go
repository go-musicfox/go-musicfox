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

	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
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

// wasmTestKey derives a unique menu/command key from the test name so the
// package-global registries stay pollution-free across a single test run.
func wasmTestKey(t *testing.T) string {
	t.Helper()
	return "wasm_test_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, t.Name())
}

// writeWasmTestPlugin builds a real WASM plugin tree (examples/wasm/hello +
// a manifest whose menu key is derived from the test name) under root and
// returns the plugin directory.
func writeWasmTestPlugin(t *testing.T, root string) string {
	t.Helper()
	key := wasmTestKey(t)
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
description = "sink integration test"

[[menus]]
key = %q
title = "Hello"
after = ""
export = "run"
`, key)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return pluginDir
}

// loadWasmTestPlugin loads a freshly built test plugin through the P6 scope
// pipeline (wasm.LoadIntoScope) with the given sink and returns the manager
// plus the loaded plugin. The scope — which owns the manager and plugin
// instances — is disposed on test cleanup. The load order mirrors the TUI: the
// wasm sub-scope (ManagerPlugin) is started first, then LoadIntoScope
// AddAndStarts the adapter on the already-started scope.
func loadWasmTestPlugin(t *testing.T, sink wasm.RegistrySink) (*wasm.Manager, *wasm.Plugin) {
	t.Helper()
	root := t.TempDir()
	writeWasmTestPlugin(t, root)

	scope := framework.NewScope()
	fctx := &framework.Context{}
	if err := scope.Add(&wasm.ManagerPlugin{}); err != nil {
		t.Fatalf("Add(wasm.ManagerPlugin): %v", err)
	}
	if err := scope.Start(fctx); err != nil {
		t.Fatalf("scope Start: %v", err)
	}
	mgr, errs := wasm.LoadIntoScope(context.Background(), fctx, scope, root, sink)
	if len(errs) != 0 {
		t.Fatalf("LoadIntoScope returned errors: %v", errs)
	}
	t.Cleanup(func() { _ = scope.Dispose() })

	plugins := mgr.Plugins()
	if len(plugins) != 1 {
		t.Fatalf("Plugins() returned %d plugins, want 1", len(plugins))
	}
	return mgr, plugins[0]
}

// --- CommandsOf mapping ---

// TestWasmSinkCommandsOf proves CommandsOf maps each MenuDecl into a
// frontend.Command with the plugin id stamped, the manifest key/title/after
// preserved and a nil Show plus a non-nil Run closure. Run is exercised in
// TestWasmSinkCallWasm.
func TestWasmSinkCommandsOf(t *testing.T) {
	key := wasmTestKey(t)
	p := &wasm.Plugin{
		ID:   "hello",
		Name: "Hello WASM",
		Menus: []wasm.MenuDecl{
			{Key: key, Title: "Hello", After: "help", Export: "run", Args: map[string]any{"name": "musicfox"}},
			{Key: key + "_b", Title: "Hello B", After: ""},
		},
	}

	cmds := wasm.CommandsOf(p)
	if len(cmds) != 2 {
		t.Fatalf("CommandsOf() returned %d commands, want 2", len(cmds))
	}

	cmd := cmds[0]
	if cmd.Key != key || cmd.Title != "Hello" || cmd.After != "help" {
		t.Fatalf("command = %+v, want key %q title Hello after help", cmd, key)
	}
	if cmd.PluginID != p.ID {
		t.Fatalf("PluginID = %q, want %q", cmd.PluginID, p.ID)
	}
	if cmd.Show != nil {
		t.Fatal("Show must be nil (always available)")
	}
	if cmd.Run == nil {
		t.Fatal("Run must be non-nil")
	}

	// A MenuDecl with an empty After stays empty (end-append chain tail).
	if cmds[1].After != "" {
		t.Fatalf("second command After = %q, want empty", cmds[1].After)
	}
}

// --- callWasm end-to-end ---

// TestWasmSinkCallWasm builds the example plugin, maps it through CommandsOf
// and invokes the command's Run closure with a full context snapshot, proving
// callWasm copies the frontend context into the guest request and maps the
// response back into a CommandResult.
func TestWasmSinkCallWasm(t *testing.T) {
	root := t.TempDir()
	pluginDir := writeWasmTestPlugin(t, root)

	// args action=toast switches the example plugin's response to a toast.
	manifest, err := os.ReadFile(filepath.Join(pluginDir, "manifest.toml"))
	if err != nil {
		t.Fatal(err)
	}
	manifest = []byte(strings.Replace(string(manifest), "export = \"run\"", "export = \"run\"\nargs = { name = \"musicfox\", action = \"toast\" }", 1))
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.toml"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	mgr, errs := wasm.LoadAndRegister(context.Background(), root, &noopSink{})
	if len(errs) != 0 {
		t.Fatalf("LoadAndRegister returned errors: %v", errs)
	}
	defer mgr.Close(context.Background())

	p := mgr.Plugins()[0]
	cmd := wasm.CommandsOf(p)[0]

	res := cmd.Run(frontend.CommandContext{
		UserID:   123,
		UserName: "musicfox",
		Playing:  true,
		Song:     &frontend.SongInfo{ID: 1, Name: "Song", Artist: "Artist", Album: "Album"},
	})
	if res.Action != "toast" {
		t.Fatalf("Action = %q, want toast", res.Action)
	}
	if res.Title != "WASM Hello" {
		t.Fatalf("Title = %q, want %q", res.Title, "WASM Hello")
	}
	if !strings.Contains(res.Message, "musicfox") || !strings.Contains(res.Message, "Song") || !strings.Contains(res.Message, "Artist") {
		t.Fatalf("Message %q does not echo the args/context song", res.Message)
	}
	if res.Level != "info" {
		t.Fatalf("Level = %q, want info", res.Level)
	}

	// A second call works: the instance survives between runs.
	if res := cmd.Run(frontend.CommandContext{}); res.Action != "toast" {
		t.Fatalf("second Run Action = %q, want toast", res.Action)
	}
}

// --- Sink registration wiring (the valuable one) ---

// TestWasmSinkRegisterAndMenus loads a real WASM plugin through the P6 scope
// pipeline with the TUI sink and proves the command lands in the frontend
// registry with the plugin id stamped, registerCommandMenus adapts it into a
// *CommandMenu provider plus main-menu item, and the WithPlugin attribution
// records the command key. The registries are package-global, so the menu key
// is derived from the test name to stay unique.
func TestWasmSinkRegisterAndMenus(t *testing.T) {
	key := wasmTestKey(t)

	_, p := loadWasmTestPlugin(t, tuiWasmSink{})

	// The command is registered in the frontend registry with plugin attribution.
	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered", key)
	}
	if cmd.PluginID != p.ID {
		t.Fatalf("PluginID = %q, want %q (stamped via the WithPlugin scope)", cmd.PluginID, p.ID)
	}

	// registerCommandMenus adapts it into a *CommandMenu provider.
	registerCommandMenus()
	menu, err := BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(%q): %v", key, err)
	}
	cm, ok := menu.(*CommandMenu)
	if !ok {
		t.Fatalf("BuildMenu(%q) = %T, want *CommandMenu", key, menu)
	}
	if got := cm.GetMenuKey(); got != key {
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

	// The WithPlugin attribution records the command key.
	info := pluginInfoSnapshot(t, p.ID)
	if info == nil {
		t.Fatalf("plugin %q not declared", p.ID)
	}
	if !containsString(info.CommandKeys, key) {
		t.Fatalf("CommandKeys = %v, want to contain %q", info.CommandKeys, key)
	}
}

// --- [plugins] disabled gate ---

// TestWasmCommandDisabledGate proves the wasm-derived command carries its
// plugin id so the [plugins] disabled config gates it: commandActionCmd rejects
// before Run when the plugin id is disabled (T4's gate pattern).
func TestWasmCommandDisabledGate(t *testing.T) {
	key := wasmTestKey(t)

	_, p := loadWasmTestPlugin(t, tuiWasmSink{})

	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered", key)
	}
	if cmd.PluginID != p.ID {
		t.Fatalf("PluginID = %q, want %q (the gate keys off this id)", cmd.PluginID, p.ID)
	}

	withPluginConfig(t, []string{p.ID})
	if IsPluginEnabled(p.ID) {
		t.Fatalf("IsPluginEnabled(%q) = true after disabling", p.ID)
	}

	cm := &CommandMenu{BaseMenu: BaseMenu{}, cmd: cmd}
	a := &model.App{} // zero app: Notify is a no-op without a program
	if msg := commandActionCmd(a, cm)(); msg != nil {
		t.Fatalf("commandActionCmd() msg = %v, want nil (rejected by the disabled gate)", msg)
	}
}

// --- Sink Replace semantics (P6) ---

// TestWasmSinkRegisterReplaceSemantics proves tuiWasmSink.RegisterCommands uses
// Replace semantics: a command key already registered is overwritten (no
// duplicate-key panic — the P6 sink 归并), the replacement carries the new
// definition, and the WithPlugin scope still records the key under the
// plugin's CommandKeys.
func TestWasmSinkRegisterReplaceSemantics(t *testing.T) {
	key := wasmTestKey(t)
	RegisterCommand(testCommand(key)) // pre-register so the sink's Replace overwrites it

	err := tuiWasmSink{}.RegisterCommands(&wasm.Plugin{ID: "wasm_sink_replace", Name: "Replace"}, []frontend.Command{
		{Key: key, Title: "Replaced", PluginID: "wasm_sink_replace", Run: func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} }},
	})
	if err != nil {
		t.Fatalf("RegisterCommands returned error: %v", err)
	}
	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered", key)
	}
	if cmd.Title != "Replaced" {
		t.Fatalf("Title = %q, want %q (replaced, not panicked)", cmd.Title, "Replaced")
	}

	info := pluginInfoSnapshot(t, "wasm_sink_replace")
	if info == nil {
		t.Fatalf("plugin %q not declared", "wasm_sink_replace")
	}
	if !containsString(info.CommandKeys, key) {
		t.Fatalf("CommandKeys = %v, want to contain %q", info.CommandKeys, key)
	}
}

// noopSink is a test RegistrySink that drops commands (used to exercise the
// wasm-side CommandsOf/callWasm paths without touching the TUI registries).
type noopSink struct{}

func (noopSink) RegisterCommands(_ *wasm.Plugin, _ []frontend.Command) error { return nil }
