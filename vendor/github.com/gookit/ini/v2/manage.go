package ini

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gookit/goutil/envutil"
	"github.com/mitchellh/mapstructure"
)

/*************************************************************
 * get config
 *************************************************************/

// GetValue get a value by key string.
// you can use '.' split for get value in a special section
func GetValue(key string) (string, bool) { return dc.GetValue(key) }

// GetValue a value by key string.
// you can use '.' split for get value in a special section
func (c *Ini) GetValue(key string) (val string, ok bool) {
	// if not is readonly
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
		// must close lock. because parseVarReference() maybe loop call Get()
		if !c.opts.Readonly {
			c.lock.Unlock()
			val = c.parseVarReference(key, val, strMap)
			c.lock.Lock()
		} else {
			val = c.parseVarReference(key, val, strMap)
		}
	}

	// if opts.ParseEnv is true. will parse like: "${SHELL}"
	if c.opts.ParseEnv {
		val = envutil.ParseEnvValue(val)
	}
	return
}

// Get get a value by key string.
// you can use '.' split for get value in a special section
func Get(key string, defVal ...string) string { return dc.Get(key, defVal...) }

// Get a value by key string.
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
// of following(case insensitive):
//  - true
//  - false
//  - yes
//  - no
//  - off
//  - on
//  - 0
//  - 1
// The `ok` boolean will be false in the event that the value could not be parsed as a bool
func (c *Ini) Bool(key string, defVal ...bool) (value bool) {
	rawVal, ok := c.GetValue(key)
	if !ok {
		if len(defVal) > 0 {
			return defVal[0]
		}
		return
	}

	lowerCase := strings.ToLower(rawVal)
	switch lowerCase {
	case "", "0", "false", "no", "off":
		value = false
	case "1", "true", "yes", "on":
		value = true
	default:
		c.addErrorf("the value '%s' cannot be convert to bool", lowerCase)
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

	if len(sep) > 0 {
		return stringToArray(str, sep[0])
	}
	return stringToArray(str, ",")
}

// StringMap get a section data map
func StringMap(name string) map[string]string { return dc.StringMap(name) }

// StringMap get a section data map
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

	// parser Var refer
	if c.opts.ParseVar {
		for k, v := range mp {
			mp[k] = c.parseVarReference(k, v, mp)
		}
	}

	// if opts.ParseEnv is true. will parse like: "${SHELL}"
	if c.opts.ParseEnv {
		for k, v := range mp {
			mp[k] = envutil.ParseEnvValue(v)
		}
	}
	return
}

// MapStruct get config data and binding to the structure.
func MapStruct(key string, ptr interface{}) error { return dc.MapStruct(key, ptr) }

// MapStruct get config data and binding to the structure.
// If the key is empty, will binding all data to the struct ptr.
//
// Usage:
// 	user := &Db{}
// 	ini.MapStruct("user", &user)
func (c *Ini) MapStruct(key string, ptr interface{}) error {
	// binding all data
	if key == "" {
		defSec := c.opts.DefSection
		if defMap, ok := c.data[defSec]; ok {
			data := make(map[string]interface{}, len(defMap)+len(c.data)-1)
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

		// no data of the default section
		return mapStruct(c.opts.TagName, c.data, ptr)
	}

	// parts data of the config
	data := c.StringMap(key)
	if len(data) == 0 {
		return errNotFound
	}

	return mapStruct(c.opts.TagName, data, ptr)
}

// MapTo mapping all data to struct pointer
func (c *Ini) MapTo(ptr interface{}) error {
	return c.MapStruct("", ptr)
}

func mapStruct(tagName string, data interface{}, ptr interface{}) error {
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
 * config set
 *************************************************************/

// Set a value to the section by key.
// if section is empty, will set to default section
func Set(key string, val interface{}, section ...string) error {
	return dc.Set(key, val, section...)
}

// Set a value to the section by key.
// if section is empty, will set to default section
func (c *Ini) Set(key string, val interface{}, section ...string) (err error) {
	// if is readonly
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
	name := c.opts.DefSection
	if len(section) > 0 {
		name = section[0]
	}

	strVal, isString := val.(string)
	if !isString {
		strVal = fmt.Sprint(val)
	}

	// allow section name is empty string ""
	name = c.formatKey(name)
	sec, ok := c.data[name]
	if ok {
		sec[key] = strVal
	} else {
		sec = Section{key: strVal}
	}

	c.data[name] = sec
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
	// open file
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)

	if err != nil {
		return 0, err
	}
	return c.WriteTo(fd)
}

// WriteTo out an INI File representing the current state to a writer.
func (c *Ini) WriteTo(out io.Writer) (n int64, err error) {
	n = 0
	counter := 0
	thisWrite := 0
	// section
	defaultSection := c.opts.DefSection
	orderedSections := make([]string, len(c.data))

	for section := range c.data {
		orderedSections[counter] = section
		counter++
	}

	sort.Strings(orderedSections)

	for _, section := range orderedSections {
		// don't add section title for DefSection
		if section != defaultSection {
			thisWrite, err = fmt.Fprintln(out, "["+section+"]")
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		items := c.data[section]
		orderedStringKeys := make([]string, len(items))
		counter = 0
		for key := range items {
			orderedStringKeys[counter] = key
			counter++
		}

		sort.Strings(orderedStringKeys)
		for _, key := range orderedStringKeys {
			thisWrite, err = fmt.Fprintln(out, key, "=", items[key])
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		thisWrite, err = fmt.Fprintln(out)
		n += int64(thisWrite)
		if err != nil {
			return
		}
	}
	return
}

/*************************************************************
 * section operate
 *************************************************************/

// Section get a section data map. is alias of StringMap()
func (c *Ini) Section(name string) Section {
	return c.StringMap(name)
}

// SetSection if not exist, add new section. If exist, will merge to old section.
func (c *Ini) SetSection(name string, values map[string]string) (err error) {
	// if is readonly
	if c.opts.Readonly {
		return errReadonly
	}

	name = c.formatKey(name)

	if old, ok := c.data[name]; ok {
		c.data[name] = mergeStringMap(values, old, c.opts.IgnoreCase)
	} else {
		if c.opts.IgnoreCase {
			values = mapKeyToLower(values)
		}
		c.data[name] = values
	}
	return
}

// NewSection add new section data, existed will be replace
func (c *Ini) NewSection(name string, values map[string]string) (err error) {
	// if is readonly
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

// HasSection has section
func (c *Ini) HasSection(name string) bool {
	name = c.formatKey(name)
	_, ok := c.data[name]
	return ok
}

// DelSection del section by name
func (c *Ini) DelSection(name string) (ok bool) {
	// if is readonly
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
func SectionKeys(withDefaultSection bool) (ls []string) {
	return dc.SectionKeys(withDefaultSection)
}

// SectionKeys get all section names
func (c *Ini) SectionKeys(withDefaultSection bool) (ls []string) {
	// default section name
	defaultSection := c.opts.DefSection

	for section := range c.data {
		if !withDefaultSection && section == defaultSection {
			continue
		}

		ls = append(ls, section)
	}
	return
}
