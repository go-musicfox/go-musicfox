package structs

import "reflect"

// Wrapper struct for read or set field value TODO
type Wrapper struct {
	// src any // source data struct
	rv reflect.Value

	// FieldTagName field name for read/write value. default tag: json
	FieldTagName string
}

// Wrap create a struct wrapper
func Wrap(src any) *Wrapper {
	return NewWrapper(src)
}

// NewWrapper create a struct wrapper
func NewWrapper(src any) *Wrapper {
	return WrapValue(reflect.ValueOf(src))
}

// WrapValue create a struct wrapper
func WrapValue(rv reflect.Value) *Wrapper {
	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		panic("must be provider an struct value")
	}

	return &Wrapper{rv: rv}
}

// Get field value by name
func (r *Wrapper) Get(name string) any {
	val, ok := r.Lookup(name)
	if !ok {
		return nil
	}
	return val
}

// Lookup field value by name
func (r *Wrapper) Lookup(name string) (val any, ok bool) {
	fv := r.rv.FieldByName(name)
	if !fv.IsValid() {
		return
	}

	return fv.Interface(), true
}
