package core

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
)

var importOnce sync.Once

func importFramework() {
	importOnce.Do(func() {
		_, err := purego.Dlopen("Foundation.framework/Foundation", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
	})
}

type NSUInteger = uint32
type NSInteger = int32

type NSObject struct {
	objc.ID
}

func (o *NSObject) SetObjcID(id objc.ID) {
	o.ID = id
}

func (o *NSObject) Release() {
	o.Send(macdriver.SEL_release)
}

type NSError struct {
	NSObject
}

type NSAutoreleasePool struct {
	NSObject
}
