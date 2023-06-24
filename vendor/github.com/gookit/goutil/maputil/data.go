package maputil

import (
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// Data an map data type
type Data map[string]any

// Map alias of Data
type Map = Data

// Has value on the data map
func (d Data) Has(key string) bool {
	_, ok := d.GetByPath(key)
	return ok
}

// IsEmtpy if the data map
func (d Data) IsEmtpy() bool {
	return len(d) == 0
}

// Value get from the data map
func (d Data) Value(key string) (any, bool) {
	val, ok := d.GetByPath(key)
	return val, ok
}

// Get value from the data map.
// Supports dot syntax to get deep values. eg: top.sub
func (d Data) Get(key string) any {
	if val, ok := d.GetByPath(key); ok {
		return val
	}
	return nil
}

// GetByPath get value from the data map by path. eg: top.sub
// Supports dot syntax to get deep values.
func (d Data) GetByPath(path string) (any, bool) {
	if val, ok := d[path]; ok {
		return val, true
	}

	// is key path.
	if strings.ContainsRune(path, '.') {
		val, ok := GetByPath(path, d)
		if ok {
			return val, true
		}
	}
	return nil, false
}

// Set value to the data map
func (d Data) Set(key string, val any) {
	d[key] = val
}

// SetByPath sets a value in the map.
// Supports dot syntax to set deep values.
//
// For example:
//
//	d.SetByPath("name.first", "Mat")
func (d Data) SetByPath(path string, value any) error {
	if path == "" {
		return nil
	}
	return d.SetByKeys(strings.Split(path, KeySepStr), value)
}

// SetByKeys sets a value in the map by path keys.
// Supports dot syntax to set deep values.
//
// For example:
//
//	d.SetByKeys([]string{"name", "first"}, "Mat")
func (d Data) SetByKeys(keys []string, value any) error {
	kln := len(keys)
	if kln == 0 {
		return nil
	}

	// special handle d is empty.
	if len(d) == 0 {
		if kln == 1 {
			d.Set(keys[0], value)
		} else {
			d.Set(keys[0], MakeByKeys(keys[1:], value))
		}
		return nil
	}

	return SetByKeys((*map[string]any)(&d), keys, value)
	// It's ok, but use `func (d *Data)`
	// return SetByKeys((*map[string]any)(d), keys, value)
}

// Default get value from the data map with default value
func (d Data) Default(key string, def any) any {
	if val, ok := d.GetByPath(key); ok {
		return val
	}
	return def
}

// Int value get
func (d Data) Int(key string) int {
	if val, ok := d.GetByPath(key); ok {
		return mathutil.QuietInt(val)
	}
	return 0
}

// Int64 value get
func (d Data) Int64(key string) int64 {
	if val, ok := d.GetByPath(key); ok {
		return mathutil.QuietInt64(val)
	}
	return 0
}

// Uint value get
func (d Data) Uint(key string) uint64 {
	if val, ok := d.GetByPath(key); ok {
		return mathutil.QuietUint(val)
	}
	return 0
}

// Str value get by key
func (d Data) Str(key string) string {
	if val, ok := d.GetByPath(key); ok {
		return strutil.QuietString(val)
	}
	return ""
}

// Bool value get
func (d Data) Bool(key string) bool {
	val, ok := d.GetByPath(key)
	if !ok {
		return false
	}

	switch tv := val.(type) {
	case string:
		return strutil.QuietBool(tv)
	case bool:
		return tv
	default:
		return false
	}
}

// Strings get []string value
func (d Data) Strings(key string) []string {
	val, ok := d.GetByPath(key)
	if !ok {
		return nil
	}

	switch typVal := val.(type) {
	case string:
		return []string{typVal}
	case []string:
		return typVal
	case []any:
		return arrutil.SliceToStrings(typVal)
	default:
		return nil
	}
}

// StrSplit get strings by split key value
func (d Data) StrSplit(key, sep string) []string {
	if val, ok := d.GetByPath(key); ok {
		return strings.Split(strutil.QuietString(val), sep)
	}
	return nil
}

// StringsByStr value get by key
func (d Data) StringsByStr(key string) []string {
	if val, ok := d.GetByPath(key); ok {
		return strings.Split(strutil.QuietString(val), ",")
	}
	return nil
}

// StrMap get map[string]string value
func (d Data) StrMap(key string) map[string]string {
	return d.StringMap(key)
}

// StringMap get map[string]string value
func (d Data) StringMap(key string) map[string]string {
	val, ok := d.GetByPath(key)
	if !ok {
		return nil
	}

	switch tv := val.(type) {
	case map[string]string:
		return tv
	case map[string]any:
		return ToStringMap(tv)
	default:
		return nil
	}
}

// Sub get sub value as new Data
func (d Data) Sub(key string) Data {
	if val, ok := d.GetByPath(key); ok {
		if sub, ok := val.(map[string]any); ok {
			return sub
		}
	}
	return nil
}

// Keys of the data map
func (d Data) Keys() []string {
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	return keys
}

// ToStringMap convert to map[string]string
func (d Data) ToStringMap() map[string]string {
	return ToStringMap(d)
}

// String data to string
func (d Data) String() string {
	return ToString(d)
}

// Load other data to current data map
func (d Data) Load(sub map[string]any) {
	for name, val := range sub {
		d[name] = val
	}
}

// LoadSMap to data
func (d Data) LoadSMap(smp map[string]string) {
	for name, val := range smp {
		d[name] = val
	}
}
