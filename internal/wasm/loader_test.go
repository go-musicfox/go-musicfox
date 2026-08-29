package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDirNonexistentDir(t *testing.T) {
	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close(context.Background())

	errs := m.LoadDir(context.Background(), filepath.Join(t.TempDir(), "missing"))
	if len(errs) != 0 {
		t.Fatalf("want no errors for missing dir, got %v", errs)
	}
	if got := m.Plugins(); len(got) != 0 {
		t.Fatalf("Plugins() = %v, want none", got)
	}
}

// TestLoadDirCollectsErrorsAndLoadsGood builds a temp plugin tree with one
// valid plugin and several broken ones, asserting that a single failing
// plugin does not abort the rest and that all errors are collected.
func TestLoadDirCollectsErrorsAndLoadsGood(t *testing.T) {
	root := t.TempDir()

	// good: builds the example wasm and writes a valid manifest that also
	// pins the correct (uppercase) sha256 to exercise the verification path.
	goodDir := filepath.Join(root, "good")
	if err := os.MkdirAll(goodDir, 0o755); err != nil {
		t.Fatal(err)
	}
	buildExample(t, goodDir)
	goodWasm, err := os.ReadFile(filepath.Join(goodDir, "main.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(goodWasm)
	writeManifest(t, goodDir, fmt.Sprintf(`
id = "good"
name = "Good"
wasm = "main.wasm"
sha256 = "%s"

[[menus]]
key = "wasm_good"
title = "Good"
export = "run"
`, strings.ToUpper(hex.EncodeToString(sum[:]))))

	// nomanifest: no manifest.toml at all.
	if err := os.MkdirAll(filepath.Join(root, "nomanifest"), 0o755); err != nil {
		t.Fatal(err)
	}

	// badsha: manifest sha256 does not match the wasm.
	badshaDir := filepath.Join(root, "badsha")
	if err := os.MkdirAll(badshaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badshaDir, "main.wasm"), goodWasm, 0o644); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, badshaDir, fmt.Sprintf(`
id = "badsha"
name = "BadSha"
sha256 = "%s"

[[menus]]
key = "wasm_badsha"
title = "BadSha"
`, strings.Repeat("0", 64)))

	// nowasm: manifest references a wasm file that does not exist.
	nowasmDir := filepath.Join(root, "nowasm")
	if err := os.MkdirAll(nowasmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, nowasmDir, `
id = "nowasm"
name = "NoWasm"
wasm = "missing.wasm"

[[menus]]
key = "wasm_nowasm"
title = "NoWasm"
`)

	// nomenus: manifest declares no menus.
	nomenusDir := filepath.Join(root, "nomenus")
	if err := os.MkdirAll(nomenusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, nomenusDir, `id = "nomenus" name = "NoMenus"`)

	// badexport: manifest references an export the wasm does not have.
	badexportDir := filepath.Join(root, "badexport")
	if err := os.MkdirAll(badexportDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badexportDir, "main.wasm"), goodWasm, 0o644); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, badexportDir, `
id = "badexport"
name = "BadExport"

[[menus]]
key = "wasm_badexport"
title = "BadExport"
export = "nope"
`)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close(context.Background())

	errs := m.LoadDir(context.Background(), root)
	if len(errs) != 5 {
		t.Fatalf("LoadDir returned %d errors, want 5: %v", len(errs), errs)
	}

	plugins := m.Plugins()
	if len(plugins) != 1 {
		t.Fatalf("Plugins() = %d plugins, want 1 (errors: %v)", len(plugins), errs)
	}
	if plugins[0].ID != "good" {
		t.Fatalf("loaded plugin ID = %q, want %q", plugins[0].ID, "good")
	}
	if _, ok := m.PluginByID("good"); !ok {
		t.Fatal("PluginByID(good) not found")
	}
	if _, ok := m.PluginByID("badsha"); ok {
		t.Fatal("PluginByID(badsha) should not be registered")
	}
}

// TestLoadDirDeterministicOrder asserts plugins are registered in
// directory-name order regardless of on-disk order.
func TestLoadDirDeterministicOrder(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"zebra", "alpha"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeManifest(t, dir, fmt.Sprintf("id = %q\nname = %q\n", name, name))
	}

	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close(context.Background())

	// Both plugins fail at the "no menus" step; the error order must still be
	// deterministic (alphabetical by directory name).
	errs := m.LoadDir(context.Background(), root)
	if len(errs) != 2 {
		t.Fatalf("LoadDir returned %d errors, want 2", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "alpha") {
		t.Fatalf("first error %q does not mention alpha", errs[0])
	}
	if !strings.Contains(errs[1].Error(), "zebra") {
		t.Fatalf("second error %q does not mention zebra", errs[1])
	}
}
