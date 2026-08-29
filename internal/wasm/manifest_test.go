package wasm

import (
	"strings"
	"testing"
)

func TestParseManifestValid(t *testing.T) {
	data := []byte(`
id = "hello"
name = "Hello WASM"
version = "0.1.0"
author = "you"
description = "示例 WASM 插件"
sha256 = "abcd"
wasm = "main.wasm"

[[menus]]
key = "wasm_hello"
title = "你好 WASM"
after = "daily_songs"
export = "run"
args = { greeting = "hi" }
`)
	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if m.ID != "hello" || m.Name != "Hello WASM" {
		t.Fatalf("unexpected manifest: %+v", m)
	}
	if len(m.Menus) != 1 {
		t.Fatalf("want 1 menu, got %d", len(m.Menus))
	}
	menu := m.Menus[0]
	if menu.Key != "wasm_hello" || menu.Title != "你好 WASM" || menu.After != "daily_songs" || menu.Export != "run" {
		t.Fatalf("unexpected menu: %+v", menu)
	}
	if got, _ := menu.Args["greeting"].(string); got != "hi" {
		t.Fatalf("menu args greeting = %q, want %q", got, "hi")
	}
}

func TestParseManifestInvalid(t *testing.T) {
	if _, err := ParseManifest([]byte("not valid toml =")); err == nil {
		t.Fatal("want parse error for invalid TOML")
	}
}

func TestManifestValidateDefaults(t *testing.T) {
	m := &Manifest{
		ID: "hello",
		Menus: []MenuDecl{
			{Key: "k", Title: "t"},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if m.WasmFile != "main.wasm" {
		t.Fatalf("WasmFile = %q, want %q", m.WasmFile, "main.wasm")
	}
	if m.Menus[0].Export != "run" {
		t.Fatalf("Export = %q, want %q", m.Menus[0].Export, "run")
	}
	if got := m.WasmFileName(); got != "main.wasm" {
		t.Fatalf("WasmFileName() = %q, want %q", got, "main.wasm")
	}
}

func TestManifestValidateEmptyID(t *testing.T) {
	m := &Manifest{Menus: []MenuDecl{{Key: "k", Title: "t"}}}
	err := m.Validate()
	if err == nil || !strings.Contains(err.Error(), "id") {
		t.Fatalf("want id error, got %v", err)
	}
}

func TestManifestValidateNoMenus(t *testing.T) {
	m := &Manifest{ID: "hello"}
	err := m.Validate()
	if err == nil || !strings.Contains(err.Error(), "no menus") {
		t.Fatalf("want no-menus error, got %v", err)
	}
}

func TestManifestValidateEmptyKeyAndTitle(t *testing.T) {
	cases := []struct {
		name string
		menu MenuDecl
		want string
	}{
		{name: "empty key", menu: MenuDecl{Title: "t"}, want: "key"},
		{name: "empty title", menu: MenuDecl{Key: "k"}, want: "title"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := &Manifest{ID: "hello", Menus: []MenuDecl{c.menu}}
			err := m.Validate()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q error, got %v", c.want, err)
			}
		})
	}
}

func TestManifestValidateBadSHA256(t *testing.T) {
	cases := []struct {
		name  string
		sha   string
		valid bool
	}{
		{"empty is fine", "", true},
		{"valid hex", strings.Repeat("ab", 32), true},
		{"uppercase hex", strings.Repeat("AB", 32), true},
		{"too short", "abcd", false},
		{"not hex", strings.Repeat("zz", 32), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := &Manifest{ID: "hello", SHA256: c.sha, Menus: []MenuDecl{{Key: "k", Title: "t"}}}
			err := m.Validate()
			if c.valid && err != nil {
				t.Fatalf("Validate = %v, want nil", err)
			}
			if !c.valid && err == nil {
				t.Fatal("Validate = nil, want error")
			}
		})
	}
}
