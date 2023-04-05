package core

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
)

func init() {
	importFramework()
	class_NSMutableDictionary = objc.GetClass("NSMutableDictionary")
}

var (
	class_NSMutableDictionary objc.Class
)

var (
	sel_setValueForKey = objc.RegisterName("setValue:forKey:")
)

type NSMutableDictionary struct {
	NSDictionary
}

func NSMutableDictionary_init() NSMutableDictionary {
	return NSMutableDictionary{
		NSDictionary{NSObject: NSObject{ID: objc.ID(class_NSMutableDictionary).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init)}},
	}
}

func NSMutableDictionary_initWithCapacity(cap NSUInteger) NSMutableDictionary {
	return NSMutableDictionary{
		NSDictionary{NSObject: NSObject{ID: objc.ID(class_NSMutableDictionary).Send(macdriver.SEL_alloc).Send(macdriver.SEL_initWithCapacity, cap)}},
	}
}

func (d NSMutableDictionary) SetValueForKey(key NSString, value NSObject) {
	d.Send(sel_setValueForKey, value.ID, key.ID)
}
