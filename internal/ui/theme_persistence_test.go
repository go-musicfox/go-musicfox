package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

func TestWriteActiveThemePersistsNextStartupSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	const original = "# User configuration\n[theme]\n# Preserve this comment\nactiveTheme = \"Default\" # Keep this trailing comment\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := writeActiveTheme(path, "Nord"); err != nil {
		t.Fatalf("writeActiveTheme() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "# Preserve this comment\nactiveTheme = \"Nord\" # Keep this trailing comment") {
		t.Fatalf("theme edit lost comments or value:\n%s", content)
	}

	config, err := configs.NewConfigFromTomlFile(path)
	if err != nil {
		t.Fatalf("NewConfigFromTomlFile() error = %v", err)
	}
	if got, want := config.Theme.ActiveTheme, "Nord"; got != want {
		t.Fatalf("next startup active theme = %q, want %q", got, want)
	}
}
