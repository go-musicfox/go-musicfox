package strutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/gookit/goutil/mathutil"
)

var (
	ErrConvertFail  = errors.New("convert data type is failure")
	ErrInvalidParam = errors.New("invalid input parameter")

	// some regex for convert string.
	toSnakeReg  = regexp.MustCompile("[A-Z][a-z]")
	toCamelRegs = map[string]*regexp.Regexp{
		" ": regexp.MustCompile(" +[a-zA-Z]"),
		"-": regexp.MustCompile("-+[a-zA-Z]"),
		"_": regexp.MustCompile("_+[a-zA-Z]"),
	}
)

// Internal func refers:
// strconv.Quote()
// strconv.QuoteRune()
// strconv.QuoteToASCII()
// strconv.AppendQuote()
// strconv.AppendQuoteRune()

// Join alias of strings.Join
func Join(sep string, ss ...string) string {
	return strings.Join(ss, sep)
}

// Implode alias of strings.Join
func Implode(sep string, ss ...string) string {
	return strings.Join(ss, sep)
}

/*************************************************************
 * convert value to string
 *************************************************************/

// String convert val to string
func String(val interface{}) (string, error) {
	return AnyToString(val, true)
}

// MustString convert value to string
func MustString(in interface{}) string {
	val, _ := AnyToString(in, false)
	return val
}

// ToString convert value to string
func ToString(val interface{}) (string, error) {
	return AnyToString(val, true)
}

// AnyToString convert value to string.
//
// if defaultAsErr is False, will use fmt.Sprint convert complex type
func AnyToString(val interface{}, defaultAsErr bool) (str string, err error) {
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
		str = strconv.Itoa(int(value))
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
		str = value.String()
	case json.Number:
		str = value.String()
	default:
		if defaultAsErr {
			err = ErrConvertFail
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
	return Bool(s)
}

// MustBool convert.
func MustBool(s string) bool {
	val, _ := Bool(strings.TrimSpace(s))
	return val
}

// Bool parse string to bool
func Bool(s string) (bool, error) {
	// return strconv.ParseBool(Trim(s))
	lower := strings.ToLower(s)
	switch lower {
	case "1", "on", "yes", "true":
		return true, nil
	case "0", "off", "no", "false":
		return false, nil
	}

	return false, fmt.Errorf("'%s' cannot convert to bool", s)
}

/*************************************************************
 * convert string value to int, float
 *************************************************************/

// Int convert string to int, alias of ToInt()
func Int(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// ToInt convert string to int
func ToInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// MustInt convert string to int
func MustInt(s string) int {
	val, _ := ToInt(s)
	return val
}

// IntOrPanic convert value to int, will panic on error
func IntOrPanic(s string) int {
	val, err := ToInt(s)
	if err != nil {
		panic(err)
	}
	return val
}

/*************************************************************
 * convert string value to int/string slice, time.Time
 *************************************************************/

// Ints alias of the ToIntSlice()
func Ints(s string, sep ...string) []int {
	ints, _ := ToIntSlice(s, sep...)
	return ints
}

// ToInts alias of the ToIntSlice()
func ToInts(s string, sep ...string) ([]int, error) {
	return ToIntSlice(s, sep...)
}

// ToIntSlice split string to slice and convert item to int.
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
func ToArray(s string, sep ...string) []string {
	return ToSlice(s, sep...)
}

// Strings alias of the ToSlice()
func Strings(s string, sep ...string) []string {
	return ToSlice(s, sep...)
}

// ToStrings alias of the ToSlice()
func ToStrings(s string, sep ...string) []string {
	return ToSlice(s, sep...)
}

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

// MustToTime convert date string to time.Time
func MustToTime(s string, layouts ...string) time.Time {
	t, err := ToTime(s, layouts...)
	if err != nil {
		panic(err)
	}
	return t
}

// ToTime convert date string to time.Time
func ToTime(s string, layouts ...string) (t time.Time, err error) {
	var layout string
	if len(layouts) > 0 { // custom layout
		layout = layouts[0]
		if layout == "" {
			err = ErrInvalidParam
			return
		}
	} else { // auto match layout.
		switch len(s) {
		case 8:
			layout = "20060102"
		case 10:
			layout = "2006-01-02"
		case 13:
			layout = "2006-01-02 15"
		case 16:
			layout = "2006-01-02 15:04"
		case 19:
			layout = "2006-01-02 15:04:05"
		case 20: // time.RFC3339
			layout = "2006-01-02T15:04:05Z07:00"
		default:
			err = ErrInvalidParam
			return
		}

		// has 'T' eg: "2006-01-02T15:04:05"
		if strings.ContainsRune(s, 'T') {
			layout = strings.Replace(layout, " ", "T", -1)
		}

		// eg: "2006/01/02 15:04:05"
		if strings.ContainsRune(s, '/') {
			layout = strings.Replace(layout, "-", "/", -1)
		}
	}

	t, err = time.Parse(layout, s)
	// t, err = time.ParseInLocation(layout, s, time.Local)
	return
}
