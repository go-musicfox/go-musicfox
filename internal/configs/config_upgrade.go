package configs

import (
	"fmt"
	"os"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	tomledit "github.com/pelletier/go-toml/v2/unstable/edit"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/filex"
)

// UpgradeConfig appends built-in default keys that are missing from path while
// preserving the user's existing TOML layout, comments, values, and file mode.
func UpgradeConfig(path string) (int, error) {
	userConfig, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read user TOML file: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("stat user TOML file: %w", err)
	}
	defaultConfig, err := filex.ReadFileFromEmbed("embed/" + types.AppTomlFile)
	if err != nil {
		return 0, fmt.Errorf("read embedded default TOML: %w", err)
	}

	upgraded, added, err := upgradeTOML(userConfig, defaultConfig)
	if err != nil {
		return 0, err
	}
	if added == 0 {
		return 0, nil
	}
	if err := replaceTOMLFile(path, upgraded, info.Mode().Perm()); err != nil {
		return 0, err
	}
	return added, nil
}

type tomlAddition struct {
	path  []string
	value any
}

func upgradeTOML(userConfig, defaultConfig []byte) ([]byte, int, error) {
	document, err := tomledit.Parse(userConfig)
	if err != nil {
		return nil, 0, fmt.Errorf("parse user TOML: %w", err)
	}

	var (
		userValues    map[string]any
		defaultValues map[string]any
	)
	if err := toml.Unmarshal(userConfig, &userValues); err != nil {
		return nil, 0, fmt.Errorf("decode user TOML: %w", err)
	}
	if err := toml.Unmarshal(defaultConfig, &defaultValues); err != nil {
		return nil, 0, fmt.Errorf("decode embedded default TOML: %w", err)
	}

	additions := missingTOMLAdditions(defaultValues, userValues, nil)
	if len(additions) == 0 {
		return userConfig, 0, nil
	}
	for _, addition := range additions {
		if err := document.Set(addition.path, tomlEditValue(addition.value)); err != nil {
			return nil, 0, fmt.Errorf("append TOML key %s: %w", strings.Join(addition.path, "."), err)
		}
	}
	return document.Bytes(), len(additions), nil
}

func missingTOMLAdditions(defaultValues, userValues map[string]any, prefix []string) []tomlAddition {
	keys := make([]string, 0, len(defaultValues))
	for key := range defaultValues {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var additions []tomlAddition
	for _, key := range keys {
		defaultValue := defaultValues[key]
		path := append(append([]string(nil), prefix...), key)
		defaultTable, isTable := defaultValue.(map[string]any)
		if !isTable {
			if _, exists := userValues[key]; !exists {
				additions = append(additions, tomlAddition{path: path, value: defaultValue})
			}
			continue
		}

		userValue, exists := userValues[key]
		if exists {
			userTable, isUserTable := userValue.(map[string]any)
			if !isUserTable {
				continue
			}
			additions = append(additions, missingTOMLAdditions(defaultTable, userTable, path)...)
			continue
		}
		additions = append(additions, missingTOMLAdditions(defaultTable, nil, path)...)
	}
	return additions
}
