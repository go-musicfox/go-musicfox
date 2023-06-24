package maputil

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/strutil"
)

// KeyToLower convert keys to lower case.
func KeyToLower(src map[string]string) map[string]string {
	newMp := make(map[string]string, len(src))
	for k, v := range src {
		k = strings.ToLower(k)
		newMp[k] = v
	}
	return newMp
}

// ToStringMap convert map[string]any to map[string]string
func ToStringMap(src map[string]any) map[string]string {
	strMp := make(map[string]string, len(src))
	for k, v := range src {
		strMp[k] = strutil.SafeString(v)
	}
	return strMp
}

// CombineToSMap combine two string-slice to SMap(map[string]string)
func CombineToSMap(keys, values []string) SMap {
	return arrutil.CombineToSMap(keys, values)
}

// CombineToMap combine two any slice to map[K]V. alias of arrutil.CombineToMap
func CombineToMap[K comdef.SortedType, V any](keys []K, values []V) map[K]V {
	return arrutil.CombineToMap(keys, values)
}

// ToAnyMap convert map[TYPE1]TYPE2 to map[string]any
func ToAnyMap(mp any) map[string]any {
	amp, _ := TryAnyMap(mp)
	return amp
}

// TryAnyMap convert map[TYPE1]TYPE2 to map[string]any
func TryAnyMap(mp any) (map[string]any, error) {
	if aMp, ok := mp.(map[string]any); ok {
		return aMp, nil
	}

	rv := reflect.Indirect(reflect.ValueOf(mp))
	if rv.Kind() != reflect.Map {
		return nil, errors.New("input is not a map value")
	}

	anyMp := make(map[string]any, rv.Len())
	for _, key := range rv.MapKeys() {
		anyMp[key.String()] = rv.MapIndex(key).Interface()
	}
	return anyMp, nil
}

// HTTPQueryString convert map[string]any data to http query string.
func HTTPQueryString(data map[string]any) string {
	ss := make([]string, 0, len(data))
	for k, v := range data {
		ss = append(ss, k+"="+strutil.QuietString(v))
	}

	return strings.Join(ss, "&")
}

// StringsMapToAnyMap convert map[string][]string to map[string]any
//
//	Example:
//	{"k1": []string{"v1", "v2"}, "k2": []string{"v3"}}
//	=>
//	{"k": []string{"v1", "v2"}, "k2": "v3"}
//
//	mp := StringsMapToAnyMap(httpReq.Header)
func StringsMapToAnyMap(ssMp map[string][]string) map[string]any {
	if len(ssMp) == 0 {
		return nil
	}

	anyMp := make(map[string]any, len(ssMp))
	for k, v := range ssMp {
		if len(v) == 1 {
			anyMp[k] = v[0]
			continue
		}
		anyMp[k] = v
	}
	return anyMp
}

// ToString simple and quickly convert map[string]any to string.
func ToString(mp map[string]any) string {
	if mp == nil {
		return ""
	}
	if len(mp) == 0 {
		return "{}"
	}

	buf := make([]byte, 0, len(mp)*16)
	buf = append(buf, '{')

	for k, val := range mp {
		buf = append(buf, k...)
		buf = append(buf, ':')

		str := strutil.QuietString(val)
		buf = append(buf, str...)
		buf = append(buf, ',', ' ')
	}

	// remove last ', '
	buf = append(buf[:len(buf)-2], '}')
	return strutil.Byte2str(buf)
}

// ToString2 simple and quickly convert a map to string.
func ToString2(mp any) string {
	return NewFormatter(mp).Format()
}

// FormatIndent format map data to string with newline and indent.
func FormatIndent(mp any, indent string) string {
	return NewFormatter(mp).WithIndent(indent).Format()
}

/*************************************************************
 * Flat convert tree map to flatten key-value map.
 *************************************************************/

// Flatten convert tree map to flat key-value map.
//
// Examples:
//
//	{"top": {"sub": "value", "sub2": "value2"} }
//	->
//	{"top.sub": "value", "top.sub2": "value2" }
func Flatten(mp map[string]any) map[string]any {
	if mp == nil {
		return nil
	}

	flatMp := make(map[string]any, len(mp)*2)
	reflects.FlatMap(reflect.ValueOf(mp), func(path string, val reflect.Value) {
		flatMp[path] = val.Interface()
	})

	return flatMp
}

// FlatWithFunc flat a tree-map with custom collect handle func
func FlatWithFunc(mp map[string]any, fn reflects.FlatFunc) {
	if mp == nil || fn == nil {
		return
	}
	reflects.FlatMap(reflect.ValueOf(mp), fn)
}
