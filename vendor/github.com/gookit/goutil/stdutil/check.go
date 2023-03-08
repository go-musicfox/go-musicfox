package stdutil

import (
	"reflect"
	"strings"

	"github.com/gookit/goutil/reflects"
)

// IsNil value check
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	return reflects.IsNil(reflect.ValueOf(v))
}

// IsEmpty value check
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}
	return reflects.IsEmpty(reflect.ValueOf(v))
}

// IsFunc value
func IsFunc(val any) bool {
	if val == nil {
		return false
	}
	return reflect.TypeOf(val).Kind() == reflect.Func
}

// IsEqual determines if two objects are considered equal.
//
// TIP: cannot compare function type
func IsEqual(src, dst any) bool {
	if src == nil || dst == nil {
		return src == dst
	}

	// cannot compare function type
	if IsFunc(src) || IsFunc(dst) {
		return false
	}
	return reflects.IsEqual(src, dst)
}

// Contains try loop over the data check if the data includes the element.
// alias of the IsContains
//
// TIP: only support types: string, map, array, slice
//
//	map         - check key exists
//	string 	    - check sub-string exists
//	array,slice - check sub-element exists
func Contains(data, elem any) bool {
	_, found := CheckContains(data, elem)
	return found
}

// IsContains try loop over the data check if the data includes the element.
//
// TIP: only support types: string, map, array, slice
//
//	map         - check key exists
//	string 	    - check sub-string exists
//	array,slice - check sub-element exists
func IsContains(data, elem any) bool {
	_, found := CheckContains(data, elem)
	return found
}

// CheckContains try loop over the data check if the data includes the element.
//
// TIP: only support types: string, map, array, slice
//
//	map         - check key exists
//	string 	    - check sub-string exists
//	array,slice - check sub-element exists
//
// return (false, false) if impossible.
// return (true, false) if element was not found.
// return (true, true) if element was found.
func CheckContains(data, elem any) (valid, found bool) {
	dataRv := reflect.ValueOf(data)
	dataRt := reflect.TypeOf(data)
	if dataRt == nil {
		return false, false
	}

	dataKind := dataRt.Kind()

	// string
	if dataKind == reflect.String {
		return true, strings.Contains(dataRv.String(), reflect.ValueOf(elem).String())
	}

	// map
	if dataKind == reflect.Map {
		mapKeys := dataRv.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if reflects.IsEqual(mapKeys[i].Interface(), elem) {
				return true, true
			}
		}
		return true, false
	}

	// array, slice - other return false
	if dataKind != reflect.Slice && dataKind != reflect.Array {
		return false, false
	}

	for i := 0; i < dataRv.Len(); i++ {
		if reflects.IsEqual(dataRv.Index(i).Interface(), elem) {
			return true, true
		}
	}
	return true, false
}

// ValueIsEmpty reflect value check.
//
// Deprecated: please use reflects.IsEmpty()
func ValueIsEmpty(v reflect.Value) bool {
	return reflects.IsEmpty(v)
}

// ValueLen get reflect value length
//
// Deprecated: please use reflects.Len()
func ValueLen(v reflect.Value) int {
	return reflects.Len(v)
}
