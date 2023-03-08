package envutil

import (
	"os"

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

// Environ like os.Environ, but will returns key-value map[string]string data.
func Environ() map[string]string {
	return comfunc.Environ()
}
