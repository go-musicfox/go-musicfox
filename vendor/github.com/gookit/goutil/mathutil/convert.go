package mathutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/comdef"
)

/*************************************************************
 * convert value to int
 *************************************************************/

// Int convert value to int
func Int(in any) (int, error) {
	return ToInt(in)
}

// QuietInt convert value to int, will ignore error
func QuietInt(in any) int {
	val, _ := ToInt(in)
	return val
}

// MustInt convert value to int, will panic on error
func MustInt(in any) int {
	val, _ := ToInt(in)
	return val
}

// IntOrPanic convert value to int, will panic on error
func IntOrPanic(in any) int {
	val, err := ToInt(in)
	if err != nil {
		panic(err)
	}
	return val
}

// IntOrErr convert value to int, return error on failed
func IntOrErr(in any) (iVal int, err error) {
	return ToInt(in)
}

// ToInt convert value to int, return error on failed
func ToInt(in any) (iVal int, err error) {
	switch tVal := in.(type) {
	case nil:
		iVal = 0
	case int:
		iVal = tVal
	case int8:
		iVal = int(tVal)
	case int16:
		iVal = int(tVal)
	case int32:
		iVal = int(tVal)
	case int64:
		iVal = int(tVal)
	case uint:
		iVal = int(tVal)
	case uint8:
		iVal = int(tVal)
	case uint16:
		iVal = int(tVal)
	case uint32:
		iVal = int(tVal)
	case uint64:
		iVal = int(tVal)
	case float32:
		iVal = int(tVal)
	case float64:
		iVal = int(tVal)
	case time.Duration:
		iVal = int(tVal)
	case string:
		iVal, err = strconv.Atoi(strings.TrimSpace(tVal))
	case json.Number:
		var i64 int64
		i64, err = tVal.Int64()
		iVal = int(i64)
	default:
		err = comdef.ErrConvType
	}
	return
}

// StrInt convert.
func StrInt(s string) int {
	iVal, _ := strconv.Atoi(strings.TrimSpace(s))
	return iVal
}

/*************************************************************
 * convert value to uint
 *************************************************************/

// Uint convert string to uint, return error on failed
func Uint(in any) (uint64, error) {
	return ToUint(in)
}

// QuietUint convert string to uint, will ignore error
func QuietUint(in any) uint64 {
	val, _ := ToUint(in)
	return val
}

// MustUint convert string to uint, will panic on error
func MustUint(in any) uint64 {
	val, _ := ToUint(in)
	return val
}

// UintOrErr convert value to uint, return error on failed
func UintOrErr(in any) (uint64, error) {
	return ToUint(in)
}

// ToUint convert value to uint, return error on failed
func ToUint(in any) (u64 uint64, err error) {
	switch tVal := in.(type) {
	case nil:
		u64 = 0
	case int:
		u64 = uint64(tVal)
	case int8:
		u64 = uint64(tVal)
	case int16:
		u64 = uint64(tVal)
	case int32:
		u64 = uint64(tVal)
	case int64:
		u64 = uint64(tVal)
	case uint:
		u64 = uint64(tVal)
	case uint8:
		u64 = uint64(tVal)
	case uint16:
		u64 = uint64(tVal)
	case uint32:
		u64 = uint64(tVal)
	case uint64:
		u64 = tVal
	case float32:
		u64 = uint64(tVal)
	case float64:
		u64 = uint64(tVal)
	case time.Duration:
		u64 = uint64(tVal)
	case json.Number:
		var i64 int64
		i64, err = tVal.Int64()
		u64 = uint64(i64)
	case string:
		u64, err = strconv.ParseUint(strings.TrimSpace(tVal), 10, 0)
	default:
		err = comdef.ErrConvType
	}
	return
}

/*************************************************************
 * convert value to int64
 *************************************************************/

// Int64 convert string to int64, return error on failed
func Int64(in any) (int64, error) {
	return ToInt64(in)
}

// SafeInt64 convert value to int64, will ignore error
func SafeInt64(in any) int64 {
	i64, _ := ToInt64(in)
	return i64
}

// QuietInt64 convert value to int64, will ignore error
func QuietInt64(in any) int64 {
	i64, _ := ToInt64(in)
	return i64
}

// MustInt64 convert value to int64, will panic on error
func MustInt64(in any) int64 {
	i64, _ := ToInt64(in)
	return i64
}

