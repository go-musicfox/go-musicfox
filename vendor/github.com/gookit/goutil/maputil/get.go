package maputil

import (
	"reflect"
	"strconv"
	"strings"
)

// some consts for separators
const (
	Wildcard = "*"
	PathSep  = "."
)

// DeepGet value by key path. eg "top" "top.sub"
func DeepGet(mp map[string]any, path string) (val any) {
	val, _ = GetByPath(path, mp)
	return
}

// QuietGet value by key path. eg "top" "top.sub"
func QuietGet(mp map[string]any, path string) (val any) {
	val, _ = GetByPath(path, mp)
	return
}

// GetByPath get value by key path from a map(map[string]any). eg "top" "top.sub"
func GetByPath(path string, mp map[string]any) (val any, ok bool) {
	if val, ok := mp[path]; ok {
		return val, true
	}

	// no sub key
	if len(mp) == 0 || strings.IndexByte(path, '.') < 1 {
		return nil, false
	}

	// has sub key. eg. "top.sub"
	keys := strings.Split(path, ".")
	return GetByPathKeys(mp, keys)
}

// GetByPathKeys get value by path keys from a map(map[string]any). eg "top" "top.sub"
//
// Example:
//
//	mp := map[string]any{
//		"top": map[string]any{
//			"sub": "value",
//		},
//	}
//	val, ok := GetByPathKeys(mp, []string{"top", "sub"}) // return "value", true
func GetByPathKeys(mp map[string]any, keys []string) (val any, ok bool) {
	kl := len(keys)
	if kl == 0 {
		return mp, true
	}

	// find top item data use top key
	var item any

	topK := keys[0]
	if item, ok = mp[topK]; !ok {
		return
	}

	// find sub item data use sub key
	for i, k := range keys[1:] {
		switch tData := item.(type) {
		case map[string]string: // is string map
			if item, ok = tData[k]; !ok {
				return
			}
		case map[string]any: // is map(decode from toml/json/yaml)
			if item, ok = tData[k]; !ok {
				return
			}
		case map[any]any: // is map(decode from yaml.v2)
			if item, ok = tData[k]; !ok {
				return
			}
		case []map[string]any: // is an any-map slice
			if k == Wildcard {
				if kl == i+2 {
					return tData, true
				}

				sl := make([]any, 0, len(tData))
				for _, v := range tData {
					if val, ok = GetByPathKeys(v, keys[i+2:]); ok {
						sl = append(sl, val)
					}
				}
				return sl, true
			}

			// k is index number
			idx, err := strconv.Atoi(k)
			if err != nil {
				return nil, false
			}

			if idx >= len(tData) {
				return nil, false
			}
			item = tData[idx]
		default:
			rv := reflect.ValueOf(tData)
			// check is slice
			if rv.Kind() == reflect.Slice {
				i, err := strconv.Atoi(k)
				if err != nil {
					return nil, false
				}
				if i >= rv.Len() {
					return nil, false
				}

				item = rv.Index(i).Interface()
				continue
			}

			// as error
			return nil, false
		}
	}

	return item, true
}

// Keys get all keys of the given map.
func Keys(mp any) (keys []string) {
	rftVal := reflect.Indirect(reflect.ValueOf(mp))
	if rftVal.Kind() != reflect.Map {
		return
	}

	keys = make([]string, 0, rftVal.Len())
	for _, key := range rftVal.MapKeys() {
		keys = append(keys, key.String())
	}
	return
}

// Values get all values from the given map.
func Values(mp any) (values []any) {
	rv := reflect.Indirect(reflect.ValueOf(mp))
	if rv.Kind() != reflect.Map {
		return
	}

	values = make([]any, 0, rv.Len())
	for _, key := range rv.MapKeys() {
		values = append(values, rv.MapIndex(key).Interface())
	}
	return
}

// EachAnyMap iterates the given map and calls the given function for each item.
func EachAnyMap(mp any, fn func(key string, val any)) {
	rv := reflect.Indirect(reflect.ValueOf(mp))
	if rv.Kind() != reflect.Map {
		panic("not a map value")
	}

	for _, key := range rv.MapKeys() {
		fn(key.String(), rv.MapIndex(key).Interface())
	}
}
