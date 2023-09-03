//go:build darwin

package core

import (
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_NSInvocation = objc.GetClass("NSInvocation")
}

var (
	class_NSInvocation objc.Class
)

var (
	sel_setSelector                   = objc.RegisterName("setSelector:")
	sel_setTarget                     = objc.RegisterName("setTarget:")
	sel_invocationWithMethodSignature = objc.RegisterName("invocationWithMethodSignature:")
	sel_setArgumentAtIndex            = objc.RegisterName("setArgument:atIndex:")
	sel_getReturnValue                = objc.RegisterName("getReturnValue:")
	sel_invoke                        = objc.RegisterName("invoke")
	sel_invokeWithTarget              = objc.RegisterName("invokeWithTarget:")
)

type NSInvocation struct {
	NSObject
}

func NSInvocation_invocationWithMethodSignature(sig NSMethodSignature) NSInvocation {
	return NSInvocation{NSObject{ID: objc.ID(class_NSInvocation).Send(sel_invocationWithMethodSignature, sig.ID)}}
}

func (i NSInvocation) SetSelector(cmd objc.SEL) {
	i.Send(sel_setSelector, cmd)
}

func (i NSInvocation) SetTarget(target objc.ID) {
	i.Send(sel_setTarget, target)
}

func (i NSInvocation) SetArgumentAtIndex(arg unsafe.Pointer, idx int) {
	i.Send(sel_setArgumentAtIndex, arg, idx)
}

func (i NSInvocation) GetReturnValue(ret unsafe.Pointer) {
	i.Send(sel_getReturnValue, ret)
}

func (i NSInvocation) Invoke() {
	i.Send(sel_invoke)
}

func (i NSInvocation) InvokeWithTarget(target objc.ID) {
	i.Send(sel_invokeWithTarget, target)
}
