//go:build darwin

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
	if o.ID == 0 {
		return
	}
	o.Send(macdriver.SEL_release)
}

func (o *NSObject) Autorelease() {
	if o.ID == 0 {
		return
	}
	o.Send(macdriver.SEL_autorelease)
}

type NSError struct {
	NSObject
}
