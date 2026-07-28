package configs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetTOMLValuePreservesLayoutAndFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	const original = "# User configuration\n[main]\n# Keep this comment\nenableMouseEvent = true # Keep this trailing comment\n\n[theme]\nactiveTheme = \"Default\"\n"
	if err := os.WriteFile(path, []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := SetTOMLValue(path, []string{"main", "enableMouseEvent"}, false); err != nil {
		t.Fatalf("SetTOMLValue() error = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(updated)
	if !strings.Contains(content, "# Keep this comment\nenableMouseEvent = false # Keep this trailing comment") {
		t.Fatalf("edited TOML lost target comments or value:\n%s", content)
	}
	if !strings.Contains(content, "[theme]\nactiveTheme = \"Default\"") {
		t.Fatalf("edited TOML changed unrelated table:\n%s", content)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("config mode = %04o, want %04o", got, want)
	}
}

func TestSetTOMLValueSetsStringAtEndOfDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	const original = "[theme]\nactiveTheme = \"Default\" # Keep this trailing comment\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := SetTOMLValue(path, []string{"theme", "activeTheme"}, "Nord"); err != nil {
		t.Fatalf("SetTOMLValue() error = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(updated), "[theme]\nactiveTheme = \"Nord\" # Keep this trailing comment\n"; got != want {
		t.Fatalf("edited TOML = %q, want %q", got, want)
	}
}
