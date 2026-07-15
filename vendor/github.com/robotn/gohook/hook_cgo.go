// Copyright 2016 The go-vgo Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

//go:build !wayland && !(darwin && purego) && !(windows && purego) && !(linux && purego)

// Package hook (cgo backend). This is the default backend, a thin wrapper
// around the native libuiohook C engine (X11 on Linux, Cocoa on macOS,
// Win32 on Windows). Build with the "purego" tag to use the CGo-free
// backends instead: Quartz event tap on macOS (darwin.go), Win32 low-level
// hooks on Windows (windows.go), X RECORD on Linux (x11.go). Build with
// the "wayland" tag on Linux for the focused-surface Wayland backend
// (wayland.go).

package hook

/*
#cgo darwin CFLAGS: -x objective-c -Wno-deprecated-declarations
#cgo darwin LDFLAGS: -framework Cocoa

#cgo linux CFLAGS:-I/usr/src -std=gnu99
#cgo linux LDFLAGS: -L/usr/src -lX11 -lXtst
#cgo linux LDFLAGS: -lX11-xcb -lxcb -lxcb-xkb -lxkbcommon -lxkbcommon-x11
//#cgo windows LDFLAGS: -lgdi32 -luser32

#include "event/goEvent.h"
*/
import "C"

import (
	"time"
	"unsafe"
)

// Start adds global event hook to OS
// returns event channel
func Start(tm ...int) chan Event {
	ev = make(chan Event, 1024)
	go C.start_ev()

	tm1 := 50
	if len(tm) > 0 {
		tm1 = tm[0]
	}

	asyncon = true
	go func() {
		for {
			if !asyncon {
				return
			}

			C.pollEv()
			time.Sleep(time.Millisecond * time.Duration(tm1))
			//todo: find smallest time that does not destroy the cpu utilization
		}
	}()

	return ev
}

// End removes global event hook
func End(tm ...int) {
	tm1 := 10
	if len(tm) > 0 {
		tm1 = tm[0]
	}

	asyncon = false
	C.endPoll()
	C.stop_event()
	time.Sleep(time.Millisecond * time.Duration(tm1))

	for len(ev) != 0 {
		<-ev
	}
	close(ev)

	resetState()
}

// addEvent add the block event listener
func addEvent(key string) int {
	cs := C.CString(key)
	defer C.free(unsafe.Pointer(cs))

	eve := C.add_event(cs)
	geve := int(eve)

	return geve
}

// StopEvent stop the block event listener
func StopEvent() {
	C.stop_event()
}
