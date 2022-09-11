package goutil

import (
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// Value data store
type Value struct {
	// V value
	V interface{}
}

// Reset value
func (v *Value) Reset() {
	v.V = nil
}

// Val get
func (v Value) Val() interface{} {
	return v.V
}

// Int value
func (v Value) Int() int {
	if v.V == nil {
		return 0
	}

	return mathutil.MustInt(v.V)
}

// Int64 value
func (v Value) Int64() int64 {
	if v.V == nil {
		return 0
	}

	return mathutil.MustInt64(v.V)
}

// Bool value
func (v Value) Bool() bool {
	if v.V == nil {
		return false
	}

	if bl, ok := v.V.(bool); ok {
		return bl
	}

	if str, ok := v.V.(string); ok {
		return strutil.MustBool(str)
	}
	return false
}

// Float64 value
func (v Value) Float64() float64 {
	if v.V == nil {
		return 0
	}

	return mathutil.MustFloat(v.V)
}

// String value
func (v Value) String() string {
	if v.V == nil {
		return ""
	}

	if str, ok := v.V.(string); ok {
		return str
	}

	return strutil.MustString(v.V)
}

// Strings value
func (v Value) Strings() (ss []string) {
	if v.V == nil {
		return
	}

	if ss, ok := v.V.([]string); ok {
		return ss
	}
	return
}

// IsEmpty value
func (v Value) IsEmpty() bool {
	return v.V == nil
}
