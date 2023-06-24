package comfunc

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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

var (
	// TIP: extend unit d,w
	// time.ParseDuration() is not supported. eg: "1d", "2w"
	durStrReg = regexp.MustCompile(`^(-?\d+)(ns|us|µs|ms|s|m|h|d|w)$`)
	// match long duration string, such as "1hour", "2hours", "3minutes", "4mins", "5days", "1weeks"
	// time.ParseDuration() is not supported.
	durStrRegL = regexp.MustCompile(`^(-?\d+)([a-zA-Z]{3,})$`)
)

// IsDuration check the string is a duration string.
func IsDuration(s string) bool {
	if s == "0" || durStrReg.MatchString(s) {
		return true
	}
	return durStrRegL.MatchString(s)
}

// ToDuration parses a duration string. such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
//
// Diff of time.ParseDuration:
//   - support extend unit d, w at the end of string. such as "1d", "2w".
//   - support long string unit at end. such as "1hour", "2hours", "3minutes", "4mins", "5days", "1weeks".
//
// If the string is not a valid duration string, it will return an error.
func ToDuration(s string) (time.Duration, error) {
	ln := len(s)
	if ln == 0 {
		return 0, fmt.Errorf("empty duration string")
	}

	s = strings.ToLower(s)
	if s == "0" {
		return 0, nil
	}

	// extend unit d,w, time.ParseDuration() is not supported. eg: "1d", "2w"
	if lastUnit := s[ln-1]; lastUnit == 'd' {
		s = s + "ay"
	} else if lastUnit == 'w' {
		s = s + "eek"
	}

	// long unit, time.ParseDuration() is not supported. eg: "-3sec" => [3sec -3 sec]
	ss := durStrRegL.FindStringSubmatch(s)
	if len(ss) == 3 {
		num, unit := ss[1], ss[2]

		// convert to short unit
		switch unit {
		case "week", "weeks":
			// max unit is hour, so need convert by 24 * 7 * n
			n, _ := strconv.Atoi(num)
			s = strconv.Itoa(n*24*7) + "h"
		case "day", "days":
			// max unit is hour, so need convert by 24 * n
			n, _ := strconv.Atoi(num)
			s = strconv.Itoa(n*24) + "h"
		case "hour", "hours":
			s = num + "h"
		case "min", "mins", "minute", "minutes":
			s = num + "m"
		case "sec", "secs", "second", "seconds":
			s = num + "s"
		}
	}

	return time.ParseDuration(s)
}
