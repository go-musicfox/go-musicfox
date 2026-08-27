package wasm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// cordisTestKey derives a unique frontend command key from the test name so the
// package-global frontend registry stays pollution-free across a test run.
func cordisTestKey(t *testing.T) string {
	t.Helper()
	return "cordis_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, t.Name())
}

// sinkCall records one RegisterCommands invocation.
type sinkCall struct {
	p    *Plugin
	cmds []frontend.Command
}

// cordisSink registers commands into the frontend registry with the same
// Replace semantics as the production sinks and records the plugin/command
// pairs it saw.
type cordisSink struct {
	mu    sync.Mutex
	calls []sinkCall
}

func (s *cordisSink) RegisterCommands(p *Plugin, cmds []frontend.Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, sinkCall{p: p, cmds: cmds})
	for _, c := range cmds {
		frontend.ReplaceCommand(c)
	}
	return nil
}

// writeCordisPlugin writes a valid plugin directory (example wasm + manifest
// whose menu key derives from the test name) under root.
func writeCordisPlugin(t *testing.T, root, id string) {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	buildExample(t, dir)
	writeManifest(t, dir, fmt.Sprintf(`
id = %q
name = "Cordis %s"
version = "0.1.0"
author = "test"
description = "cordis scope integration test"

[[menus]]
key = %q
title = "Hello"
after = ""
export = "run"
`, id, id, cordisTestKey(t)))
}

// TestManagerPluginStartProvidesManagerAndDisposeCloses proves the root-scope
// plugin contract: Start creates the manager and provides it under
// ServiceWasmManager, and Dispose closes it (plugins cleared).
func TestManagerPluginStartProvidesManagerAndDisposeCloses(t *testing.T) {
	root := t.TempDir()
	writeCordisPlugin(t, root, "hello")

	ctx := &framework.Context{}
	scope := framework.NewScope()
	mp := &ManagerPlugin{}
	if err := scope.Add(mp); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := scope.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	mgr, ok := framework.ServiceOf[*Manager](ctx, ServiceWasmManager)
	if !ok {
		t.Fatal("ServiceWasmManager not resolvable after Start")
	}
	if mgr != mp.mgr {
		t.Fatal("resolved manager is not the plugin-owned manager")
	}

	// Load a plugin through the manager so Dispose has something to close.
	if errs := mgr.LoadDir(context.Background(), root); len(errs) != 0 {
		t.Fatalf("LoadDir: %v", errs)
	}
	if got := len(mgr.Plugins()); got != 1 {
		t.Fatalf("Plugins() = %d, want 1", got)
	}

	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose: %v", err)
	}
	if got := len(mgr.Plugins()); got != 0 {
		t.Fatalf("Plugins() after Dispose = %d, want 0 (manager closed)", got)
	}
}

