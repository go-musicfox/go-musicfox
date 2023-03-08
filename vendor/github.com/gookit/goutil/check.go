package goutil

import (
	"reflect"

	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/stdutil"
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
	_, found := stdutil.CheckContains(data, elem)
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
	_, found := stdutil.CheckContains(data, elem)
	return found
}
