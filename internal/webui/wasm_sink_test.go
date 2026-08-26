package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
)

// wasmTestKey derives a unique command key from the test name so the
// package-global frontend registry stays pollution-free across one test run
// (duplicate keys would panic, same convention as commands_test.go).
func wasmTestKey(t *testing.T) string {
	t.Helper()
	return "webui_wasm_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, t.Name())
}

// wasmTestRepoRoot returns the repository root, derived from this test file's
// path: internal/webui/wasm_sink_test.go -> internal/webui -> internal -> root.
func wasmTestRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// buildWasmExample compiles the examples/wasm/hello plugin with Go's wasip1
// toolchain into dir (as main.wasm). The build is skipped when the toolchain
// is unavailable so CI stays robust.
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

// writeWasmTestPlugin builds a real WASM plugin tree (examples/wasm/hello +
// a manifest whose menu key is derived from the test name and whose args make
// the plugin answer with a toast) under root.
func writeWasmTestPlugin(t *testing.T, root string) {
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
description = "webui sink integration test"

[[menus]]
key = %q
title = "Hello"
after = ""
export = "run"
args = { name = "musicfox", action = "toast" }
`, key)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestWebuiWasmSinkRegistersCommands proves webuiWasmSink forwards the
// CommandsOf-derived commands into the frontend registry with the plugin id
// stamped and key/title/after preserved.
func TestWebuiWasmSinkRegistersCommands(t *testing.T) {
	key := wasmTestKey(t)
	p := &wasm.Plugin{
		ID:   "hello",
		Name: "Hello WASM",
		Menus: []wasm.MenuDecl{
			{Key: key, Title: "Hello", After: "help", Export: "run", Args: map[string]any{"name": "musicfox"}},
		},
	}

	cmds := wasm.CommandsOf(p)
	if err := (webuiWasmSink{}).RegisterCommands(p, cmds); err != nil {
		t.Fatalf("RegisterCommands: %v", err)
	}

	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered", key)
	}
	if cmd.PluginID != p.ID {
		t.Fatalf("PluginID = %q, want %q (stamped by CommandsOf)", cmd.PluginID, p.ID)
	}
	if cmd.Title != "Hello" || cmd.After != "help" {
		t.Fatalf("command = %+v, want title Hello after help", cmd)
	}
	if cmd.Run == nil {
		t.Fatal("Run must be non-nil")
	}
}

// TestWebuiWasmSinkPanicIsolation proves webuiWasmSink.RegisterCommands recovers
// a registration panic (here: a duplicate command key) into a returned error
// instead of crashing the caller.
func TestWebuiWasmSinkPanicIsolation(t *testing.T) {
	key := wasmTestKey(t)
	frontend.RegisterCommand(frontend.Command{
		Key: key,
		Run: func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} },
	})

	err := (webuiWasmSink{}).RegisterCommands(&wasm.Plugin{ID: "wasm_sink_panic", Name: "Panic"}, []frontend.Command{
		{Key: key, Title: "Dup", Run: func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} }},
	})
	if err == nil {
		t.Fatal("RegisterCommands returned nil error, want the recovered duplicate-key panic")
	}
	if !strings.Contains(err.Error(), "wasm_sink_panic") {
		t.Fatalf("error %q does not mention the plugin id", err)
	}
}

// TestWebuiWasmSinkEndToEnd loads a real WASM plugin through the same
// LoadAndRegister -> webuiWasmSink pipeline Run uses and proves the command
// lands in the registry, surfaces on GET /api/commands and executes through
// POST /api/commands/{key} with a toast result.
func TestWebuiWasmSinkEndToEnd(t *testing.T) {
	key := wasmTestKey(t)
	root := t.TempDir()
	writeWasmTestPlugin(t, root)

	mgr, errs := wasm.LoadAndRegister(context.Background(), root, webuiWasmSink{})
	if len(errs) != 0 {
		t.Fatalf("LoadAndRegister returned errors: %v", errs)
	}
	defer mgr.Close(context.Background())

	// Registered in the frontend registry with the manifest id stamped.
	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered", key)
	}
	if cmd.PluginID != "hello" {
		t.Fatalf("PluginID = %q, want %q", cmd.PluginID, "hello")
	}

	// Surfaces on GET /api/commands.
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/commands", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want 200", resp.StatusCode)
	}
	var listBody struct {
		OK   bool          `json:"ok"`
		Data []commandItem `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&listBody)
	var found bool
	for _, item := range listBody.Data {
		if item.Key == key {
			found = true
			if item.PluginID != "hello" {
				t.Fatalf("list item PluginID = %q, want %q", item.PluginID, "hello")
			}
		}
	}
	if !found {
		t.Fatalf("list missing wasm command %q: %+v", key, listBody.Data)
	}

	// Executes through POST /api/commands/{key}.
	execResp := doAuthedPost(t, ts, "/api/commands/"+key, s.token)
	if execResp.StatusCode != http.StatusOK {
		t.Fatalf("exec status = %d, want 200", execResp.StatusCode)
	}
	var execBody struct {
		OK      bool   `json:"ok"`
		Action  string `json:"action"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(execResp.Body).Decode(&execBody)
	if !execBody.OK || execBody.Action != "toast" || execBody.Title != "WASM Hello" {
		t.Fatalf("exec body = %+v", execBody)
	}
	if !strings.Contains(execBody.Message, "musicfox") {
		t.Fatalf("Message %q does not echo the args name", execBody.Message)
	}
}
