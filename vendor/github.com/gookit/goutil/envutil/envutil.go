// Package envutil provide some commonly ENV util functions.
package envutil

import (
	"os"

	"github.com/gookit/goutil/internal/comfunc"
)

// ValueGetter Env value provider func.
//
// TIPS: you can custom provide data.
var ValueGetter = os.Getenv

// VarReplace replaces ${var} or $var in the string according to the values.
//
// is alias of the os.ExpandEnv()
func VarReplace(s string) string { return os.ExpandEnv(s) }

// VarParse alias of the ParseValue
func VarParse(val string) string {
	return comfunc.ParseEnvVar(val, ValueGetter)
}

// ParseEnvValue alias of the ParseValue
func ParseEnvValue(val string) string {
	return comfunc.ParseEnvVar(val, ValueGetter)
}

// ParseValue parse ENV var value from input string, support default value.
//
// Format:
//
//	${var_name}            Only var name
//	${var_name | default}  With default value
//
// Usage:
//
//	envutil.ParseValue("${ APP_NAME }")
//	envutil.ParseValue("${ APP_ENV | dev }")
func ParseValue(val string) (newVal string) {
	return comfunc.ParseEnvVar(val, ValueGetter)
}

// SetEnvMap set multi ENV(string-map) to os
func SetEnvMap(mp map[string]string) {
	for key, value := range mp {
		_ = os.Setenv(key, value)
	}
}

// SetEnvs set multi k-v ENV pairs to os
func SetEnvs(kvPairs ...string) {
	if len(kvPairs)%2 == 1 {
		panic("envutil.SetEnvs: odd argument count")
	}

	for i := 0; i < len(kvPairs); i += 2 {
		_ = os.Setenv(kvPairs[i], kvPairs[i+1])
	}
}

// UnsetEnvs from os
func UnsetEnvs(keys ...string) {
	for _, key := range keys {
		_ = os.Unsetenv(key)
	}
}
