//go:build darwin

package core

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_NSMethodSignature = objc.GetClass("NSMethodSignature")
}

var (
	class_NSMethodSignature objc.Class
)

var (
	sel_instanceMethodSignatureForSelector = objc.RegisterName("instanceMethodSignatureForSelector:")
	sel_signatureWithObjCTypes             = objc.RegisterName("signatureWithObjCTypes:")
)

type NSMethodSignature struct {
	NSObject
}

func NSMethodSignature_instanceMethodSignatureForSelector(self objc.ID, cmd objc.SEL) NSMethodSignature {
	return NSMethodSignature{NSObject{ID: self.Send(sel_instanceMethodSignatureForSelector, cmd)}}
}

// NSMethodSignature_signatureWithObjCTypes takes a string that represents the type signature of a method.
// It follows the encoding specified in the Apple Docs.
//
// [Apple Docs]: https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/ObjCRuntimeGuide/Articles/ocrtTypeEncodings.html#//apple_ref/doc/uid/TP40008048-CH100
func NSMethodSignature_signatureWithObjCTypes(types string) NSMethodSignature {
	return NSMethodSignature{NSObject{ID: objc.ID(class_NSMethodSignature).Send(sel_signatureWithObjCTypes, types)}}
}
