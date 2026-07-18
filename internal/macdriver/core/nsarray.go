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
	sel_count           = objc.RegisterName("count")
	sel_objectAtIndex   = objc.RegisterName("objectAtIndex:")
)

type NSArray struct {
	NSObject
}

func NSArray_arrayWithObject(obj NSObject) NSArray {
	return NSArray{NSObject{ID: objc.ID(class_NSArray).Send(sel_arrayWithObject, obj.ID)}}
}

func (a NSArray) Count() uint {
	if a.ID == 0 {
		return 0
	}
	return objc.Send[uint](a.ID, sel_count)
}

func (a NSArray) ObjectAtIndex(index uint) NSObject {
	return NSObject{ID: a.Send(sel_objectAtIndex, index)}
}
