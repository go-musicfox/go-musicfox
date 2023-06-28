package ini

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/goutil/envutil"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/ini/v2/parser"
	"github.com/mitchellh/mapstructure"
)

/*************************************************************
 * read config value
 *************************************************************/

// GetValue get a value by key string.
// you can use '.' split for get value in a special section
func GetValue(key string) (string, bool) { return dc.GetValue(key) }

// GetValue a value by key string.
//
// you can use '.' split for get value in a special section
func (c *Ini) GetValue(key string) (val string, ok bool) {
	if !c.opts.Readonly {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	if key = c.formatKey(key); key == "" {
		return
	}

	// get section data
	name, key := c.splitSectionAndKey(key)
	strMap, ok := c.data[name]
	if !ok {
		return
	}

	val, ok = strMap[key]

	// if enable parse var refer
	if c.opts.ParseVar {
		val = c.parseVarReference(key, val, strMap)
	}

	// if opts.ParseEnv is true. will parse like: "${SHELL}"
	if c.opts.ParseEnv {
		val = envutil.ParseEnvValue(val)
	}
	return
}

func (c *Ini) getValue(key string) (val string, ok bool) {
	if key = c.formatKey(key); key == "" {
		return
	}

	// get section data
	name, key := c.splitSectionAndKey(key)

	// get value
	if strMap, has := c.data[name]; has {
		val, ok = strMap[key]
	}
	return
}

// Get a value by key string.
// you can use '.' split for get value in a special section
func Get(key string, defVal ...string) string { return dc.Get(key, defVal...) }

// Get a value by key string.
//
// you can use '.' split for get value in a special section
func (c *Ini) Get(key string, defVal ...string) string {
	value, ok := c.GetValue(key)

	if !ok && len(defVal) > 0 {
		value = defVal[0]
	}
	return value
}

// String get a string by key
func String(key string, defVal ...string) string { return dc.String(key, defVal...) }

// String like Get method
func (c *Ini) String(key string, defVal ...string) string {
	return c.Get(key, defVal...)
}

// Int get a int by key
func Int(key string, defVal ...int) int { return dc.Int(key, defVal...) }

// Int get a int value, if not found return default value
func (c *Ini) Int(key string, defVal ...int) (value int) {
	i64, exist := c.tryInt64(key)

	if exist {
		value = int(i64)
	} else if len(defVal) > 0 {
		value = defVal[0]
	}
	return
}

// Uint get a uint value, if not found return default value
func Uint(key string, defVal ...uint) uint { return dc.Uint(key, defVal...) }

// Uint get a int value, if not found return default value
func (c *Ini) Uint(key string, defVal ...uint) (value uint) {
	i64, exist := c.tryInt64(key)

	if exist {
		value = uint(i64)
	} else if len(defVal) > 0 {
		value = defVal[0]
	}
	return
}

// Int64 get a int value, if not found return default value
func Int64(key string, defVal ...int64) int64 { return dc.Int64(key, defVal...) }

// Int64 get a int value, if not found return default value
func (c *Ini) Int64(key string, defVal ...int64) (value int64) {
	value, exist := c.tryInt64(key)

	if !exist && len(defVal) > 0 {
		value = defVal[0]
	}
	return
}

// try get a int64 value by given key
func (c *Ini) tryInt64(key string) (value int64, ok bool) {
	strVal, ok := c.GetValue(key)
	if !ok {
		return
	}

	value, err := strconv.ParseInt(strVal, 10, 0)
	if err != nil {
		c.err = err
	}
	return
}

// Bool get a bool value, if not found return default value
func Bool(key string, defVal ...bool) bool { return dc.Bool(key, defVal...) }

// Bool Looks up a value for a key in this section and attempts to parse that value as a boolean,
// along with a boolean result similar to a map lookup.
//
// The `value` boolean will be false in the event that the value could not be parsed as a bool
func (c *Ini) Bool(key string, defVal ...bool) (value bool) {
	rawVal, ok := c.GetValue(key)
	if !ok {
		if len(defVal) > 0 {
			return defVal[0]
		}
		return
	}

	var err error
	value, err = strutil.ToBool(rawVal)
	if err != nil {
		c.err = err
	}

	return
}

// Strings get a string array, by split a string
func Strings(key string, sep ...string) []string { return dc.Strings(key, sep...) }

// Strings get a string array, by split a string
func (c *Ini) Strings(key string, sep ...string) (ss []string) {
	str, ok := c.GetValue(key)
	if !ok {
		return
	}

	sepChar := ","
	if len(sep) > 0 {
		sepChar = sep[0]
	}
	return strutil.Split(str, sepChar)
}

// StringMap get a section data map
func StringMap(name string) map[string]string { return dc.StringMap(name) }

// Section get a section data map. is alias of StringMap()
func (c *Ini) Section(name string) Section { return c.StringMap(name) }

// StringMap get a section data map by name
func (c *Ini) StringMap(name string) (mp map[string]string) {
	name = c.formatKey(name)
	// empty name, return default section
	if name == "" {
		name = c.opts.DefSection
	}

	mp, ok := c.data[name]
	if !ok {
		return
	}

	if c.opts.ParseVar || c.opts.ParseEnv {
		for k, v := range mp {
			// parser Var refer
			if c.opts.ParseVar {
				v = c.parseVarReference(k, v, mp)
			}

			// parse ENV. like: "${SHELL}"
			if c.opts.ParseEnv {
				v = envutil.ParseEnvValue(v)
			}

			mp[k] = v
		}
	}

	return
}

// MapStruct get config data and binding to the structure.
func MapStruct(key string, ptr any) error { return dc.MapStruct(key, ptr) }

// Decode all data to struct pointer
func (c *Ini) Decode(ptr any) error { return c.MapStruct("", ptr) }

// MapTo mapping all data to struct pointer.
//
// Deprecated: please use Decode()
func (c *Ini) MapTo(ptr any) error { return c.MapStruct("", ptr) }

// MapStruct get config data and binding to the structure.
// If the key is empty, will bind all data to the struct ptr.
//
// Usage:
//
//	user := &Db{}
//	ini.MapStruct("user", &user)
func (c *Ini) MapStruct(key string, ptr any) error {
	// binding all data
	if key == "" {
		defSec := c.opts.DefSection
		if defMap, ok := c.data[defSec]; ok {
			data := make(map[string]any, len(defMap)+len(c.data)-1)
			for key, val := range defMap {
				data[key] = val
			}

			for secKey, secVals := range c.data {
				if secKey != defSec {
					data[secKey] = secVals
				}
			}
			return mapStruct(c.opts.TagName, data, ptr)
		}

		// no default section
		return mapStruct(c.opts.TagName, c.data, ptr)
	}

	// parts data of the config
	data := c.StringMap(key)
	if len(data) == 0 {
		return errNotFound
	}
	return mapStruct(c.opts.TagName, data, ptr)
}

func mapStruct(tagName string, data any, ptr any) error {
	mapConf := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   ptr,
		TagName:  tagName,
		// will auto convert string to int/uint
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(mapConf)
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}

/*************************************************************
 * write config value
 *************************************************************/

// Set a value to the section by key.
//
// if section is empty, will set to default section
func Set(key string, val any, section ...string) error {
	return dc.Set(key, val, section...)
}

// Set a value to the section by key.
//
// if section is empty, will set to default section
func (c *Ini) Set(key string, val any, section ...string) (err error) {
	if c.opts.Readonly {
		return errReadonly
	}

	c.ensureInit()
	c.lock.Lock()
	defer c.lock.Unlock()

	key = c.formatKey(key)
	if key == "" {
		return errEmptyKey
	}

	// section name
	group := c.opts.DefSection
	if len(section) > 0 && section[0] != "" {
		group = section[0]
	}

	// allow section name is empty string ""
	group = c.formatKey(group)
	strVal := strutil.QuietString(val)

	sec, ok := c.data[group]
	if ok {
		sec[key] = strVal
	} else {
		sec = Section{key: strVal}
	}

	c.data[group] = sec
	return
}

// SetSection if not exist, add new section. If existed, will merge to old section.
func (c *Ini) SetSection(name string, values map[string]string) (err error) {
	if c.opts.Readonly {
		return errReadonly
	}

	name = c.formatKey(name)
	if old, ok := c.data[name]; ok {
		c.data[name] = maputil.MergeStringMap(values, old, c.opts.IgnoreCase)
		return
	}

	if c.opts.IgnoreCase {
		values = mapKeyToLower(values)
	}
	c.data[name] = values
	return
}

// NewSection add new section data, existed will be replaced
func (c *Ini) NewSection(name string, values map[string]string) (err error) {
	if c.opts.Readonly {
		return errReadonly
	}

	if c.opts.IgnoreCase {
		name = strings.ToLower(name)
		c.data[name] = mapKeyToLower(values)
	} else {
		c.data[name] = values
	}
	return
}

/*************************************************************
 * config dump
 *************************************************************/

// PrettyJSON translate to pretty JSON string
func (c *Ini) PrettyJSON() string {
	if len(c.data) == 0 {
		return ""
	}

	out, _ := json.MarshalIndent(c.data, "", "    ")
	return string(out)
}

// WriteToFile write config data to a file
func (c *Ini) WriteToFile(file string) (int64, error) {
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return 0, err
	}

	defer fd.Close()
	return c.WriteTo(fd)
}

