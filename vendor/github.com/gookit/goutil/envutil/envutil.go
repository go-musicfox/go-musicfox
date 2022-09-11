package envutil

import (
	"os"
	"regexp"
	"strings"
)

// VarReplace replaces ${var} or $var in the string according to the values.
// is alias of the os.ExpandEnv()
func VarReplace(s string) string {
	return os.ExpandEnv(s)
}

// ValueGetter Env value provider func.
// TIPS: you can custom provide data.
var ValueGetter = os.Getenv

// parse env value, allow:
// 	only key 	 - "${SHELL}"
// 	with default - "${NotExist|defValue}"
//	multi key 	 - "${GOPATH}/${APP_ENV | prod}/dir"
// Notice:
//  must add "?" - To ensure that there is no greedy match
//  var envRegex = regexp.MustCompile(`\${[\w-| ]+}`)
var envRegex = regexp.MustCompile(`\${.+?}`)

// VarParse alias of the ParseValue
func VarParse(str string) string {
	return ParseEnvValue(str)
}

// ParseEnvValue alias of the ParseValue
func ParseEnvValue(str string) string {
	return ParseValue(str)
}

// ParseValue parse ENV var value from input string, support default value.
// vars like ${var}, ${var| default}
//
// Usage:
// 	envutil.ParseValue()
func ParseValue(val string) (newVal string) {
	if strings.Index(val, "${") == -1 {
		return val
	}

	var name, def string
	return envRegex.ReplaceAllStringFunc(val, func(eVar string) string {
		// eVar like "${NotExist|defValue}", first remove "${" and "}", then split it
		ss := strings.SplitN(eVar[2:len(eVar)-1], "|", 2)

		// with default value. ${NotExist|defValue}
		if len(ss) == 2 {
			name, def = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
		} else {
			def = eVar // use raw value
			name = strings.TrimSpace(ss[0])
		}

		// get ENV value by name
		eVal := ValueGetter(name)
		if eVal == "" {
			eVal = def
		}
		return eVal
	})
}
