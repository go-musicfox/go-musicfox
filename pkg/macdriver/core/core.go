package core

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
)

var (
	class_NSAutoreleasePool = objc.GetClass("NSAutoreleasePool")
	class_NSString          = objc.GetClass("NSString")
)

var (
	sel_initWithUTF8String = objc.RegisterName("initWithUTF8String:")
	sel_UTF8String         = objc.RegisterName("UTF8String")
	sel_unsignedIntValue   = objc.RegisterName("unsignedIntValue")
)

type NSUInteger = uint
type NSInteger = int

type NSObject struct {
	objc.ID
}

func (o NSObject) Release() {
	o.Send(macdriver.SEL_release)
}

type NSError struct {
	NSObject
}

type NSAutoreleasePool struct {
	NSObject
}

func NSAutoreleasePool_new() NSAutoreleasePool {
	return NSAutoreleasePool{
		NSObject{
			ID: objc.ID(class_NSAutoreleasePool).Send(macdriver.SEL_new),
		},
	}
}
