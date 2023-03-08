package maputil

import (
	"reflect"

	"github.com/gookit/goutil/reflects"
)

// HasKey check of the given map.
func HasKey(mp, key any) (ok bool) {
	rftVal := reflect.Indirect(reflect.ValueOf(mp))
	if rftVal.Kind() != reflect.Map {
		return
	}

	for _, keyRv := range rftVal.MapKeys() {
		if reflects.IsEqual(keyRv.Interface(), key) {
			return true
		}
	}
	return
}

// HasAllKeys check of the given map.
func HasAllKeys(mp any, keys ...any) (ok bool, noKey any) {
	rftVal := reflect.Indirect(reflect.ValueOf(mp))
	if rftVal.Kind() != reflect.Map {
		return
	}

	for _, key := range keys {
		var exist bool
		for _, keyRv := range rftVal.MapKeys() {
			if reflects.IsEqual(keyRv.Interface(), key) {
				exist = true
				break
			}
		}

		if !exist {
			return false, key
		}
	}

	return true, nil
}
