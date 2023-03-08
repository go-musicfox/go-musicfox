package structs

import (
	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// Value data store
type Value struct {
	// V value
	V any
}

// NewValue instance.
func NewValue(val any) *Value {
	return &Value{
		V: val,
	}
}

// Set value
func (v *Value) Set(val any) {
	v.V = val
}

// Reset value
func (v *Value) Reset() {
	v.V = nil
}

// Val get
func (v *Value) Val() any {
	return v.V
}

// Val get
// func (v *Value) ValOr[T any](defVal T) T {
// 	return v.V
// }

// Int value get
func (v *Value) Int() int {
	if v.V == nil {
		return 0
	}
	return mathutil.QuietInt(v.V)
}

// Int64 value
func (v *Value) Int64() int64 {
	if v.V == nil {
		return 0
	}
	return mathutil.QuietInt64(v.V)
}

// Bool value
func (v *Value) Bool() bool {
	if v.V == nil {
		return false
	}

	if bl, ok := v.V.(bool); ok {
		return bl
	}

	if str, ok := v.V.(string); ok {
		return strutil.QuietBool(str)
	}
	return false
}

// Float64 value
func (v *Value) Float64() float64 {
	if v.V == nil {
		return 0
	}
	return mathutil.QuietFloat(v.V)
}

// String value
func (v *Value) String() string {
	if v.V == nil {
		return ""
	}

	if str, ok := v.V.(string); ok {
		return str
	}
	return strutil.QuietString(v.V)
}

// Strings value
func (v *Value) Strings() (ss []string) {
	if v.V == nil {
		return
	}

	if ss, ok := v.V.([]string); ok {
		return ss
	}
	if str, ok := v.V.(string); ok {
		return strutil.Split(str, comdef.DefaultSep)
	}
	return
}

// SplitToStrings split string value to strings
func (v *Value) SplitToStrings(sep ...string) (ss []string) {
	if v.V == nil {
		return
	}

	if str, ok := v.V.(string); ok {
		return strutil.Split(str, sepStr(sep))
	}
	return
}

// SplitToInts split string value to []int
func (v *Value) SplitToInts(sep ...string) (ss []int) {
	if v.V == nil {
		return
	}

	if str, ok := v.V.(string); ok {
		ints, err := arrutil.StringsToInts(strutil.Split(str, sepStr(sep)))
		if err == nil {
			return ints
		}
	}
	return
}

// IsEmpty value
func (v *Value) IsEmpty() bool {
	return v.V == nil
}

func sepStr(seps []string) string {
	if len(seps) > 0 {
		return seps[0]
	}
	return comdef.DefaultSep
}
