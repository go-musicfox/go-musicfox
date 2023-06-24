package reflects

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
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

// SliceSubKind get sub-elem kind of the array, slice, variadic-var. alias SliceElemKind()
func SliceSubKind(typ reflect.Type) reflect.Kind {
	return SliceElemKind(typ)
}

// SliceElemKind get sub-elem kind of the array, slice, variadic-var.
//
// Usage:
//
//	SliceElemKind(reflect.TypeOf([]string{"abc"})) // reflect.String
func SliceElemKind(typ reflect.Type) reflect.Kind {
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		return typ.Elem().Kind()
	}
	return reflect.Invalid
}

// UnexportedValue quickly get unexported value by reflect.Value
//
// NOTE: this method is unsafe, use it carefully.
// should ensure rv is addressable by field.CanAddr()
//
// refer: https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields
func UnexportedValue(rv reflect.Value) any {
	if rv.CanAddr() {
		// create new value from addr, now can be read and set.
		return reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Interface()
	}

	// If the rv is not addressable this trick won't work, but you can create an addressable copy like this
	rs2 := reflect.New(rv.Type()).Elem()
	rs2.Set(rv)
	rv = rs2.Field(0)
	rv = reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem()
	// Now rv can be read. TIP: Setting will succeed but only affects the temporary copy.
	return rv.Interface()
}

// SetUnexportedValue quickly set unexported field value by reflect
//
// NOTE: this method is unsafe, use it carefully.
// should ensure rv is addressable by field.CanAddr()
func SetUnexportedValue(rv reflect.Value, value any) {
	reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

// SetValue to a `reflect.Value`. will auto convert type if needed.
func SetValue(rv reflect.Value, val any) error {
	// get real type of the ptr value
	if rv.Kind() == reflect.Ptr {
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

// SetRValue to a `reflect.Value`. will direct set value without convert type.
func SetRValue(rv, val reflect.Value) {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			elemTyp := rv.Type().Elem()
			rv.Set(reflect.New(elemTyp))
		}
		rv = reflect.Indirect(rv)
	}

	rv.Set(val)
}

// EachMap process any map data
func EachMap(mp reflect.Value, fn func(key, val reflect.Value)) {
	if fn == nil {
		return
	}
	if mp.Kind() != reflect.Map {
		panic("only allow map value data")
	}

	for _, key := range mp.MapKeys() {
		fn(key, mp.MapIndex(key))
	}
}

// EachStrAnyMap process any map data as string key and any value
func EachStrAnyMap(mp reflect.Value, fn func(key string, val any)) {
	EachMap(mp, func(key, val reflect.Value) {
		fn(String(key), val.Interface())
	})
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
