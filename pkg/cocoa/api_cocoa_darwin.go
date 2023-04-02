package cocoa

import (
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

var (
	class_NSString          = objc.GetClass("NSString")
	class_NSAutoreleasePool = objc.GetClass("NSAutoreleasePool")
)

var (
	sel_alloc              = objc.RegisterName("alloc")
	sel_new                = objc.RegisterName("new")
	sel_init               = objc.RegisterName("init")
	sel_release            = objc.RegisterName("release")
	sel_initWithUTF8String = objc.RegisterName("initWithUTF8String:")
	sel_UTF8String         = objc.RegisterName("UTF8String")
	sel_length             = objc.RegisterName("length")
	sel_object             = objc.RegisterName("object")
	sel_objectForKey       = objc.RegisterName("objectForKey:")
	sel_unsignedIntValue   = objc.RegisterName("unsignedIntValue")
)

type CGFloat = float64

type CGSize struct {
	Width, Height CGFloat
}

type CGPoint struct {
	X, Y float64
}

type CGRect struct {
	Origin CGPoint
	Size   CGSize
}

type NSUInteger = uint
type NSInteger = int

type NSPoint = CGPoint
type NSRect = CGRect
type NSSize = CGSize

type NSError struct {
	objc.ID
}

type NSColor struct {
	objc.ID
}

type NSAutoreleasePool struct {
	objc.ID
}

func NSAutoreleasePool_new() NSAutoreleasePool {
	return NSAutoreleasePool{objc.ID(class_NSAutoreleasePool).Send(sel_new)}
}

func (p NSAutoreleasePool) Release() {
	p.Send(sel_release)
}

type NSString struct {
	objc.ID
}

func NSString_alloc() NSString {
	return NSString{objc.ID(class_NSString).Send(sel_alloc)}
}

func (s NSString) InitWithUTF8String(utf8 string) NSString {
	return NSString{s.Send(sel_initWithUTF8String, utf8)}
}

func (s NSString) String() string {
	return string(unsafe.Slice((*byte)(unsafe.Pointer(s.Send(sel_UTF8String))), s.Send(sel_length)))
}

type NSNotification struct {
	objc.ID
}

func (n NSNotification) Object() objc.ID {
	return n.Send(sel_object)
}

type NSDictionary struct {
	objc.ID
}

func (d NSDictionary) ObjectForKey(object objc.ID) objc.ID {
	return d.Send(sel_objectForKey, object)
}

type NSNumber struct {
	objc.ID
}

func (n NSNumber) UnsignedIntValue() uint {
	return uint(n.Send(sel_unsignedIntValue))
}