// TODO StrictInt64,AsInt64 strict convert to int64

// Int64OrErr convert string to int64, return error on failed
func Int64OrErr(in any) (int64, error) {
	return ToInt64(in)
}

// ToInt64 convert string to int64, return error on failed
func ToInt64(in any) (i64 int64, err error) {
	switch tVal := in.(type) {
	case nil:
		i64 = 0
	case string:
		i64, err = strconv.ParseInt(strings.TrimSpace(tVal), 10, 0)
	case int:
		i64 = int64(tVal)
	case int8:
		i64 = int64(tVal)
	case int16:
		i64 = int64(tVal)
	case int32:
		i64 = int64(tVal)
	case int64:
		i64 = tVal
	case uint:
		i64 = int64(tVal)
	case uint8:
		i64 = int64(tVal)
	case uint16:
		i64 = int64(tVal)
	case uint32:
		i64 = int64(tVal)
	case uint64:
		i64 = int64(tVal)
	case float32:
		i64 = int64(tVal)
	case float64:
		i64 = int64(tVal)
	case time.Duration:
		i64 = int64(tVal)
	case json.Number:
		i64, err = tVal.Int64()
	default:
		err = comdef.ErrConvType
	}
	return
}

/*************************************************************
 * convert value to float
 *************************************************************/

// QuietFloat convert value to float64, will ignore error
func QuietFloat(in any) float64 {
	val, _ := ToFloat(in)
	return val
}

// FloatOrPanic convert value to float64, will panic on error
func FloatOrPanic(in any) float64 {
	val, err := ToFloat(in)
	if err != nil {
		panic(err)
	}
	return val
}

// MustFloat convert value to float64 TODO will panic on error
func MustFloat(in any) float64 {
	val, _ := ToFloat(in)
	return val
}

// Float convert value to float64, return error on failed
func Float(in any) (float64, error) {
	return ToFloat(in)
}

// FloatOrErr convert value to float64, return error on failed
func FloatOrErr(in any) (float64, error) {
	return ToFloat(in)
}

// ToFloat convert value to float64, return error on failed
func ToFloat(in any) (f64 float64, err error) {
	switch tVal := in.(type) {
	case nil:
		f64 = 0
	case string:
		f64, err = strconv.ParseFloat(strings.TrimSpace(tVal), 64)
	case int:
		f64 = float64(tVal)
	case int8:
		f64 = float64(tVal)
	case int16:
		f64 = float64(tVal)
	case int32:
		f64 = float64(tVal)
	case int64:
		f64 = float64(tVal)
	case uint:
		f64 = float64(tVal)
	case uint8:
		f64 = float64(tVal)
	case uint16:
		f64 = float64(tVal)
	case uint32:
		f64 = float64(tVal)
	case uint64:
		f64 = float64(tVal)
	case float32:
		f64 = float64(tVal)
	case float64:
		f64 = tVal
	case time.Duration:
		f64 = float64(tVal)
	case json.Number:
		f64, err = tVal.Float64()
	default:
		err = comdef.ErrConvType
	}
	return
}

/*************************************************************
 * convert intX/floatX to string
 *************************************************************/

// StringOrPanic convert intX/floatX value to string, will panic on error
func StringOrPanic(val any) string {
	str, err := TryToString(val, true)
	if err != nil {
		panic(err)
	}
	return str
}

// MustString convert intX/floatX value to string, will panic on error
func MustString(val any) string {
	return StringOrPanic(val)
}

// ToString convert intX/floatX value to string, return error on failed
func ToString(val any) (string, error) {
	return TryToString(val, true)
}

// StringOrErr convert intX/floatX value to string, return error on failed
func StringOrErr(val any) (string, error) {
	return TryToString(val, true)
}

// QuietString convert intX/floatX value to string, other type convert by fmt.Sprint
func QuietString(val any) string {
	str, _ := TryToString(val, false)
	return str
}

// String convert intX/floatX value to string, other type convert by fmt.Sprint
func String(val any) string {
	str, _ := TryToString(val, false)
	return str
}

// TryToString try convert intX/floatX value to string
//
// if defaultAsErr is False, will use fmt.Sprint convert other type
func TryToString(val any, defaultAsErr bool) (str string, err error) {
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
