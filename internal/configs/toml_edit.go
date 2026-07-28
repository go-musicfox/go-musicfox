package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pelletier/go-toml/v2/unstable"
	tomledit "github.com/pelletier/go-toml/v2/unstable/edit"
)

// SetTOMLValue updates a TOML value while preserving unrelated layout and comments.
func SetTOMLValue(path string, keyPath []string, value any) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("TOML key path must not be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read TOML file: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat TOML file: %w", err)
	}

	document, err := tomledit.Parse(data)
	if err != nil {
		return fmt.Errorf("parse TOML file: %w", err)
	}
	if err := document.Set(keyPath, tomlEditValue(value)); err != nil {
		return fmt.Errorf("set TOML value: %w", err)
	}

	return replaceTOMLFile(path, document.Bytes(), info.Mode().Perm())
}

func tomlEditValue(value any) any {
	if text, ok := value.(string); ok {
		return unstable.RawMessage(strconv.AppendQuote(nil, text))
	}
	return value
}

func replaceTOMLFile(path string, data []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temporary TOML file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary TOML file mode: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary TOML file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary TOML file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary TOML file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace TOML file: %w", err)
	}
	return nil
}
