package structs

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gookit/goutil/maputil"
)

// ToMap quickly convert structs to map by reflect
func ToMap(st any, optFns ...MapOptFunc) map[string]any {
	mp, _ := StructToMap(st, optFns...)
	return mp
}

// MustToMap alis of TryToMap, but will panic on error
func MustToMap(st any, optFns ...MapOptFunc) map[string]any {
	mp, err := StructToMap(st, optFns...)
	if err != nil {
		panic(err)
	}
	return mp
}

// TryToMap simple convert structs to map by reflect
func TryToMap(st any, optFns ...MapOptFunc) (map[string]any, error) {
	return StructToMap(st, optFns...)
}

// ToString format
func ToString(st any, optFns ...MapOptFunc) string {
	mp, err := StructToMap(st, optFns...)
	if err == nil {
		return maputil.ToString(mp)
	}
	return fmt.Sprint(st)
}

const defaultFieldTag = "json"

// MapOptions struct
type MapOptions struct {
	TagName string
}

// MapOptFunc define
type MapOptFunc func(opt *MapOptions)

// StructToMap quickly convert structs to map[string]any by reflect.
// Can custom export field name by tag `json` or custom tag
func StructToMap(st any, optFns ...MapOptFunc) (map[string]any, error) {
	mp := make(map[string]any)
	if st == nil {
		return mp, nil
	}

	obj := reflect.ValueOf(st)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}

	if obj.Kind() != reflect.Struct {
		return mp, errors.New("must be an struct")
	}

	opt := &MapOptions{TagName: defaultFieldTag}
	for _, fn := range optFns {
		fn(opt)
	}

	mp, err := structToMap(obj, opt.TagName)
	return mp, err
}

func structToMap(obj reflect.Value, tagName string) (map[string]any, error) {
	refType := obj.Type()
	mp := make(map[string]any)

	for i := 0; i < obj.NumField(); i++ {
		ft := refType.Field(i)
		name := ft.Name
		// skip don't exported field
		if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}

		tagVal, ok := ft.Tag.Lookup(tagName)
		if ok && tagVal != "" {
			sMap, err := ParseTagValueDefault(name, tagVal)
			if err != nil {
				return nil, err
			}

			name = sMap.Default("name", name)
			// un-exported field
			if name == "" {
				continue
			}
		}

		field := obj.Field(i)
		if field.Kind() == reflect.Struct {
			sub, err := structToMap(field, tagName)
			if err != nil {
				return nil, err
			}
			mp[name] = sub
			continue
		}

		if field.CanInterface() {
			mp[name] = field.Interface()
		}
	}

	return mp, nil
}
