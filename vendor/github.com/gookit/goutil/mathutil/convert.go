package mathutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrConvertFail convert error
	ErrConvertFail = errors.New("convert data type is failure")
	// ErrConvertFail = errors.New("convert data type is failure")
)

/*************************************************************
 * convert value to int
 *************************************************************/

// Int convert value to int
func Int(in interface{}) (int, error) {
	return ToInt(in)
}

// MustInt convert value to int
func MustInt(in interface{}) int {
	val, _ := ToInt(in)
	return val
}

// IntOrPanic convert value to int, will panic on error
func IntOrPanic(in interface{}) int {
	val, err := ToInt(in)
	if err != nil {
		panic(err)
	}
	return val
}

// ToInt convert string to int
func ToInt(in interface{}) (iVal int, err error) {
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
		err = ErrConvertFail
	}
	return
}

/*************************************************************
 * convert value to uint
 *************************************************************/

// Uint convert string to uint
func Uint(in interface{}) (uint64, error) {
	return ToUint(in)
}

// MustUint convert string to uint
func MustUint(in interface{}) uint64 {
	val, _ := ToUint(in)
	return val
}

// ToUint convert string to uint
func ToUint(in interface{}) (u64 uint64, err error) {
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
		err = ErrConvertFail
	}
	return
}

/*************************************************************
 * convert value to int64
 *************************************************************/

// Int64 convert string to int64
func Int64(in interface{}) (int64, error) {
	return ToInt64(in)
}

// MustInt64 convert
func MustInt64(in interface{}) int64 {
	i64, _ := ToInt64(in)
	return i64
}

// ToInt64 convert string to int64
func ToInt64(in interface{}) (i64 int64, err error) {
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
		err = ErrConvertFail
	}
	return
}

/*************************************************************
 * convert value to float
 *************************************************************/

// Float convert value to float64
func Float(in interface{}) (float64, error) {
	return ToFloat(in)
}

// ToFloat convert value to float64
func ToFloat(in interface{}) (f64 float64, err error) {
	switch tVal := in.(type) {
	case nil:
		f64 = 0
	case string:
		f64, err = strconv.ParseFloat(strings.TrimSpace(tVal), 0)
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
		err = ErrConvertFail
	}
	return
}

// FloatOrPanic convert value to float64, will panic on error
func FloatOrPanic(in interface{}) float64 {
	val, err := ToFloat(in)
	if err != nil {
		panic(err)
	}
	return val
}

// MustFloat convert value to float64 TODO will panic on error
func MustFloat(in interface{}) float64 {
	val, _ := ToFloat(in)
	return val
}

/*************************************************************
 * convert intX/floatX to string
 *************************************************************/

// TryToString try convert intX/floatX value to string
//
// if defaultAsErr is False, will use fmt.Sprint convert other type
func TryToString(val interface{}, defaultAsErr bool) (str string, err error) {
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
	case time.Duration:
		str = strconv.FormatUint(uint64(value.Nanoseconds()), 10)
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

// StringOrPanic convert intX/floatX value to string, will panic on error
func StringOrPanic(val interface{}) string {
	str, err := TryToString(val, true)
	if err != nil {
		panic(err)
	}
	return str
}

// MustString convert intX/floatX value to string, will panic on error
func MustString(val interface{}) string {
	return StringOrPanic(val)
}

// ToString convert intX/floatX value to string
func ToString(val interface{}) (string, error) {
	return TryToString(val, true)
}

// String convert intX/floatX value to string, other type convert by fmt.Sprint
func String(val interface{}) string {
	str, _ := TryToString(val, false)
	return str
}
