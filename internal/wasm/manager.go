package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	wazero "github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// memoryLimitPages caps each plugin instance's linear memory at 128 pages
// (8 MiB). Configurable later.
const memoryLimitPages = 128

// Manager discovers, loads and owns WASM plugin instances. Loading is
// centralized and single-threaded: wazero recommends not compiling the same
// wasm concurrently. The Manager holds no global state; it is owned by
// whoever calls LoadDir.
type Manager struct {
	runtime    wazero.Runtime
	plugins    []*Plugin
	pluginByID map[string]*Plugin
}

// NewManager creates a Manager with a wazero compiler runtime hardened with a
// memory limit and close-on-context-done, and instantiates WASI once (Go
// wasip1 guests link it).
func NewManager() (*Manager, error) {
	ctx := context.Background()
	config := wazero.NewRuntimeConfigCompiler().
		WithMemoryLimitPages(memoryLimitPages).
		WithCloseOnContextDone(true)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	wasi.MustInstantiate(ctx, r)
	return &Manager{
		runtime:    r,
		pluginByID: make(map[string]*Plugin),
	}, nil
}

// LoadDir scans dir for plugin directories and loads each of them. A missing
// dir is not an error. A single failing plugin does not stop the others; all
// errors are collected and returned. Plugins are registered in directory-name
// order for deterministic registration.
func (m *Manager) LoadDir(ctx context.Context, dir string) []error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []error{fmt.Errorf("wasm: read plugin dir %s: %w", dir, err)}
	}

	// Deterministic registration order: process directories sorted by name.
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)

	var errs []error
	for _, name := range dirs {
		p, loadErr := m.loadPlugin(ctx, filepath.Join(dir, name))
		if loadErr != nil {
			errs = append(errs, fmt.Errorf("wasm: plugin %q: %w", name, loadErr))
			continue
		}
		m.plugins = append(m.plugins, p)
		m.pluginByID[p.ID] = p
	}
	return errs
}

// loadPlugin loads a single plugin directory: reads and validates the
// manifest, verifies the wasm SHA-256 when configured, then instantiates it.
func (m *Manager) loadPlugin(ctx context.Context, dir string) (*Plugin, error) {
	manifestPath := filepath.Join(dir, "manifest.toml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", manifestPath, err)
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return nil, err
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}

	wasmBytes, err := os.ReadFile(filepath.Join(dir, manifest.WasmFileName()))
	if err != nil {
		return nil, fmt.Errorf("read wasm file %q: %w", manifest.WasmFileName(), err)
	}

	if manifest.SHA256 != "" {
		sum := sha256.Sum256(wasmBytes)
		got := hex.EncodeToString(sum[:])
		if !strings.EqualFold(got, manifest.SHA256) {
			return nil, fmt.Errorf("sha256 mismatch: got %s want %s", got, manifest.SHA256)
		}
	}

	return newPlugin(m.runtime, manifest, wasmBytes)
}

// Plugins returns a snapshot copy of the loaded plugins.
func (m *Manager) Plugins() []*Plugin {
	cp := make([]*Plugin, len(m.plugins))
	copy(cp, m.plugins)
	return cp
}

// PluginByID returns the plugin registered under the given manifest id.
func (m *Manager) PluginByID(id string) (*Plugin, bool) {
	p, ok := m.pluginByID[id]
	return p, ok
}

// Close closes all plugins (in reverse load order) and then the runtime.
func (m *Manager) Close(ctx context.Context) {
	for i := len(m.plugins) - 1; i >= 0; i-- {
		m.plugins[i].Close(ctx)
	}
	m.plugins = nil
	m.pluginByID = make(map[string]*Plugin)
	_ = m.runtime.Close(ctx)
}
