package configs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func TestUpgradeTOMLAppendsMissingDefaultKeysWithoutChangingExistingContent(t *testing.T) {
	const defaults = `# Default configuration
[main]
enableMouseEvent = true
frameRate = 5

[main.notification]
inApp = true

[theme]
primaryColor = "#EA403F"

[reporter.lastfm]
enable = false
`
	const user = `# User configuration
[main]
# Keep this comment
enableMouseEvent = false # Keep this trailing comment
customSetting = "untouched"

[theme]
primaryColor = "#123456"
`

	upgraded, added, err := upgradeTOML([]byte(user), []byte(defaults))
	if err != nil {
		t.Fatalf("upgradeTOML() error = %v", err)
	}
	if got, want := added, 3; got != want {
		t.Fatalf("added key count = %d, want %d", got, want)
	}
	content := string(upgraded)
	if !strings.Contains(content, "# Keep this comment\nenableMouseEvent = false # Keep this trailing comment\ncustomSetting = \"untouched\"") {
		t.Fatalf("upgrade changed existing comments or unknown key:\n%s", content)
	}
	for _, expected := range []string{
		"frameRate = 5",
		"[main.notification]\ninApp = true",
		"[reporter.lastfm]\nenable = false",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("upgraded TOML missing %q:\n%s", expected, content)
		}
	}

	var values map[string]any
	if err := toml.Unmarshal(upgraded, &values); err != nil {
		t.Fatalf("upgraded TOML does not parse: %v", err)
	}

	again, addedAgain, err := upgradeTOML(upgraded, []byte(defaults))
	if err != nil {
		t.Fatalf("second upgradeTOML() error = %v", err)
	}
	if addedAgain != 0 {
		t.Fatalf("second added key count = %d, want 0", addedAgain)
	}
	if got := string(again); got != content {
		t.Fatalf("second upgrade changed TOML:\n%s", got)
	}
}

func TestUpgradeConfigUsesEmbeddedDefaultsPreservesContentAndMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	const original = "# User configuration\n[main]\n# Keep this comment\nenableMouseEvent = false # Keep this trailing comment\n"
	if err := os.WriteFile(path, []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}

	added, err := UpgradeConfig(path)
	if err != nil {
		t.Fatalf("UpgradeConfig() error = %v", err)
	}
	if added == 0 {
		t.Fatal("UpgradeConfig() added no built-in defaults")
	}
	upgraded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(upgraded)
	if !strings.Contains(content, "# Keep this comment\nenableMouseEvent = false # Keep this trailing comment") {
		t.Fatalf("UpgradeConfig() changed existing content:\n%s", content)
	}
	if !strings.Contains(content, "[startup]") || !strings.Contains(content, "loadingSeconds = 2") {
		t.Fatalf("UpgradeConfig() did not append built-in startup defaults:\n%s", content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("config mode = %04o, want %04o", got, want)
	}

	again, err := UpgradeConfig(path)
	if err != nil {
		t.Fatalf("second UpgradeConfig() error = %v", err)
	}
	if again != 0 {
		t.Fatalf("second added key count = %d, want 0", again)
	}
	repeated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(repeated); got != content {
		t.Fatalf("second UpgradeConfig() changed TOML:\n%s", got)
	}
}
