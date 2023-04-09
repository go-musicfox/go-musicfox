//go:build darwin

package core

import (
	"testing"
)

func TestNSArray(t *testing.T) {
	arr := NSArray_arrayWithObject(String("abc").NSObject)
	if arr.ID == 0 {
		panic("new arr failed")
	}
}
