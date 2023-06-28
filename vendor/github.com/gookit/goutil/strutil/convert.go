package strutil

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/mathutil"
)

var (
	// ErrDateLayout error
	ErrDateLayout = errors.New("invalid date layout string")
	// ErrInvalidParam error
	ErrInvalidParam = errors.New("invalid input for parse time")

	// some regex for convert string.
	toSnakeReg  = regexp.MustCompile("[A-Z][a-z]")
	toCamelRegs = map[string]*regexp.Regexp{
		" ": regexp.MustCompile(" +[a-zA-Z]"),
		"-": regexp.MustCompile("-+[a-zA-Z]"),
		"_": regexp.MustCompile("_+[a-zA-Z]"),
	}
)

// Internal func refers:
// strconv.QuoteRune()
// strconv.QuoteToASCII()
// strconv.AppendQuote()
// strconv.AppendQuoteRune()

// Quote alias of strings.Quote
func Quote(s string) string { return strconv.Quote(s) }

// Unquote remove start and end quotes by single-quote or double-quote
//
// tip: strconv.Unquote cannot unquote single-quote
func Unquote(s string) string {
	ln := len(s)
	if ln < 2 {
		return s
	}

	qs, qe := s[0], s[ln-1]

	var valid bool
	if qs == '"' && qe == '"' {
		valid = true
	} else if qs == '\'' && qe == '\'' {
		valid = true
	}

	if valid {
		s = s[1 : ln-1] // exclude quotes
	}
	// strconv.Unquote cannot unquote single-quote
	// if ns, err := strconv.Unquote(s); err == nil {
	// 	return ns
	// }
	return s
}

// Join alias of strings.Join
func Join(sep string, ss ...string) string { return strings.Join(ss, sep) }

// JoinList alias of strings.Join
func JoinList(sep string, ss []string) string { return strings.Join(ss, sep) }

// JoinAny type to string
func JoinAny(sep string, parts ...any) string {
	ss := make([]string, 0, len(parts))
	for _, part := range parts {
		ss = append(ss, QuietString(part))
	}

	return strings.Join(ss, sep)
}

// Implode alias of strings.Join
func Implode(sep string, ss ...string) string { return strings.Join(ss, sep) }

/*************************************************************
 * convert value to string
 *************************************************************/

// String convert value to string, return error on failed
func String(val any) (string, error) {
	return AnyToString(val, true)
}

// ToString convert value to string, return error on failed
func ToString(val any) (string, error) {
	return AnyToString(val, true)
}

// QuietString convert value to string, will ignore error
func QuietString(in any) string {
	val, _ := AnyToString(in, false)
	return val
}

// SafeString convert value to string, will ignore error
func SafeString(in any) string {
	val, _ := AnyToString(in, false)
	return val
}

// MustString convert value to string, will panic on error
func MustString(in any) string {
	val, err := AnyToString(in, false)
	if err != nil {
		panic(err)
	}
	return val
}

// StringOrErr convert value to string, return error on failed
func StringOrErr(val any) (string, error) {
	return AnyToString(val, true)
}

// AnyToString convert value to string.
//
// For defaultAsErr:
//
//   - False  will use fmt.Sprint convert complex type
//   - True   will return error on fail.
func AnyToString(val any, defaultAsErr bool) (str string, err error) {
	if val == nil {
		return
	}

	switch value := val.(type) {
	case int:
		str = strconv.Itoa(value)
	case int8:
		str = strconv.Itoa(int(value))
	case int16:
		str = strconv.Itoa(int(value))
	case int32: // same as `rune`
		str = strconv.Itoa(int(value))
	case int64:
		str = strconv.FormatInt(value, 10)
	case uint:
		str = strconv.FormatUint(uint64(value), 10)
	case uint8:
		str = strconv.FormatUint(uint64(value), 10)
	case uint16:
		str = strconv.FormatUint(uint64(value), 10)
	case uint32:
		str = strconv.FormatUint(uint64(value), 10)
	case uint64:
		str = strconv.FormatUint(value, 10)
	case float32:
		str = strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		str = strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(value)
	case string:
		str = value
	case []byte:
		str = string(value)
	case time.Duration:
		str = strconv.FormatInt(int64(value), 10)
	case fmt.Stringer:
		str = value.String()
	default:
		if defaultAsErr {
			err = comdef.ErrConvType
		} else {
			str = fmt.Sprint(value)
		}
	}
	return
}

