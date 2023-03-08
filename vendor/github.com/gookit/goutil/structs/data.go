package structs

import (
	"sync"

	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// LiteData simple map[string]any struct. no lock
type LiteData struct {
	data map[string]any
}

// Data get all
func (d *LiteData) Data() map[string]any {
	return d.data
}

// SetData set all data
func (d *LiteData) SetData(data map[string]any) {
	d.data = data
}

// Value get from data
func (d *LiteData) Value(key string) any {
	return d.data[key]
}

// GetVal get from data
func (d *LiteData) GetVal(key string) any {
	return d.data[key]
}

// StrValue get from data
func (d *LiteData) StrValue(key string) string {
	return strutil.QuietString(d.data[key])
}

// IntVal get from data
func (d *LiteData) IntVal(key string) int {
	return mathutil.QuietInt(d.data[key])
}

// SetValue to data
func (d *LiteData) SetValue(key string, val any) {
	if d.data == nil {
		d.data = make(map[string]any)
	}
	d.data[key] = val
}

// ResetData all data
func (d *LiteData) ResetData() {
	d.data = nil
}

/*************************************************************
 * data struct and allow enable lock
 *************************************************************/

// Data struct, allow enable lock TODO
type Data struct {
	sync.RWMutex
	enableLock bool
	// data store
	data map[string]any
}

// NewData create
func NewData() *Data {
	return &Data{
		data: make(map[string]any),
	}
}

// EnableLock for operate data
func (d *Data) EnableLock() *Data {
	d.enableLock = true
	return d
}

// Data get all
func (d *Data) Data() map[string]any {
	return d.data
}

// SetData set all data
func (d *Data) SetData(data map[string]any) {
	if !d.enableLock {
		d.data = data
		return
	}

	d.RLock()
	d.data = data
	d.RUnlock()
}

// DataLen of data
func (d *Data) DataLen() int {
	return len(d.data)
}

// ResetData all data
func (d *Data) ResetData() {
	d.data = make(map[string]any)
}

// Set value to data
func (d *Data) Set(key string, val any) {
	d.SetValue(key, val)
}

// SetValue to data
func (d *Data) SetValue(key string, val any) {
	if d.enableLock {
		d.Lock()
		defer d.Unlock()
	}

	d.data[key] = val
}

// Value get from data
func (d *Data) Value(key string) (val any, ok bool) {
	if d.enableLock {
		d.RLock()
		defer d.RUnlock()
	}

	val, ok = d.data[key]
	return
}

// Get val from data
func (d *Data) Get(key string) any {
	return d.GetVal(key)
}

// GetVal get from data
func (d *Data) GetVal(key string) any {
	if d.enableLock {
		d.RLock()
		defer d.RUnlock()
	}

	return d.data[key]
}

// StrVal get from data
func (d *Data) StrVal(key string) string {
	return strutil.QuietString(d.GetVal(key))
}

// IntVal get from data
func (d *Data) IntVal(key string) int {
	return mathutil.QuietInt(d.GetVal(key))
}

// BoolVal get from data
func (d *Data) BoolVal(key string) bool {
	val, ok := d.Value(key)
	if !ok {
		return false
	}
	return comfunc.Bool(val)
}

// String format data
func (d *Data) String() string {
	return maputil.ToString(d.data)
}

// OrderedMap data TODO
type OrderedMap struct {
	maputil.Data
	len  int
	keys []string
	// vals []any
}

// NewOrderedMap instance.
func NewOrderedMap(len int) *OrderedMap {
	return &OrderedMap{len: len}
}

// Set key and value to map
func (om *OrderedMap) Set(key string, val any) {
	om.keys = append(om.keys, key)
	om.Data.Set(key, val)
}
