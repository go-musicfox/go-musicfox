package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	wazero "github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// repoRoot returns the repository root, derived from this test file's path:
// internal/wasm/plugin_test.go -> internal/wasm -> internal -> repo root.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// buildExample compiles the examples/wasm/hello plugin with Go's wasip1
// toolchain into outDir (as main.wasm). The build is skipped when the
// toolchain is unavailable so CI stays robust.
func buildExample(t *testing.T, outDir string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", filepath.Join(outDir, "main.wasm"), "./examples/wasm/hello")
	cmd.Dir = repoRoot(t)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("go wasm build unavailable: %v (%s)", err, out)
	}
}

// writeManifest writes manifest.toml into dir.
func writeManifest(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "manifest.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const helloManifest = `
id = "hello"
name = "Hello WASM"
version = "0.1.0"
author = "test"
description = "integration test plugin"

[[menus]]
key = "wasm_hello"
title = "Hello"
after = ""
export = "run"
`

// TestPluginRunIntegration builds the example plugin, loads it through the
// Manager and exercises the full call protocol: request marshalling, memory
// allocation, export dispatch and response decoding.
func TestPluginRunIntegration(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	buildExample(t, pluginDir)
	writeManifest(t, pluginDir, helloManifest)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
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
	if p.ID != "hello" || p.Name != "Hello WASM" {
		t.Fatalf("unexpected plugin: ID=%q Name=%q", p.ID, p.Name)
	}
	if got, ok := m.PluginByID("hello"); !ok || got != p {
		t.Fatalf("PluginByID(hello) = %v, %v; want plugin, true", got, ok)
	}
	if len(p.Menus) != 1 || p.Menus[0].Key != "wasm_hello" {
		t.Fatalf("unexpected menus: %+v", p.Menus)
	}

	ctx := context.Background()
	req := Request{
		Version: ProtocolVersion,
		Action:  "wasm_hello",
		Args:    map[string]any{"name": "musicfox"},
		Context: RequestContext{
			UserID:   123,
			UserName: "musicfox",
			Playing:  true,
			Song:     &SongInfo{ID: 1, Name: "Song", Artist: "Artist", Album: "Album"},
		},
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	out, err := p.Run(ctx, p.Menus[0], reqJSON)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("unmarshal response %q: %v", out, err)
	}
	if resp.Action != "view" {
		t.Fatalf("Action = %q, want %q", resp.Action, "view")
	}
	if !strings.Contains(resp.Message, "musicfox") {
		t.Fatalf("message %q does not echo args name", resp.Message)
	}
	if !strings.Contains(resp.Message, "Song") || !strings.Contains(resp.Message, "Artist") {
		t.Fatalf("message %q does not include the context song", resp.Message)
	}
	if !resp.LevelValid() {
		t.Fatalf("level %q is not valid", resp.Level)
	}

	// A second Run must work: the instance survives between calls.
	if _, err := p.Run(ctx, p.Menus[0], reqJSON); err != nil {
		t.Fatalf("second Run: %v", err)
	}

	// A request asking for a toast must switch the response action.
	req.Args = map[string]any{"name": "musicfox", "action": "toast"}
	reqJSON, err = json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	out, err = p.Run(ctx, p.Menus[0], reqJSON)
	if err != nil {
		t.Fatalf("Run toast: %v", err)
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("unmarshal response %q: %v", out, err)
	}
	if resp.Action != "toast" {
		t.Fatalf("Action = %q, want %q", resp.Action, "toast")
	}
}

// TestPluginRunUnknownExport checks that Run rejects a MenuDecl whose export
// was not resolved at load time.
func TestPluginRunUnknownExport(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	buildExample(t, pluginDir)
	writeManifest(t, pluginDir, helloManifest)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close(context.Background())

	if errs := m.LoadDir(context.Background(), root); len(errs) != 0 {
		t.Fatalf("unexpected load errors: %v", errs)
	}
	p := m.Plugins()[0]

	_, err = p.Run(context.Background(), MenuDecl{Key: "nope", Title: "Nope", Export: "nope"}, []byte(`{}`))
	if err == nil {
		t.Fatal("Run with unknown export returned nil error")
	}
}

// TestPluginRunTimeout verifies the watchdog: a hung export is interrupted
// after the plugin's timeout, ErrTimeout is returned and the instance is
// closed for subsequent calls.
func TestPluginRunTimeout(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "hang")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	buildExample(t, pluginDir)

	wasmBytes, err := os.ReadFile(filepath.Join(pluginDir, "main.wasm"))
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}

	// Construct the plugin directly with a short watchdog timeout, bypassing
	// the Manager's DefaultTimeout.
	ctx := context.Background()
	config := wazero.NewRuntimeConfigCompiler().
		WithMemoryLimitPages(memoryLimitPages).
		WithCloseOnContextDone(true)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	wasi.MustInstantiate(ctx, r)
	defer r.Close(ctx)

	manifest := &Manifest{
		ID:       "hang",
		Name:     "Hang",
		WasmFile: "main.wasm",
		Menus:    []MenuDecl{{Key: "wasm_hang", Title: "Hang", Export: "hang"}},
	}
	p, err := newPluginWithTimeout(r, manifest, wasmBytes, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("newPluginWithTimeout: %v", err)
	}
	defer p.Close(ctx)

	start := time.Now()
	_, err = p.Run(ctx, p.Menus[0], []byte(`{}`))
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Run = %v, want ErrTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("timeout watchdog took %v, want ~200ms", elapsed)
	}

	// The instance is closed after a timeout; further calls must fail fast.
	if _, err := p.Run(ctx, p.Menus[0], []byte(`{}`)); !errors.Is(err, ErrPluginClosed) {
		t.Fatalf("Run after timeout = %v, want ErrPluginClosed", err)
	}
}
