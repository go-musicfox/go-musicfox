package comdef

// Int interface type
type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Uint interface type
type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Xint interface type. all int or uint types
type Xint interface {
	// equal: ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint32 | ~uint64
	Int | Uint
}

// Float interface type
type Float interface {
	~float32 | ~float64
}

// IntOrFloat interface type. all int and float types
type IntOrFloat interface {
	Int | Float
}

// XintOrFloat interface type. all (x)int and float types
type XintOrFloat interface {
	Int | Uint | Float
}

// ScalarType interface type.
//
// contains: (x)int, float, ~string, ~bool types
type ScalarType interface {
	Int | Uint | Float | ~string | ~bool
}
