//go:build darwin

package core

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
)

func init() {
	importFramework()
	class_NSAutoreleasePool = objc.GetClass("NSAutoreleasePool")
}

var (
	class_NSAutoreleasePool objc.Class
)

var (
	sel_drain = objc.RegisterName("drain")
)

type NSAutoreleasePool struct {
	NSObject
}

func NSAutoreleasePool_new() NSAutoreleasePool {
	return NSAutoreleasePool{NSObject{ID: objc.ID(class_NSAutoreleasePool).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init)}}
}

func (p NSAutoreleasePool) Drain() {
	if p.ID == 0 {
		return
	}
	p.Send(sel_drain)
}
