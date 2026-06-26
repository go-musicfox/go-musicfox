// Copyright 2016 The go-vgo Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

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
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const (
	// Version get the gohook version
	Version = "v0.40.0.123, Sierra Nevada!"

	// HookEnabled honk enable status
	HookEnabled  = 1 // iota
	HookDisabled = 2

	KeyDown = 4 // 3
	KeyHold = 3 // 4
	KeyUp   = 5 // 5

	MouseDown = 7 // 6
	MouseHold = 8 // 7
	MouseUp   = 6 // 8

	MouseMove  = 9
	MouseDrag  = 10
	MouseWheel = 11

	FakeEvent = 12

	// Keychar could be v
	CharUndefined = 0xFFFF
	WheelUp       = -1
	WheelDown     = 1
)

// Event Holds a system event
//
// If it's a Keyboard event the relevant fields are:
// Mask, Keycode, Rawcode, and Keychar,
// Keychar is probably what you want.
//
// If it's a Mouse event the relevant fields are:
// Button, Clicks, X, Y, Amount, Rotation and Direction
type Event struct {
	Kind     uint8 `json:"id"`
	When     time.Time
	Mask     uint16 `json:"mask"`
	Reserved uint16 `json:"reserved"`

	Keycode uint16 `json:"keycode"`
	Rawcode uint16 `json:"rawcode"`
	Keychar rune   `json:"keychar"`

	Button uint16 `json:"button"`
	Clicks uint16 `json:"clicks"`

	X int16 `json:"x"`
	Y int16 `json:"y"`

	Amount    uint16 `json:"amount"`
	Rotation  int32  `json:"rotation"`
	Direction uint8  `json:"direction"`
}

var (
	ev      = make(chan Event, 1024)
	asyncon = false

	lck sync.RWMutex

	pressed   = make(map[uint16]bool, 256)
	uppressed = make(map[uint16]bool, 256)
	used      = []int{}

	keys   = map[int][]uint16{}
	upkeys = map[int][]uint16{}
	cbs    = map[int]func(Event){}
	events = map[uint8][]int{}
)

func allPressed(pressed map[uint16]bool, keys ...uint16) bool {
	for _, i := range keys {
		// fmt.Println(i)
		if !pressed[i] {
			return false
		}
	}

	return true
}

// Register register gohook event
func Register(when uint8, cmds []string, cb func(Event)) {
	key := len(used)
	used = append(used, key)
	tmp := []uint16{}
	uptmp := []uint16{}

	for _, v := range cmds {
		if when == KeyUp {
			uptmp = append(uptmp, Keycode[v])
		}
		tmp = append(tmp, Keycode[v])
	}

	keys[key] = tmp
	upkeys[key] = uptmp
	cbs[key] = cb
	events[when] = append(events[when], key)
	// return
}

// Process return go hook process
func Process(evChan <-chan Event) (out chan bool) {
	out = make(chan bool)
	go func() {
		for ev := range evChan {
			if ev.Kind == KeyDown || ev.Kind == KeyHold {
				pressed[ev.Keycode] = true
				uppressed[ev.Keycode] = true
			} else if ev.Kind == KeyUp {
				pressed[ev.Keycode] = false
			}

			for _, v := range events[ev.Kind] {
				if !asyncon {
					break
				}

				if allPressed(pressed, keys[v]...) {
					cbs[v](ev)
				} else if ev.Kind == KeyUp {
					//uppressed[ev.Keycode] = true
					if allPressed(uppressed, upkeys[v]...) {
						uppressed = make(map[uint16]bool, 256)
						cbs[v](ev)
					}
				}
			}
		}

		// fmt.Println("exiting after end (process)")
		out <- true
	}()

	return
}

// String return formatted hook kind string
func (e Event) String() string {
	switch e.Kind {
	case HookEnabled:
		return fmt.Sprintf("%v - Event: {Kind: HookEnabled}", e.When)
	case HookDisabled:
		return fmt.Sprintf("%v - Event: {Kind: HookDisabled}", e.When)
	case KeyDown:
		return fmt.Sprintf("%v - Event: {Kind: KeyDown, Rawcode: %v, Keychar: %v}",
			e.When, e.Rawcode, e.Keychar)
	case KeyHold:
		return fmt.Sprintf("%v - Event: {Kind: KeyHold, Rawcode: %v, Keychar: %v}",
			e.When, e.Rawcode, e.Keychar)
	case KeyUp:
		return fmt.Sprintf("%v - Event: {Kind: KeyUp, Rawcode: %v, Keychar: %v}",
			e.When, e.Rawcode, e.Keychar)
	case MouseDown:
		return fmt.Sprintf("%v - Event: {Kind: MouseDown, Button: %v, X: %v, Y: %v, Clicks: %v}",
			e.When, e.Button, e.X, e.Y, e.Clicks)
	case MouseHold:
		return fmt.Sprintf("%v - Event: {Kind: MouseHold, Button: %v, X: %v, Y: %v, Clicks: %v}",
			e.When, e.Button, e.X, e.Y, e.Clicks)
	case MouseUp:
		return fmt.Sprintf("%v - Event: {Kind: MouseUp, Button: %v, X: %v, Y: %v, Clicks: %v}",
			e.When, e.Button, e.X, e.Y, e.Clicks)
	case MouseMove:
		return fmt.Sprintf("%v - Event: {Kind: MouseMove, Button: %v, X: %v, Y: %v, Clicks: %v}",
			e.When, e.Button, e.X, e.Y, e.Clicks)
	case MouseDrag:
		return fmt.Sprintf("%v - Event: {Kind: MouseDrag, Button: %v, X: %v, Y: %v, Clicks: %v}",
			e.When, e.Button, e.X, e.Y, e.Clicks)
	case MouseWheel:
		return fmt.Sprintf("%v - Event: {Kind: MouseWheel, Amount: %v, Rotation: %v, Direction: %v}",
			e.When, e.Amount, e.Rotation, e.Direction)
	case FakeEvent:
		return fmt.Sprintf("%v - Event: {Kind: FakeEvent}", e.When)
	}

	return "Unknown event, contact the mantainers."
}

// RawcodetoKeychar rawcode to keychar
func RawcodetoKeychar(r uint16) string {
	lck.RLock()
	defer lck.RUnlock()

	return raw2key[r]
}

// KeychartoRawcode key char to rawcode
func KeychartoRawcode(kc string) uint16 {
	return keytoraw[kc]
}

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
func End() {
	asyncon = false
	C.endPoll()
	C.stop_event()
	time.Sleep(time.Millisecond * 10)

	for len(ev) != 0 {
		<-ev
	}
	close(ev)

	pressed = make(map[uint16]bool, 256)
	uppressed = make(map[uint16]bool, 256)
	used = []int{}

	keys = map[int][]uint16{}
	cbs = map[int]func(Event){}
	events = map[uint8][]int{}
}

// AddEvent add the block event listener
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
