package strutil

import (
	"regexp"
	"strings"
	"unicode"
)

// Title Some alias methods.
var (
	Title = strings.ToTitle
)

/*************************************************************
 * change string case
 *************************************************************/

// Lower alias of the strings.ToLower()
func Lower(s string) string { return strings.ToLower(s) }

// Lowercase alias of the strings.ToLower()
func Lowercase(s string) string { return strings.ToLower(s) }

// Upper alias of the strings.ToUpper()
func Upper(s string) string { return strings.ToUpper(s) }

// Uppercase alias of the strings.ToUpper()
func Uppercase(s string) string { return strings.ToUpper(s) }

// UpperWord Change the first character of each word to uppercase
func UpperWord(s string) string {
	if len(s) == 0 {
		return s
	}

	if len(s) == 1 {
		return strings.ToUpper(s)
	}

	inWord := true
	buf := make([]byte, 0, len(s))

	i := 0
	rs := []rune(s)
	if runeIsLowerChar(rs[i]) {
		buf = append(buf, []byte(string(unicode.ToUpper(rs[i])))...)
	} else {
		buf = append(buf, []byte(string(rs[i]))...)
	}

	for j := i + 1; j < len(rs); j++ {
		if !runeIsWord(rs[i]) && runeIsWord(rs[j]) {
			inWord = false
		}

		if runeIsLowerChar(rs[j]) && !inWord {
			buf = append(buf, []byte(string(unicode.ToUpper(rs[j])))...)
			inWord = true
		} else {
			buf = append(buf, []byte(string(rs[j]))...)
		}

		if runeIsWord(rs[j]) {
			inWord = true
		}

		i++
	}

	return string(buf)
}

// LowerFirst lower first char
func LowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	rs := []rune(s)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		return string(unicode.ToLower(f)) + string(rs[1:])
	}

	return s
}

// UpperFirst upper first char
func UpperFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	rs := []rune(s)
	f := rs[0]
	if 'a' <= f && f <= 'z' {
		return string(unicode.ToUpper(f)) + string(rs[1:])
	}

	return s
}

// Snake alias of the SnakeCase
func Snake(s string, sep ...string) string {
	return SnakeCase(s, sep...)
}

// SnakeCase convert. eg "RangePrice" -> "range_price"
func SnakeCase(s string, sep ...string) string {
	sepChar := "_"
	if len(sep) > 0 {
		sepChar = sep[0]
	}

	newStr := toSnakeReg.ReplaceAllStringFunc(s, func(s string) string {
		return sepChar + LowerFirst(s)
	})

	return strings.TrimLeft(newStr, sepChar)
}

// Camel alias of the CamelCase
func Camel(s string, sep ...string) string {
	return CamelCase(s, sep...)
}

// CamelCase convert string to camel case.
// Support:
// 	"range_price" -> "rangePrice"
// 	"range price" -> "rangePrice"
// 	"range-price" -> "rangePrice"
func CamelCase(s string, sep ...string) string {
	sepChar := "_"
	if len(sep) > 0 {
		sepChar = sep[0]
	}

	// Not contains sep char
	if !strings.Contains(s, sepChar) {
		return s
	}

	// Get regexp instance
	rgx, ok := toCamelRegs[sepChar]
	if !ok {
		rgx = regexp.MustCompile(regexp.QuoteMeta(sepChar) + "+[a-zA-Z]")
	}

	return rgx.ReplaceAllStringFunc(s, func(s string) string {
		s = strings.TrimLeft(s, sepChar)
		return UpperFirst(s)
	})
}

func runeIsWord(c rune) bool {
	return runeIsLowerChar(c) || runeIsUpperChar(c)
}

func runeIsLowerChar(c rune) bool {
	return 'a' <= c && c <= 'z'
}

func runeIsUpperChar(c rune) bool {
	return 'A' <= c && c <= 'Z'
}