// TestWasmPluginStartStopDispose proves the per-directory adapter lifecycle:
// Start registers the commands (Replace), Stop unregisters them and a re-Start
// after a menu change refreshes the command definition (the "Stop 后 Unregister
// + 重新 Start 后 Replace 生效" hot-reload core), and Dispose closes the
// instance.
func TestWasmPluginStartStopDispose(t *testing.T) {
	key := cordisTestKey(t)
	root := t.TempDir()
	writeCordisPlugin(t, root, "hello")

	ctx := &framework.Context{}
	scope := framework.NewScope()
	if err := scope.Add(&ManagerPlugin{}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := scope.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = scope.Dispose() })

	mgr, _ := framework.ServiceOf[*Manager](ctx, ServiceWasmManager)
	if errs := mgr.LoadDir(context.Background(), root); len(errs) != 0 {
		t.Fatalf("LoadDir: %v", errs)
	}
	p := mgr.Plugins()[0]

	sink := &cordisSink{}
	wp := &wasmPlugin{p: p, sink: sink}
	if err := scope.AddAndStart(ctx, wp); err != nil {
		t.Fatalf("AddAndStart: %v", err)
	}

	// Start registered the command.
	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered after Start", key)
	}
	if cmd.PluginID != p.ID {
		t.Fatalf("PluginID = %q, want %q", cmd.PluginID, p.ID)
	}
	if cmd.Title != "Hello" {
		t.Fatalf("Title = %q, want %q", cmd.Title, "Hello")
	}

	// Stop unregisters.
	if err := wp.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if _, ok := frontend.CommandByKey(key); ok {
		t.Fatalf("command %q still registered after Stop", key)
	}

	// "目录替换" semantics at the adapter level: the manifest changed the
	// menu title; a re-Start Replace-updates the registry (no duplicate-key
	// panic, position kept).
	p.Menus[0].Title = "Hello v2"
	if err := wp.Start(ctx); err != nil {
		t.Fatalf("re-Start: %v", err)
	}
	cmd, ok = frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered after re-Start", key)
	}
	if cmd.Title != "Hello v2" {
		t.Fatalf("Title after re-Start = %q, want %q (Replace must refresh the definition)", cmd.Title, "Hello v2")
	}

	// Dispose closes the instance: Run on a closed plugin fails fast.
	if err := wp.Dispose(); err != nil {
		t.Fatalf("Dispose: %v", err)
	}
	if _, err := p.Run(context.Background(), p.Menus[0], []byte(`{}`)); err == nil {
		t.Fatal("Run on a disposed plugin succeeded, want ErrPluginClosed")
	}
}

// TestLoadIntoScopeAddThenStart proves the "先 Add 后统一 Start" path: on a
// not-yet-started scope LoadIntoScope only registers the adapters, and the
// scope Start brings up the manager + plugins together, after which the
// commands are registered and the scope Dispose unregisters + closes
// everything.
func TestLoadIntoScopeAddThenStart(t *testing.T) {
	key := cordisTestKey(t)
	root := t.TempDir()
	writeCordisPlugin(t, root, "hello")

	scope := framework.NewScope()
	fctx := &framework.Context{}
	mgr, errs := LoadIntoScope(context.Background(), fctx, scope, root, &cordisSink{})
	if len(errs) != 0 {
		t.Fatalf("LoadIntoScope errors: %v", errs)
	}
	if mgr == nil {
		t.Fatal("LoadIntoScope returned nil manager")
	}
	if got := len(mgr.Plugins()); got != 1 {
		t.Fatalf("manager has %d plugins, want 1", got)
	}
	if _, ok := frontend.CommandByKey(key); ok {
		t.Fatal("command registered before scope Start")
	}

	if err := scope.Start(fctx); err != nil {
		t.Fatalf("scope Start: %v", err)
	}
	if _, ok := frontend.CommandByKey(key); !ok {
		t.Fatal("command not registered after scope Start")
	}

	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose: %v", err)
	}
	if _, ok := frontend.CommandByKey(key); ok {
		t.Fatal("command still registered after scope Dispose")
	}
}

