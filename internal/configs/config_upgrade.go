package configs

import (
	"fmt"
	"os"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
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

// tomlTableKeyOrder returns the keys visible at the given tablePath in the
// TOML document, in source (document) order. tablePath is empty for the root
// table. Returns nil on parse error (caller should fall back to sort.Strings).
func tomlTableKeyOrder(data []byte, tablePath []string) []string {
	p := &unstable.Parser{KeepComments: false}
	p.Reset(data)

	var currentTable []string
	var keys []string
	seen := make(map[string]bool)

	for p.NextExpression() {
		e := p.Expression()
		switch e.Kind {
		case unstable.Table, unstable.ArrayTable:
			currentTable = nodeKeyParts(e)
			// A [parent.key] or [a.b.c] header introduces the next key
			// at the parent level (e.g. [reporter.lastfm] introduces
			// "reporter" at root, "lastfm" at ["reporter"]).
			if len(currentTable) > len(tablePath) &&
				stringSlicesEqual(currentTable[:len(tablePath)], tablePath) {
				key := currentTable[len(tablePath)]
				if !seen[key] {
					keys = append(keys, key)
					seen[key] = true
				}
			}
		case unstable.KeyValue:
			if !stringSlicesEqual(currentTable, tablePath) {
				continue
			}
			kps := nodeKeyParts(e)
			if len(kps) == 0 {
				continue
			}
			key := kps[0]
			if !seen[key] {
				keys = append(keys, key)
				seen[key] = true
			}
		}
	}
	if err := p.Error(); err != nil {
		return nil
	}
	return keys
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func nodeKeyParts(e *unstable.Node) []string {
	var parts []string
	it := e.Key()
	for it.Next() {
		parts = append(parts, string(it.Node().Data))
	}
	return parts
}

type tomlAddition struct {
	path            []string
	value           any
	comment         string
	trailingComment string
}

func upgradeTOML(userConfig, defaultConfig []byte) ([]byte, int, error) {
	document, err := tomledit.Parse(userConfig)
	if err != nil {
		return nil, 0, fmt.Errorf("parse user TOML: %w", err)
	}

	defaultDoc, err := tomledit.Parse(defaultConfig)
	if err != nil {
		return nil, 0, fmt.Errorf("parse default TOML: %w", err)
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

	additions := missingTOMLAdditions(defaultConfig, defaultValues, userValues, defaultDoc, nil)
	if len(additions) == 0 {
		return userConfig, 0, nil
	}

	writtenTableComments := make(map[string]bool)

	for _, addition := range additions {
		key := strings.Join(addition.path, ".")
		if err := document.Set(addition.path, tomlEditValue(addition.value)); err != nil {
			return nil, 0, fmt.Errorf("append TOML key %s: %w", key, err)
		}
		if addition.comment != "" {
			if err := document.SetComment(addition.path, addition.comment); err != nil {
				return nil, 0, fmt.Errorf("set comment for %s: %w", key, err)
			}
		}
		if addition.trailingComment != "" {
			if err := document.SetTrailingComment(addition.path, addition.trailingComment); err != nil {
				return nil, 0, fmt.Errorf("set trailing comment for %s: %w", key, err)
			}
		}

		// Write table header comment for newly created tables.
		if len(addition.path) > 1 {
			tablePath := addition.path[:len(addition.path)-1]
			tableKey := strings.Join(tablePath, ".")
			if !writtenTableComments[tableKey] && !tableExists(userValues, tablePath) {
				if comment, ok := defaultDoc.Comment(tablePath); ok && comment != "" {
					if err := document.SetComment(tablePath, comment); err != nil {
						return nil, 0, fmt.Errorf("set comment for table %s: %w", tableKey, err)
					}
					writtenTableComments[tableKey] = true
				}
			}
		}
	}
	return document.Bytes(), len(additions), nil
}

func missingTOMLAdditions(defaultConfig []byte, defaultValues, userValues map[string]any, defaultDoc *tomledit.Document, prefix []string) []tomlAddition {
	orderedKeys := tomlTableKeyOrder(defaultConfig, prefix)
	if len(orderedKeys) == 0 {
		for key := range defaultValues {
			orderedKeys = append(orderedKeys, key)
		}
		sort.Strings(orderedKeys)
	}

	var additions []tomlAddition
	for _, key := range orderedKeys {
		defaultValue, ok := defaultValues[key]
		if !ok {
			continue
		}
		path := append(append([]string(nil), prefix...), key)
		defaultTable, isTable := defaultValue.(map[string]any)
		if !isTable {
			if _, exists := userValues[key]; !exists {
				addition := tomlAddition{path: path, value: defaultValue}
				if comment, ok := defaultDoc.Comment(path); ok {
					addition.comment = comment
				}
				if tc, ok := defaultDoc.TrailingComment(path); ok {
					addition.trailingComment = tc
				}
				additions = append(additions, addition)
			}
			continue
		}

		userValue, exists := userValues[key]
		if exists {
			userTable, isUserTable := userValue.(map[string]any)
			if !isUserTable {
				continue
			}
			additions = append(additions, missingTOMLAdditions(defaultConfig, defaultTable, userTable, defaultDoc, path)...)
			continue
		}
		additions = append(additions, missingTOMLAdditions(defaultConfig, defaultTable, nil, defaultDoc, path)...)
	}
	return additions
}

// tableExists checks whether a nested path exists as a map[string]any in values.
func tableExists(values map[string]any, path []string) bool {
	for _, key := range path {
		v, ok := values[key]
		if !ok {
			return false
		}
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		values = m
	}
	return true
}
