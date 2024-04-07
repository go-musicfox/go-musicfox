//go:build darwin

package core

import (
	"unsafe"

	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
)

func init() {
	importFramework()
	class_NSString = objc.GetClass("NSString")
}

var (
	class_NSString objc.Class
)

var (
	sel_initWithUTF8String = objc.RegisterName("initWithUTF8String:")
	sel_UTF8String         = objc.RegisterName("UTF8String")
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
	s.ID = s.Send(sel_initWithUTF8String, utf8)
	return s
}

func (s NSString) String() string {
	return string(unsafe.Slice((*byte)(unsafe.Pointer(s.Send(sel_UTF8String))), s.Send(macdriver.SEL_length))) //nolint:govet
}

func String(str string) NSString {
	return NSString_alloc().InitWithUTF8String(str)
}
