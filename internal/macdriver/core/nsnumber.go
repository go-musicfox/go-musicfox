//go:build darwin

package core

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_NSNumber = objc.GetClass("NSNumber")
}

var (
	class_NSNumber objc.Class
)

var (
	sel_numberWithInt    = objc.RegisterName("numberWithInt:")
	sel_numberWithDouble = objc.RegisterName("numberWithDouble:")
	sel_intValue         = objc.RegisterName("intValue")
	sel_doubleValue      = objc.RegisterName("doubleValue")
)

type NSNumber struct {
	NSObject
}

func NSNumber_numberWithInt(val int32) NSNumber {
	return NSNumber{
		NSObject{
			ID: objc.ID(class_NSNumber).Send(sel_numberWithInt, val),
		},
	}
}

func NSNumber_numberWithDouble(val float64) NSNumber {
	return NSNumber{
		NSObject{
			ID: objc.ID(class_NSNumber).Send(sel_numberWithDouble, val),
		},
	}
}

func (n NSNumber) IntValue() int32 {
	return objc.Send[int32](n.ID, sel_intValue)
}

func (n NSNumber) DoubleValue() float64 {
	return objc.Send[float64](n.ID, sel_doubleValue)
}
