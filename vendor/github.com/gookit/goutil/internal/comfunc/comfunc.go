package comfunc

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Environ like os.Environ, but will returns key-value map[string]string data.
func Environ() map[string]string {
	envList := os.Environ()
	envMap := make(map[string]string, len(envList))

	for _, str := range envList {
		nodes := strings.SplitN(str, "=", 2)

		if len(nodes) < 2 {
			envMap[nodes[0]] = ""
		} else {
			envMap[nodes[0]] = nodes[1]
		}
	}
	return envMap
}

// parse env value, allow:
//
//	only key 	 - "${SHELL}"
//	with default - "${NotExist | defValue}"
//	multi key 	 - "${GOPATH}/${APP_ENV | prod}/dir"
//
// Notice:
//
//	must add "?" - To ensure that there is no greedy match
//	var envRegex = regexp.MustCompile(`\${[\w-| ]+}`)
var envRegex = regexp.MustCompile(`\${.+?}`)

// ParseEnvVar parse ENV var value from input string, support default value.
//
// Format:
//
//	${var_name}            Only var name
//	${var_name | default}  With default value
//
// Usage:
//
//	comfunc.ParseEnvVar("${ APP_NAME }")
//	comfunc.ParseEnvVar("${ APP_ENV | dev }")
func ParseEnvVar(val string, getFn func(string) string) (newVal string) {
	if !strings.Contains(val, "${") {
		return val
	}

	// default use os.Getenv
	if getFn == nil {
		getFn = os.Getenv
	}

	var name, def string
	return envRegex.ReplaceAllStringFunc(val, func(eVar string) string {
		// eVar like "${NotExist|defValue}", first remove "${" and "}", then split it
		ss := strings.SplitN(eVar[2:len(eVar)-1], "|", 2)

		// with default value. ${NotExist|defValue}
		if len(ss) == 2 {
			name, def = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
		} else {
			name = strings.TrimSpace(ss[0])
		}

		// get ENV value by name
		eVal := getFn(name)
		if eVal == "" {
			eVal = def
		}
		return eVal
	})
}

// FormatTplAndArgs message
func FormatTplAndArgs(fmtAndArgs []any) string {
	if len(fmtAndArgs) == 0 || fmtAndArgs == nil {
		return ""
	}

	ln := len(fmtAndArgs)
	first := fmtAndArgs[0]

	if ln == 1 {
		if msgAsStr, ok := first.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", first)
	}

	// is template string.
	if tplStr, ok := first.(string); ok {
		return fmt.Sprintf(tplStr, fmtAndArgs[1:]...)
	}
	return fmt.Sprint(fmtAndArgs...)
}
