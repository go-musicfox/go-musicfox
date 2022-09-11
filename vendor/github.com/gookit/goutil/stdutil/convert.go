package stdutil

import (
	"errors"
	"reflect"

	"github.com/gookit/goutil/strutil"
)

var (
	ErrConvertFail = errors.New("convert value type is failure")
)

// ToString always convert value to string
func ToString(v interface{}) string {
	s, _ := strutil.AnyToString(v, false)
	return s
}

// MustString convert value(basic type) to string, will panic on convert a complex type.
func MustString(v interface{}) string {
	s, err := strutil.AnyToString(v, true)
	if err != nil {
		panic(err)
	}
	return s
}

// TryString try to convert a value to string
func TryString(v interface{}) (string, error) {
	return strutil.AnyToString(v, true)
}

// BaseTypeVal2 convert custom type or intX,uintX,floatX to generic base type.
//
// 	intX/unitX 	=> int64
// 	floatX      => float64
// 	string 	    => string
//
// returns int64,string,float or error
func BaseTypeVal2(v reflect.Value) (value interface{}, err error) {
	switch v.Kind() {
	case reflect.String:
		value = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = int64(v.Uint()) // always return int64
	case reflect.Float32, reflect.Float64:
		value = v.Float()
	default:
		err = ErrConvertFail
	}
	return
}

// BaseTypeVal convert custom type or intX,uintX,floatX to generic base type.
//
// 	intX/unitX 	=> int64
// 	floatX      => float64
// 	string 	    => string
//
// returns int64,string,float or error
func BaseTypeVal(val interface{}) (value interface{}, err error) {
	return BaseTypeVal2(reflect.ValueOf(val))
}
