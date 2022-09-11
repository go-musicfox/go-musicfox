//go:build darwin
// +build darwin

package core

import (
	"fmt"
	"github.com/progrium/macdriver/objc"
)

const (
	NSKeyValueObservingOptionNew = 1 << iota
	NSKeyValueObservingOptionOld
	NSKeyValueObservingOptionInitial
	NSKeyValueObservingOptionPrior
)

var observers = map[string]map[objc.Ref]objc.Object{}

func AddObserver_forKeyPath_options_context(self objc.Object, keyPath NSStringRef, options uint32, context objc.Ref, handler func(keyPath NSString, ofObject objc.Object, change NSDictionary, context objc.Ref)) {

	RemoveObserver_forKeyPath_context(self, keyPath, context)
	clsName := fmt.Sprintf("%s_%s_Observer", self.Class().String(), keyPath.Init_asNSString().String())
	cls := objc.NewClass(clsName, "NSObject")
	cls.AddMethod("observeValueForKeyPath:ofObject:change:context:", func(_, keyPath, ofObject, change, context objc.Object) {
		handler(NSString_FromObject(keyPath), ofObject, NSDictionary_fromRef(change), context)
	})
	objc.RegisterClass(cls)

	observer := objc.Get(clsName).Alloc().Init()
	if observers[clsName] == nil {
		observers[clsName] = make(map[objc.Ref]objc.Object)
	}

	observers[clsName][context] = observer
	self.Send("addObserver:forKeyPath:options:context:", observer, keyPath, options, context)
}

func RemoveObserver_forKeyPath(self objc.Object, keyPath NSStringRef) {
	clsName := fmt.Sprintf("%s_%s_Observer", self.Class().String(), keyPath.Init_asNSString().String())

	if obs, ok := observers[clsName]; ok {
		for k, ob := range obs {
			self.Send("removeObserver:forKeyPath:", ob, keyPath)
			delete(observers[clsName], k)
		}
	}
}

func RemoveObserver_forKeyPath_context(self objc.Object, keyPath NSStringRef, context objc.Ref) {
	clsName := fmt.Sprintf("%s_%s_Observer", self.Class().String(), keyPath.Init_asNSString().String())

	if obs, ok := observers[clsName]; ok {
		if ob, ok := obs[context]; ok {
			self.Send("removeObserver:forKeyPath:", ob, keyPath)
			delete(observers[clsName], context)
		}
	}
}