/*************************************************************
 * convert string value to byte
 * refer from https://github.com/valyala/fastjson/blob/master/util.go
 *************************************************************/

// Byte2str convert bytes to string
func Byte2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Byte2string convert bytes to string
func Byte2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToBytes convert string to bytes
func ToBytes(s string) (b []byte) {
	strh := (*reflect.StringHeader)(unsafe.Pointer(&s))

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = strh.Data
	sh.Len = strh.Len
	sh.Cap = strh.Len
	return b
}

/*************************************************************
 * convert string value to bool
 *************************************************************/

// ToBool convert string to bool
func ToBool(s string) (bool, error) {
	return comfunc.StrToBool(s)
}

// QuietBool convert to bool, will ignore error
func QuietBool(s string) bool {
	val, _ := comfunc.StrToBool(strings.TrimSpace(s))
	return val
}

// MustBool convert, will panic on error
func MustBool(s string) bool {
	val, err := comfunc.StrToBool(strings.TrimSpace(s))
	if err != nil {
		panic(err)
	}
	return val
}

// Bool parse string to bool. like strconv.ParseBool()
func Bool(s string) (bool, error) {
	return comfunc.StrToBool(s)
}

/*************************************************************
 * convert string value to int, float
 *************************************************************/

// Int convert string to int, alias of ToInt()
func Int(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// ToInt convert string to int, return error on fail
func ToInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// Int2 convert string to int, will ignore error
func Int2(s string) int {
	val, _ := ToInt(s)
	return val
}

// QuietInt convert string to int, will ignore error
func QuietInt(s string) int {
	val, _ := ToInt(s)
	return val
}

// MustInt convert string to int, will panic on error
func MustInt(s string) int {
	return IntOrPanic(s)
}

// IntOrPanic convert value to int, will panic on error
func IntOrPanic(s string) int {
	val, err := ToInt(s)
	if err != nil {
		panic(err)
	}
	return val
}

// Int64 convert string to int, will ignore error
func Int64(s string) int64 {
	val, _ := Int64OrErr(s)
	return val
}

// QuietInt64 convert string to int, will ignore error
func QuietInt64(s string) int64 {
	val, _ := Int64OrErr(s)
	return val
}

// ToInt64 convert string to int, return error on fail
func ToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 0)
}

// Int64OrErr convert string to int, return error on fail
func Int64OrErr(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 0)
}

// MustInt64 convert value to int, will panic on error
func MustInt64(s string) int64 {
	return Int64OrPanic(s)
}

// Int64OrPanic convert value to int, will panic on error
func Int64OrPanic(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		panic(err)
	}
	return val
}

/*************************************************************
 * convert string value to int/string slice, time.Time
 *************************************************************/

// Ints alias of the ToIntSlice(). default sep is comma(,)
func Ints(s string, sep ...string) []int {
	ints, _ := ToIntSlice(s, sep...)
	return ints
}

// ToInts alias of the ToIntSlice(). default sep is comma(,)
func ToInts(s string, sep ...string) ([]int, error) { return ToIntSlice(s, sep...) }

// ToIntSlice split string to slice and convert item to int.
//
// Default sep is comma(,)
func ToIntSlice(s string, sep ...string) (ints []int, err error) {
	ss := ToSlice(s, sep...)
	for _, item := range ss {
		iVal, err := mathutil.ToInt(item)
		if err != nil {
			return []int{}, err
		}

		ints = append(ints, iVal)
	}
	return
}

// ToArray alias of the ToSlice()
func ToArray(s string, sep ...string) []string { return ToSlice(s, sep...) }

// Strings alias of the ToSlice()
func Strings(s string, sep ...string) []string { return ToSlice(s, sep...) }

// ToStrings alias of the ToSlice()
func ToStrings(s string, sep ...string) []string { return ToSlice(s, sep...) }

// ToSlice split string to array.
func ToSlice(s string, sep ...string) []string {
	if len(sep) > 0 {
		return Split(s, sep[0])
	}
	return Split(s, ",")
}

// ToOSArgs split string to string[](such as os.Args)
// func ToOSArgs(s string) []string {
// 	return cliutil.StringToOSArgs(s) // error: import cycle not allowed
// }

// ToDuration parses a duration string. such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func ToDuration(s string) (time.Duration, error) {
	return comfunc.ToDuration(s)
}
