package wasm

import (
	"encoding/hex"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// A plugin is a directory containing manifest.toml and a wasm reactor:
//
//	<plugin-dir>/
//	  manifest.toml
//	  main.wasm
//
// Example manifest.toml:
//
//	id = "hello"
//	name = "Hello WASM"
//	version = "0.1.0"
//	author = "you"
//	description = "示例 WASM 插件"
//	sha256 = ""          # optional hex SHA-256 of the wasm file; verified when non-empty
//	wasm = "main.wasm"   # optional, defaults to "main.wasm"
//
//	[[menus]]
//	key = "wasm_hello"   # globally unique menu registry key
//	title = "你好 WASM"   # main-menu item title
//	after = ""           # optional main-menu after-anchor key ("" = append at end)
//	export = "run"       # wasm export to call, defaults "run"
//	args = {}            # optional static args passed to the plugin

// Manifest is the parsed manifest.toml of a WASM plugin.
type Manifest struct {
	ID          string     `toml:"id"`
	Name        string     `toml:"name"`
	Version     string     `toml:"version"`
	Author      string     `toml:"author"`
	Description string     `toml:"description"`
	SHA256      string     `toml:"sha256"`
	WasmFile    string     `toml:"wasm"`
	Menus       []MenuDecl `toml:"menus"`
}

// MenuDecl declares one main-menu entry backed by a plugin export.
type MenuDecl struct {
	Key    string         `toml:"key"`
	Title  string         `toml:"title"`
	After  string         `toml:"after"`
	Export string         `toml:"export"`
	Args   map[string]any `toml:"args"`
}

// ParseManifest decodes a manifest.toml document.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("wasm: parse manifest: %w", err)
	}
	return &m, nil
}

// Validate fills in defaults and checks required fields: a non-empty ID, at
// least one menu, and a non-empty Key and Title per menu. WasmFile defaults to
// "main.wasm" and Export to "run". A non-empty SHA256 must be valid hex.
func (m *Manifest) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("wasm: manifest id is required")
	}
	if len(m.Menus) == 0 {
		return fmt.Errorf("wasm: manifest %q declares no menus", m.ID)
	}
	for i, menu := range m.Menus {
		if menu.Key == "" {
			return fmt.Errorf("wasm: manifest %q menu %d: key is required", m.ID, i)
		}
		if menu.Title == "" {
			return fmt.Errorf("wasm: manifest %q menu %q: title is required", m.ID, menu.Key)
		}
	}

	if m.WasmFile == "" {
		m.WasmFile = "main.wasm"
	}
	for i := range m.Menus {
		if m.Menus[i].Export == "" {
			m.Menus[i].Export = "run"
		}
	}

	if m.SHA256 != "" {
		if len(m.SHA256) != 64 {
			return fmt.Errorf("wasm: manifest %q: sha256 must be a 64-char hex string", m.ID)
		}
		if _, err := hex.DecodeString(m.SHA256); err != nil {
			return fmt.Errorf("wasm: manifest %q: invalid sha256 hex: %w", m.ID, err)
		}
	}
	return nil
}

// WasmFileName returns the wasm file name relative to the plugin directory.
// It is safe to call before Validate and returns "main.wasm" when unset.
func (m *Manifest) WasmFileName() string {
	if m.WasmFile == "" {
		return "main.wasm"
	}
	return m.WasmFile
}
