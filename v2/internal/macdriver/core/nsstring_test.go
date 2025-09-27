//go:build darwin

package core

import (
	"testing"
)

func TestNSString(t *testing.T) {
	url := String("abcdefg")
	defer url.Release()
	if url.String() != "abcdefg" {
		panic("string is wrong")
	}
}
