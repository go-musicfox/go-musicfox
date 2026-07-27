package configs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/utils/filex"
)

// BuiltinThemesDir is the embedded directory path for built-in themes.
const BuiltinThemesDir = "embed/themes"

// UserThemesSubDir is the subdirectory under the user config directory for custom themes.
const UserThemesSubDir = "themes"

// LoadAllThemes loads both built-in and user themes, merging them into a single map.
// User themes override built-in themes with the same name.
// Returns a map of theme name → ThemeFile.
func LoadAllThemes(userConfigDir string) map[string]*ThemeFile {
	themes := make(map[string]*ThemeFile)

	// 1. Load built-in themes from embed
	builtin := loadBuiltinThemes()
	for name, tf := range builtin {
		themes[name] = tf
		slog.Debug("loaded built-in theme", "name", name)
	}

	// 2. Load user themes from filesystem (override built-in)
	if userConfigDir == "" {
		userConfigDir = defaultUserConfigDir()
	}
	userThemesDir := filepath.Join(userConfigDir, UserThemesSubDir)
	user := loadUserThemes(userThemesDir)
	for name, tf := range user {
		themes[name] = tf
		slog.Debug("loaded user theme (overriding)", "name", name)
	}

	if len(themes) == 0 {
		slog.Warn("no themes loaded, falling back to built-in default")
		themes["Default"] = defaultBuiltinTheme()
	}

	return themes
}

// loadBuiltinThemes loads all .toml files from the embedded themes directory.
func loadBuiltinThemes() map[string]*ThemeFile {
	themes := make(map[string]*ThemeFile)

	entries, err := filex.ReadDirFromEmbed(BuiltinThemesDir)
	if err != nil {
		slog.Debug("no built-in themes directory", "err", err)
		return themes
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		data, err := filex.ReadFileFromEmbed(filepath.Join(BuiltinThemesDir, entry.Name()))
		if err != nil {
			slog.Warn("failed to read built-in theme file", "file", entry.Name(), "err", err)
			continue
		}
		tf, err := parseThemeFile(string(data))
		if err != nil {
			slog.Warn("failed to parse built-in theme", "file", entry.Name(), "err", err)
			continue
		}
		themes[tf.Name] = tf
	}
	return themes
}

// loadUserThemes loads all .toml files from the user themes directory.
func loadUserThemes(dir string) map[string]*ThemeFile {
	themes := make(map[string]*ThemeFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Debug("no user themes directory", "dir", dir, "err", err)
		return themes
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("failed to read user theme file", "file", path, "err", err)
			continue
		}
		tf, err := parseThemeFile(string(data))
		if err != nil {
			slog.Warn("failed to parse user theme", "file", path, "err", err)
			continue
		}
		themes[tf.Name] = tf
	}
	return themes
}

// parseThemeFile parses a TOML theme definition.
func parseThemeFile(content string) (*ThemeFile, error) {
	var tf ThemeFile
	// TOML uses '.' as key separator, which works well with our flat key structure.
	if _, err := toml.Decode(content, &tf); err != nil {
		return nil, fmt.Errorf("parse theme: %w", err)
	}
	if tf.Name == "" {
		return nil, fmt.Errorf("theme file missing 'name' field")
	}
	return &tf, nil
}

// BuildThemeList converts a map of theme files into a flat []style.Theme slice
// for foxful-cli's ThemeList, and returns the sorted list of names for UI display.
// Only themes matching the target brightness (dark/light) are included.
func BuildThemeList(themes map[string]*ThemeFile, darkBackground bool) ([]style.Theme, []string) {
	// Sort names for deterministic order
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	sort.Strings(names)

	themeList := make([]style.Theme, 0, len(names))
	themeNames := make([]string, 0, len(names))

	for _, name := range names {
		tf := themes[name]
		var variant ThemeVariant
		if darkBackground {
			variant = tf.Dark
		} else {
			variant = tf.Light
		}
		// Skip themes that don't have this variant configured
		if !variant.isConfigured() {
			continue
		}
		themeList = append(themeList, variant.toTheme())
		themeNames = append(themeNames, name)
	}

	return themeList, themeNames
}

// isConfigured returns true if the variant has at least a primary color set.
func (v ThemeVariant) isConfigured() bool {
	return v.Primary != "" || v.Foreground != "" || v.Background != ""
}

// defaultUserConfigDir returns the default user config directory.
func defaultUserConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "go-musicfox")
}

// defaultBuiltinTheme returns a fallback theme built from code defaults.
func defaultBuiltinTheme() *ThemeFile {
	return &ThemeFile{
		Name:        "Default",
		Description: "Default classic NetEase Music red theme (fallback)",
		Dark: ThemeVariant{
			Primary: "#EA403F",
		},
		Light: ThemeVariant{
			Primary: "#EA403F",
		},
	}
}
