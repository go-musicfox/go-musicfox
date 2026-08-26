package configs

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMainHeadlessRoundTrip verifies the [main] headless key round-trips
// through the TOML loader: true is honored, and the embedded default keeps it
// false when the user config does not mention it.
func TestMainHeadlessRoundTrip(t *testing.T) {
	dir := t.TempDir()

	enabledPath := filepath.Join(dir, "headless-enabled.toml")
	if err := os.WriteFile(enabledPath, []byte("[main]\nheadless = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	enabled, err := NewConfigFromTomlFile(enabledPath)
	if err != nil {
		t.Fatalf("NewConfigFromTomlFile() error = %v", err)
	}
	if !enabled.Main.Headless {
		t.Error("Main.Headless = false, want true (round-trip of [main] headless = true)")
	}

	// An empty user file falls back to the embedded default (headless = false).
	emptyPath := filepath.Join(dir, "empty.toml")
	if err := os.WriteFile(emptyPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	defaultCfg, err := NewConfigFromTomlFile(emptyPath)
	if err != nil {
		t.Fatalf("NewConfigFromTomlFile() error = %v", err)
	}
	if defaultCfg.Main.Headless {
		t.Error("Main.Headless = true, want false by default")
	}
}

// TestMainFrontendRoundTrip verifies the [main] frontend key round-trips
// through the TOML loader: a custom value is honored, and the embedded default
// keeps it "tui" when the user config does not mention it.
func TestMainFrontendRoundTrip(t *testing.T) {
	dir := t.TempDir()

	headlessPath := filepath.Join(dir, "frontend-headless.toml")
	if err := os.WriteFile(headlessPath, []byte("[main]\nfrontend = \"headless\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	headlessCfg, err := NewConfigFromTomlFile(headlessPath)
	if err != nil {
		t.Fatalf("NewConfigFromTomlFile() error = %v", err)
	}
	if headlessCfg.Main.Frontend != "headless" {
		t.Errorf("Main.Frontend = %q, want \"headless\" (round-trip of [main] frontend = \"headless\")", headlessCfg.Main.Frontend)
	}

	// An empty user file falls back to the embedded default (frontend = "tui").
	emptyPath := filepath.Join(dir, "empty.toml")
	if err := os.WriteFile(emptyPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	defaultCfg, err := NewConfigFromTomlFile(emptyPath)
	if err != nil {
		t.Fatalf("NewConfigFromTomlFile() error = %v", err)
	}
	if defaultCfg.Main.Frontend != "tui" {
		t.Errorf("Main.Frontend = %q, want \"tui\" by default", defaultCfg.Main.Frontend)
	}
}
