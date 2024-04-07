//go:build darwin

package core

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
)

var (
	importOnce sync.Once
	objcLib    uintptr

	objc_autoreleasePoolPush func() unsafe.Pointer
	objc_autoreleasePoolPop  func(ptr unsafe.Pointer)
)

func importFramework() {
	importOnce.Do(func() {
		var err error
		if _, err = purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL); err != nil {
			panic(err)
		}
		if objcLib, err = purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_GLOBAL); err != nil {
			panic(err)
		}
	})
}

func init() {
	importFramework()
	purego.RegisterLibFunc(&objc_autoreleasePoolPush, objcLib, "objc_autoreleasePoolPush")
	purego.RegisterLibFunc(&objc_autoreleasePoolPop, objcLib, "objc_autoreleasePoolPop")
}

type (
	NSUInteger = uint32
	NSInteger  = int32
)

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

func Autorelease(body func()) {
	// ObjC autoreleasepools are thread-local, so we need the body to be locked to
	// the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := objc_autoreleasePoolPush()
	defer objc_autoreleasePoolPop(pool)
	body()
}
