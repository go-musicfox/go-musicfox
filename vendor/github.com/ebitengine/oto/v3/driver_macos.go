// Copyright 2021 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build darwin && !ios

package oto

import (
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const defaultOneBufferSizeInBytes = 2048

// setNotificationHandler sets a handler for sleep/wake notifications.
func setNotificationHandler() error {
	appkit, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/Versions/Current/AppKit", purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}

	// Create the Observer object
	var class objc.Class
	class, err = objc.RegisterClass("OtoNotificationObserver", objc.GetClass("NSObject"), nil, nil,
		[]objc.MethodDef{
			{
				Cmd: objc.RegisterName("receiveSleepNote:"),
				Fn:  setGlobalPause,
			},
			{
				Cmd: objc.RegisterName("receiveWakeNote:"),
				Fn:  setGlobalResume,
			},
		})

	observer := objc.ID(class).Send(objc.RegisterName("new"))

	notificationCenter := objc.ID(objc.GetClass("NSWorkspace")).Send(objc.RegisterName("sharedWorkspace")).Send(objc.RegisterName("notificationCenter"))

	// Dlsym returns a pointer to the object so dereference it
	s, err := purego.Dlsym(appkit, "NSWorkspaceWillSleepNotification")
	if err != nil {
		return err
	}

	notificationCenter.Send(objc.RegisterName("addObserver:selector:name:object:"),
		observer,
		objc.RegisterName("receiveSleepNote:"),
		*(*uintptr)(unsafe.Pointer(s)),
		0,
	)

	s, err = purego.Dlsym(appkit, "NSWorkspaceDidWakeNotification")
	if err != nil {
		return err
	}

	notificationCenter.Send(objc.RegisterName("addObserver:selector:name:object:"),
		observer,
		objc.RegisterName("receiveWakeNote:"),
		*(*uintptr)(unsafe.Pointer(s)),
		0,
	)
	return nil
}
