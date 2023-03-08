package stdutil

import (
	"reflect"

	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/strutil"
)

// ToString always convert value to string, will ignore error
func ToString(v any) string {
	s, _ := strutil.AnyToString(v, false)
	return s
}

// MustString convert value(basic type) to string, will panic on convert a complex type.
func MustString(v any) string {
	s, err := strutil.AnyToString(v, true)
	if err != nil {
		panic(err)
	}
	return s
}

// TryString try to convert a value to string
func TryString(v any) (string, error) {
	return strutil.AnyToString(v, true)
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

// BaseTypeVal2 convert custom type or intX,uintX,floatX to generic base type.
//
//	intX/unitX 	=> int64
//	floatX      => float64
//	string 	    => string
//
// returns int64,string,float or error
func BaseTypeVal2(v reflect.Value) (value any, err error) {
	return reflects.BaseTypeVal(v)
}
