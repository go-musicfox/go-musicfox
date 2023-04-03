package core

import (
	"unsafe"

	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
)

type NSString struct {
	NSObject
}

func NSString_alloc() NSString {
	return NSString{
		NSObject{
			ID: objc.ID(class_NSString).Send(macdriver.SEL_alloc),
		},
	}
}

func (s NSString) InitWithUTF8String(utf8 string) NSString {
	return NSString{
		NSObject{
			ID: s.Send(sel_initWithUTF8String, utf8),
		},
	}
}

func (s NSString) String() string {
	return string(unsafe.Slice((*byte)(unsafe.Pointer(s.Send(sel_UTF8String))), s.Send(macdriver.SEL_length)))
}
