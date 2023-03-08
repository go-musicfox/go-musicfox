package goutil

import (
	"reflect"
	"strconv"

	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/strutil"
)

// Bool convert value to bool
func Bool(v any) bool {
	bl, _ := comfunc.ToBool(v)
	return bl
}

// ToBool try to convert type to bool
func ToBool(v any) (bool, error) {
	return comfunc.ToBool(v)
}

// String always convert value to string, will ignore error
func String(v any) string {
	s, _ := strutil.AnyToString(v, false)
	return s
}

// ToString convert value to string, will return error on fail.
func ToString(v any) (string, error) {
	return strutil.AnyToString(v, true)
}

// Int convert value to int
func Int(v any) int {
	iv, _ := mathutil.ToInt(v)
	return iv
}

// ToInt try to convert value to int
func ToInt(v any) (int, error) {
	return mathutil.ToInt(v)
}

// Int64 convert value to int64
func Int64(v any) int64 {
	iv, _ := mathutil.ToInt64(v)
	return iv
}

// ToInt64 try to convert value to int64
func ToInt64(v any) (int64, error) {
	return mathutil.ToInt64(v)
}

// Uint convert value to uint64
func Uint(v any) uint64 {
	iv, _ := mathutil.ToUint(v)
	return iv
}

// ToUint try to convert value to uint64
func ToUint(v any) (uint64, error) {
	return mathutil.ToUint(v)
}

// BaseTypeVal convert custom type or intX,uintX,floatX to generic base type.
//
//	intX/unitX 	=> int64
//	floatX      => float64
//	string 	    => string
//
// returns int64,string,float or error
func BaseTypeVal(val any) (value any, err error) {
	return reflects.BaseTypeVal(reflect.ValueOf(val))
}

// BoolString convert
func BoolString(bl bool) string {
	return strconv.FormatBool(bl)
}
