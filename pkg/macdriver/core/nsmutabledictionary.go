package core

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
)

var (
	class_NSMutableDictionary = objc.GetClass("NSMutableDictionary")
)

var (
	sel_setValueForKey = objc.RegisterName("setValue:forKey:")
)

type NSMutableDictionary struct {
	objc.ID
}

func NSMutableDictionary_init() NSMutableDictionary {
	return NSMutableDictionary{ID: objc.ID(class_NSMutableDictionary).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init)}
}

func NSMutableDictionary_initWithCapacity() NSMutableDictionary {
	return NSMutableDictionary{ID: objc.ID(class_NSMutableDictionary).Send(macdriver.SEL_alloc).Send(macdriver.SEL_initWithCapacity)}
}

func (d NSMutableDictionary) SetValueForKey(key NSString, value objc.ID) {
	d.Send(sel_setValueForKey, value, key.ID)
}
