//go:build darwin

package core

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_NSArray = objc.GetClass("NSArray")
}

var (
	class_NSArray objc.Class
)

var (
	sel_arrayWithObject = objc.RegisterName("arrayWithObject:")
)

type NSArray struct {
	NSObject
}

func NSArray_arrayWithObject(obj NSObject) NSArray {
	return NSArray{NSObject{ID: objc.ID(class_NSArray).Send(sel_arrayWithObject, obj.ID)}}
}
