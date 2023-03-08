package reflects

import "reflect"

// BKind base data kind type
type BKind uint

// base kinds
const (
	// Int for all intX types
	Int = BKind(reflect.Int)
	// Uint for all uintX types
	Uint = BKind(reflect.Uint)
	// Float for all floatX types
	Float = BKind(reflect.Float32)
	// Array for array,slice types
	Array = BKind(reflect.Array)
	// Complex for all complexX types
	Complex = BKind(reflect.Complex64)
)

// ToBaseKind convert reflect.Kind to base kind
func ToBaseKind(kind reflect.Kind) BKind {
	return ToBKind(kind)
}

// ToBKind convert reflect.Kind to base kind
func ToBKind(kind reflect.Kind) BKind {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Int
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return Uint
	case reflect.Float32, reflect.Float64:
		return Float
	case reflect.Complex64, reflect.Complex128:
		return Complex
	case reflect.Array, reflect.Slice:
		return Array
	default:
		// like: string, map, struct, ptr, func, interface ...
		return BKind(kind)
	}
}

// Type struct
type Type interface {
	reflect.Type
	// BaseKind value
	BaseKind() BKind
}

type xType struct {
	reflect.Type
	baseKind BKind
}

// TypeOf value
func TypeOf(v any) Type {
	rftTyp := reflect.TypeOf(v)

	return &xType{
		Type:     rftTyp,
		baseKind: ToBKind(rftTyp.Kind()),
	}
}

// BaseKind value
func (t *xType) BaseKind() BKind {
	return t.baseKind
}
