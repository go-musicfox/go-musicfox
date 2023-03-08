package reflects

import (
	"fmt"
	"reflect"
	"strconv"
)

// Elem returns the value that the interface v contains
// or that the pointer v points to.
func Elem(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		return v.Elem()
	}

	// otherwise, will return self
	return v
}

// Indirect like reflect.Indirect(), but can also indirect reflect.Interface
func Indirect(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		return v.Elem()
	}

	// otherwise, will return self
	return v
}

// Len get reflect value length
func Len(v reflect.Value) int {
	v = reflect.Indirect(v)

	// (u)int use width.
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String()))
	case reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return v.Len()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return len(strconv.FormatInt(int64(v.Uint()), 10))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return len(strconv.FormatInt(v.Int(), 10))
	case reflect.Float32, reflect.Float64:
		return len(fmt.Sprint(v.Interface()))
	}

	// cannot get length
	return -1
}

// SliceSubKind get sub-elem kind of the array, slice, variadic-var.
//
// Usage:
//
//	SliceSubKind(reflect.TypeOf([]string{"abc"})) // reflect.String
func SliceSubKind(typ reflect.Type) reflect.Kind {
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		return typ.Elem().Kind()
	}
	return reflect.Invalid
}

// SetValue to a reflect.Value
func SetValue(rv reflect.Value, val any) error {
	// get real type of the ptr value
	if rv.Kind() == reflect.Ptr {
		// init if is nil
		if rv.IsNil() {
			elemTyp := rv.Type().Elem()
			rv.Set(reflect.New(elemTyp))
		}

		// use elem for set value
		rv = reflect.Indirect(rv)
	}

	rv1, err := ValueByType(val, rv.Type())
	if err == nil {
		rv.Set(rv1)
	}
	return err
}

// FlatFunc custom collect handle func
type FlatFunc func(path string, val reflect.Value)

// FlatMap process tree map to flat key-value map.
//
// Examples:
//
//	{"top": {"sub": "value", "sub2": "value2"} }
//	->
//	{"top.sub": "value", "top.sub2": "value2" }
func FlatMap(rv reflect.Value, fn FlatFunc) {
	if fn == nil {
		return
	}

	if rv.Kind() != reflect.Map {
		panic("only allow flat map data")
	}
	flatMap(rv, fn, "")
}

func flatMap(rv reflect.Value, fn FlatFunc, parent string) {
	for _, key := range rv.MapKeys() {
		path := String(key)
		if parent != "" {
			path = parent + "." + path
		}

		fv := Indirect(rv.MapIndex(key))
		switch fv.Kind() {
		case reflect.Map:
			flatMap(fv, fn, path)
		case reflect.Array, reflect.Slice:
			flatSlice(fv, fn, path)
		default:
			fn(path, fv)
		}
	}
}

func flatSlice(rv reflect.Value, fn FlatFunc, parent string) {
	for i := 0; i < rv.Len(); i++ {
		path := parent + "[" + strconv.Itoa(i) + "]"
		fv := Indirect(rv.Index(i))

		switch fv.Kind() {
		case reflect.Map:
			flatMap(fv, fn, path)
		case reflect.Array, reflect.Slice:
			flatSlice(fv, fn, path)
		default:
			fn(path, fv)
		}
	}
}