// WriteTo out an INI File representing the current state to a writer.
func (c *Ini) WriteTo(out io.Writer) (n int64, err error) {
	mp := make(map[string]map[string]string, len(c.data))
	for group, secMp := range c.data {
		mp[group] = secMp
	}

	bs, err := parser.EncodeLite(mp, c.opts.DefSection)
	if err != nil {
		return 0, err
	}

	var ni int
	ni, err = out.Write(bs)
	return int64(ni), err
}

/*************************************************************
 * section operate
 *************************************************************/

// HasSection has section
func (c *Ini) HasSection(name string) bool {
	name = c.formatKey(name)
	_, ok := c.data[name]
	return ok
}

// DelSection del section by name
func (c *Ini) DelSection(name string) (ok bool) {
	if c.opts.Readonly {
		return
	}

	name = c.formatKey(name)
	if _, ok = c.data[name]; ok {
		delete(c.data, name)
	}
	return
}

// SectionKeys get all section names
func SectionKeys(withDefSection bool) (ls []string) {
	return dc.SectionKeys(withDefSection)
}

// SectionKeys get all section names
func (c *Ini) SectionKeys(withDefSection bool) (ls []string) {
	defaultSection := c.opts.DefSection

	for section := range c.data {
		if !withDefSection && section == defaultSection {
			continue
		}

		ls = append(ls, section)
	}
	return
}
