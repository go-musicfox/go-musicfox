//go:build darwin

package core

import (
	"testing"
)

func TestNSMutableDictionary(t *testing.T) {
	dic := NSMutableDictionary_init()

	k1 := String("k1")
	defer k1.Release()

	v1 := String("v1")
	defer v1.Release()

	k2 := String("k2")
	defer k2.Release()

	v2 := NSNumber_numberWithInt(123)
	defer v2.Release()

	dic.SetValueForKey(k1, v1.NSObject)
	dic.SetValueForKey(k2, v2.NSObject)
	defer dic.Release()

	var res1 NSString
	defer res1.Release()
	res1.SetObjcID(dic.ValueForKey(k1))
	if res1.String() != v1.String() {
		panic("v1 value error")
	}

	var res2 NSNumber
	res2.SetObjcID(dic.ValueForKey(k2))
	if res2.IntValue() != v2.IntValue() {
		panic("v2 value error")
	}
}
