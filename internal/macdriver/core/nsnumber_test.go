//go:build darwin

package core

import (
	"testing"
)

func TestNSNumber(t *testing.T) {
	i := NSNumber_numberWithInt(123)
	defer i.Release()
	if i.IntValue() != 123 {
		panic("int value is wrong")
	}

	d := NSNumber_numberWithDouble(123.4)
	defer d.Release()
	if d.DoubleValue() != 123.4 {
		panic("double value is wrong")
	}
}