// TestLoadIntoScopeOnStartedScope proves the "scope 已 Start → AddAndStart"
// path (the TUI's load order: the wasm sub-scope starts with the frontend
// scope, then loadWasmPlugins AddAndStarts the adapters): the command is
// already registered right after LoadIntoScope, without an extra Start.
func TestLoadIntoScopeOnStartedScope(t *testing.T) {
	key := cordisTestKey(t)
	root := t.TempDir()
	writeCordisPlugin(t, root, "hello")

	scope := framework.NewScope()
	fctx := &framework.Context{}
	if err := scope.Add(&ManagerPlugin{}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := scope.Start(fctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if _, errs := LoadIntoScope(context.Background(), fctx, scope, root, &cordisSink{}); len(errs) != 0 {
		t.Fatalf("LoadIntoScope errors: %v", errs)
	}
	if _, ok := frontend.CommandByKey(key); !ok {
		t.Fatal("command not registered on a started scope (AddAndStart must have run)")
	}

	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose: %v", err)
	}
	if _, ok := frontend.CommandByKey(key); ok {
		t.Fatal("command still registered after scope Dispose")
	}
}

// TestLoadIntoScopeFailureIsolation proves a single broken plugin directory
// neither stops the other plugins nor fails the whole load: errors are
// collected and returned, the good plugin still loads and starts.
func TestLoadIntoScopeFailureIsolation(t *testing.T) {
	key := cordisTestKey(t)
	root := t.TempDir()
	writeCordisPlugin(t, root, "good")
	// bad: a manifest whose referenced wasm file does not exist.
	badDir := filepath.Join(root, "bad")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, badDir, `
id = "bad"
name = "Bad"
wasm = "missing.wasm"

[[menus]]
key = "cordis_bad"
title = "Bad"
`)

	scope := framework.NewScope()
	fctx := &framework.Context{}
	mgr, errs := LoadIntoScope(context.Background(), fctx, scope, root, &cordisSink{})
	if len(errs) != 1 {
		t.Fatalf("LoadIntoScope returned %d errors, want 1 (bad plugin only): %v", len(errs), errs)
	}
	if mgr == nil {
		t.Fatal("LoadIntoScope returned nil manager")
	}
	if got := len(mgr.Plugins()); got != 1 {
		t.Fatalf("manager has %d plugins, want 1 (only the good one)", got)
	}

	if err := scope.Start(fctx); err != nil {
		t.Fatalf("scope Start: %v", err)
	}
	if _, ok := frontend.CommandByKey(key); !ok {
		t.Fatal("good plugin's command not registered after Start")
	}

	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose: %v", err)
	}
}

// TestLoadIntoScopeReloadRefreshesCommands proves the P6 hot-reload capability:
// "目录替换 → 重新 LoadIntoScope → 命令集刷新". Generation 1 loads a plugin
// whose command title is "Hello"; the directory is replaced with a new
// generation whose manifest changes the title to "Hello v2"; a fresh scope
// (new manager generation) loads the new directory and the registry now serves
// the refreshed definition, while disposing generation 1 unregistered its
// commands.
func TestLoadIntoScopeReloadRefreshesCommands(t *testing.T) {
	key := cordisTestKey(t)
	root := t.TempDir()

	writeCordisPlugin(t, root, "hello")
	manifestPath := filepath.Join(root, "hello", "manifest.toml")

	// Generation 1.
	scope1 := framework.NewScope()
	fctx1 := &framework.Context{}
	if _, errs := LoadIntoScope(context.Background(), fctx1, scope1, root, &cordisSink{}); len(errs) != 0 {
		t.Fatalf("gen1 LoadIntoScope errors: %v", errs)
	}
	if err := scope1.Start(fctx1); err != nil {
		t.Fatalf("gen1 scope Start: %v", err)
	}
	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("gen1 command %q not registered", key)
	}
	if cmd.Title != "Hello" {
		t.Fatalf("gen1 Title = %q, want %q", cmd.Title, "Hello")
	}

	// 目录替换: swap the manifest title, then dispose the old generation
	// (Stop unregisters the old command definition).
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte(strings.Replace(string(data), `title = "Hello"`, `title = "Hello v2"`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scope1.Dispose(); err != nil {
		t.Fatalf("gen1 Dispose: %v", err)
	}
	if _, ok := frontend.CommandByKey(key); ok {
		t.Fatal("gen1 command still registered after scope1 Dispose")
	}

	// Generation 2: 重新 LoadIntoScope → 命令集刷新.
	scope2 := framework.NewScope()
	fctx2 := &framework.Context{}
	if _, errs := LoadIntoScope(context.Background(), fctx2, scope2, root, &cordisSink{}); len(errs) != 0 {
		t.Fatalf("gen2 LoadIntoScope errors: %v", errs)
	}
	if err := scope2.Start(fctx2); err != nil {
		t.Fatalf("gen2 scope Start: %v", err)
	}
	cmd, ok = frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("gen2 command %q not registered", key)
	}
	if cmd.Title != "Hello v2" {
		t.Fatalf("gen2 Title = %q, want %q (command set must be refreshed)", cmd.Title, "Hello v2")
	}

	if err := scope2.Dispose(); err != nil {
		t.Fatalf("gen2 Dispose: %v", err)
	}
}
