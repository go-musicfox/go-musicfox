package envutil

import (
	"os"
	"path/filepath"

	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/strutil"
)

// Getenv get ENV value by key name, can with default value
func Getenv(name string, def ...string) string {
	val := os.Getenv(name)
	if val == "" && len(def) > 0 {
		val = def[0]
	}
	return val
}

// GetInt get int ENV value by key name, can with default value
func GetInt(name string, def ...int) int {
	if val := os.Getenv(name); val != "" {
		return strutil.QuietInt(val)
	}

	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetBool get bool ENV value by key name, can with default value
func GetBool(name string, def ...bool) bool {
	if val := os.Getenv(name); val != "" {
		return strutil.QuietBool(val)
	}

	if len(def) > 0 {
		return def[0]
	}
	return false
}

// GetMulti ENV values by input names.
func GetMulti(names ...string) map[string]string {
	valMap := make(map[string]string, len(names))

	for _, name := range names {
		if val := os.Getenv(name); val != "" {
			valMap[name] = val
		}
	}
	return valMap
}

// EnvPaths get and split $PATH to []string
func EnvPaths() []string {
	return filepath.SplitList(os.Getenv("PATH"))
}

// EnvMap like os.Environ, but will returns key-value map[string]string data.
func EnvMap() map[string]string { return comfunc.Environ() }

// Environ like os.Environ, but will returns key-value map[string]string data.
func Environ() map[string]string { return comfunc.Environ() }

// SearchEnvKeys values by given keywords
func SearchEnvKeys(keywords string) map[string]string {
	return SearchEnv(keywords, false)
}

// SearchEnv values by given keywords
func SearchEnv(keywords string, matchValue bool) map[string]string {
	founded := make(map[string]string)

	for name, val := range comfunc.Environ() {
		if strutil.IContains(name, keywords) {
			founded[name] = val
		} else if matchValue && strutil.IContains(val, keywords) {
			founded[name] = val
		}
	}

	return founded
}
