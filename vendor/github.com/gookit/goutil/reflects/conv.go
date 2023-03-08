package reflects

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// BaseTypeVal convert custom type or intX,uintX,floatX to generic base type.
//
//	intX/unitX 	=> int64
//	floatX      => float64
//	string 	    => string
//
// returns int64,string,float or error
func BaseTypeVal(v reflect.Value) (value any, err error) {
	v = reflect.Indirect(v)

	switch v.Kind() {
	case reflect.String:
		value = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value = int64(v.Uint()) // always return int64
	case reflect.Float32, reflect.Float64:
		value = v.Float()
	default:
		err = comdef.ErrConvType
	}
	return
}

// ValueByType create reflect.Value by give reflect.Type
func ValueByType(val any, typ reflect.Type) (rv reflect.Value, err error) {
	// handle kind: string, bool, intX, uintX, floatX
	if typ.Kind() == reflect.String || typ.Kind() <= reflect.Float64 {
		return ValueByKind(val, typ.Kind())
	}

	newRv := reflect.ValueOf(val)

	// try auto convert slice type
	if IsArrayOrSlice(newRv.Kind()) && IsArrayOrSlice(typ.Kind()) {
		return ConvSlice(newRv, typ.Elem())
	}

	// check type. like map
	if newRv.Type() == typ {
		return newRv, nil
	}

	err = comdef.ErrConvType
	return
}

// ValueByKind create reflect.Value by give reflect.Kind
//
// TIPs:
//
//	Only support kind: string, bool, intX, uintX, floatX
func ValueByKind(val any, kind reflect.Kind) (rv reflect.Value, err error) {
	switch kind {
	case reflect.Int:
		if dstV, err1 := mathutil.ToInt(val); err1 == nil {
			rv = reflect.ValueOf(dstV)
		}
	case reflect.Int8:
		if dstV, err1 := mathutil.ToInt(val); err1 == nil {
			rv = reflect.ValueOf(int8(dstV))
		}
	case reflect.Int16:
		if dstV, err1 := mathutil.ToInt(val); err1 == nil {
			rv = reflect.ValueOf(int16(dstV))
		}
	case reflect.Int32:
		if dstV, err1 := mathutil.ToInt(val); err1 == nil {
			rv = reflect.ValueOf(int32(dstV))
		}
	case reflect.Int64:
		if dstV, err1 := mathutil.ToInt64(val); err1 == nil {
			rv = reflect.ValueOf(dstV)
		}
	case reflect.Uint:
		if dstV, err1 := mathutil.ToUint(val); err1 == nil {
			rv = reflect.ValueOf(uint(dstV))
		}
	case reflect.Uint8:
		if dstV, err1 := mathutil.ToUint(val); err1 == nil {
			rv = reflect.ValueOf(uint8(dstV))
		}
	case reflect.Uint16:
		if dstV, err1 := mathutil.ToUint(val); err1 == nil {
			rv = reflect.ValueOf(uint16(dstV))
		}
	case reflect.Uint32:
		if dstV, err1 := mathutil.ToUint(val); err1 == nil {
			rv = reflect.ValueOf(uint32(dstV))
		}
	case reflect.Uint64:
		if dstV, err1 := mathutil.ToUint(val); err1 == nil {
			rv = reflect.ValueOf(dstV)
		}
	case reflect.Float32:
		if dstV, err1 := mathutil.ToFloat(val); err1 == nil {
			rv = reflect.ValueOf(float32(dstV))
		}
	case reflect.Float64:
		if dstV, err1 := mathutil.ToFloat(val); err1 == nil {
			rv = reflect.ValueOf(dstV)
		}
	case reflect.String:
		if dstV, err1 := strutil.ToString(val); err1 == nil {
			rv = reflect.ValueOf(dstV)
		}
	case reflect.Bool:
		if bl, err := comfunc.ToBool(val); err == nil {
			rv = reflect.ValueOf(bl)
		}
	}

	if !rv.IsValid() {
		err = comdef.ErrConvType
	}
	return
}

// ConvSlice make new type slice from old slice
func ConvSlice(oldSlRv reflect.Value, newElemTyp reflect.Type) (rv reflect.Value, err error) {
	if !IsArrayOrSlice(oldSlRv.Kind()) {
		panic("only allow array or slice type value")
	}

	// do not need convert type
	if oldSlRv.Type().Elem() == newElemTyp {
		return oldSlRv, nil
	}

	newSlTyp := reflect.SliceOf(newElemTyp)
	newSlRv := reflect.MakeSlice(newSlTyp, 0, 0)
	for i := 0; i < oldSlRv.Len(); i++ {
		newElemV, err := ValueByKind(oldSlRv.Index(i).Interface(), newElemTyp.Kind())
		if err != nil {
			return reflect.Value{}, err
		}

		newSlRv = reflect.Append(newSlRv, newElemV)
	}
	return newSlRv, nil
}

// String convert
func String(rv reflect.Value) string {
	s, _ := ValToString(rv, false)
	return s
}

// ToString convert
func ToString(rv reflect.Value) (str string, err error) {
	return ValToString(rv, true)
}

// ValToString convert handle
func ValToString(rv reflect.Value, defaultAsErr bool) (str string, err error) {
	rv = Indirect(rv)
	switch rv.Kind() {
	case reflect.Invalid:
		str = ""
	case reflect.Bool:
		str = strconv.FormatBool(rv.Bool())
	case reflect.String:
		str = rv.String()
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		str = strconv.FormatUint(rv.Uint(), 10)
	default:
		if defaultAsErr {
			err = comdef.ErrConvType
		} else {
			str = fmt.Sprint(rv.Interface())
		}
	}
	return
}
