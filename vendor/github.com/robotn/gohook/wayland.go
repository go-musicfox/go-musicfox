// Copyright 2016 The go-vgo Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

//go:build linux && wayland
// +build linux,wayland

// Package hook (Wayland backend).
//
// This is a *pure-Go* event source built on github.com/vcaesar/go-wayland
// (a CGo-free Wayland client). It is selected at build time with the
// "wayland" tag:
//
//	go build -tags wayland .
//
// It replaces the default CGo/libuiohook backend (hook_cgo.go) and needs no
// C toolchain and no X11/Xtst development libraries.
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │  IMPORTANT — Wayland security model                                       │
// │                                                                           │
// │  Wayland deliberately has NO global keylogging/mouse-hooking primitive.   │
// │  A wl_seat only delivers wl_keyboard / wl_pointer events to a surface     │
// │  while THAT surface holds input focus. There is no XRecord equivalent.    │
// │                                                                           │
// │  Consequences for this backend:                                          │
// │    • It captures input only while a surface owned by this process has     │
// │      keyboard/pointer focus (e.g. a robotgo/GUI window). Without a        │
// │      focused surface no key/pointer events are produced — by design.      │
// │    • TRUE global capture on Wayland requires the xdg-desktop-portal       │
// │      InputCapture / RemoteDesktop portals plus the libei (EI) protocol.   │
// │      See SUGGEST_WAYLAND.md for the recommended path and a survey of      │
// │      compositor support (GNOME/KDE = yes, wlroots/Hyprland = not yet).    │
// └───────────────────────────────────────────────────────────────────────────┘
package hook

import (
	"time"
	"unicode/utf8"

	"github.com/vcaesar/go-wayland/client"
)

// Linux evdev button codes (from linux/input-event-codes.h). The Wayland
// wl_pointer.button event carries these raw codes.
const (
	btnLeft   = 0x110
	btnRight  = 0x111
	btnMiddle = 0x112
	btnSide   = 0x113
	btnExtra  = 0x114
)

// wl_pointer.axis values.
const (
	axisVerticalScroll   = 0
	axisHorizontalScroll = 1
)

// libuiohook-compatible scroll direction constants (matches the C backend).
const (
	wheelVertical   = 3
	wheelHorizontal = 4
)

// waylandState holds the live connection objects for the running session so
// End() can tear them down. Guarded by the package-level lck mutex.
type waylandState struct {
	display  *client.Display
	seat     *client.Seat
	keyboard *client.Keyboard
	pointer  *client.Pointer

	// last known pointer position within the focused surface.
	x, y int16
}

var wl *waylandState

// Start adds the Wayland input listener and returns the event channel.
//
// The optional timeout argument is accepted for API parity with the CGo
// backend but is ignored: the Wayland backend is event-driven (it blocks on
// the compositor socket) rather than polled.
func Start(tm ...int) chan Event {
	_ = tm

	ev = make(chan Event, 1024)
	asyncon = true

	go waylandLoop()

	return ev
}

// End removes the Wayland input listener and resets all hook state. The
// optional timeout is the grace period (ms) to let the dispatch loop drain
// before the channel closes (API parity with the other backends).
func End(tm ...int) {
	tm1 := 10
	if len(tm) > 0 {
		tm1 = tm[0]
	}

	asyncon = false

	lck.Lock()
	st := wl
	wl = nil
	lck.Unlock()

	// Closing the context unblocks the dispatch loop's socket read.
	if st != nil && st.display != nil {
		if ctx := st.display.Context(); ctx != nil {
			if err := ctx.Close(); err != nil {
				// Best-effort teardown; the connection may already be gone.
				_ = err
			}
		}
	}

	time.Sleep(time.Millisecond * time.Duration(tm1))

	for len(ev) != 0 {
		<-ev
	}
	close(ev)

	resetState()
}

// addEvent: the single-shot *blocking* listener (AddEvent/StopEvent) is a
// CGo/libuiohook-only feature with no Wayland equivalent. The supported path
// on this backend is the channel API (Start + Register/Process). Returning a
// non-zero code makes the public AddEvent report failure rather than silently
// pretending to register a hook.
func addEvent(key string) int {
	_ = key
	return -1
}

// StopEvent is a no-op on the Wayland backend (see addEvent).
func StopEvent() {}

// waylandLoop connects to the compositor, wires up seat input handlers and
// pumps the dispatch loop until End() closes the connection.
func waylandLoop() {
	display, err := client.Connect("")
	if err != nil {
		// No compositor / not a Wayland session: report disabled and bail.
		send(Event{Kind: HookDisabled})
		return
	}

	registry, err := display.GetRegistry()
	if err != nil {
		send(Event{Kind: HookDisabled})
		_ = display.Context().Close()
		return
	}

	st := &waylandState{display: display}
	lck.Lock()
	wl = st
	lck.Unlock()

	registry.SetGlobalHandler(func(e client.RegistryGlobalEvent) {
		if e.Interface != client.SeatInterfaceName {
			return
		}

		seat := client.NewSeat(display.Context())
		if err := registry.Bind(e.Name, e.Interface, e.Version, seat); err != nil {
			return
		}

		lck.Lock()
		st.seat = seat
		lck.Unlock()

		seat.SetCapabilitiesHandler(func(ce client.SeatCapabilitiesEvent) {
			bindSeatCapabilities(st, seat, ce.Capabilities)
		})
	})

	// First roundtrip surfaces the globals (and binds the seat); the second
	// delivers the seat capabilities so keyboard/pointer get created.
	if err := display.Roundtrip(); err != nil {
		send(Event{Kind: HookDisabled})
		return
	}
	if err := display.Roundtrip(); err != nil {
		send(Event{Kind: HookDisabled})
		return
	}

	send(Event{Kind: HookEnabled})

	for asyncon {
		if err := display.Context().Dispatch(); err != nil {
			// Closed by End() or the compositor went away.
			break
		}
	}
}

// bindSeatCapabilities lazily creates the keyboard/pointer objects advertised
// by the seat and attaches the gohook event translators.
func bindSeatCapabilities(st *waylandState, seat *client.Seat, caps uint32) {
	if caps&uint32(client.SeatCapabilityKeyboard) != 0 {
		lck.Lock()
		need := st.keyboard == nil
		lck.Unlock()
		if need {
			if kb, err := seat.GetKeyboard(); err == nil {
				lck.Lock()
				st.keyboard = kb
				lck.Unlock()
				attachKeyboard(kb)
			}
		}
	}

	if caps&uint32(client.SeatCapabilityPointer) != 0 {
		lck.Lock()
		need := st.pointer == nil
		lck.Unlock()
		if need {
			if p, err := seat.GetPointer(); err == nil {
				lck.Lock()
				st.pointer = p
				lck.Unlock()
				attachPointer(st, p)
			}
		}
	}
}

// attachKeyboard wires wl_keyboard.key into KeyDown / KeyHold / KeyUp events.
func attachKeyboard(kb *client.Keyboard) {
	kb.SetKeyHandler(func(e client.KeyboardKeyEvent) {
		var kind uint8
		switch e.State {
		case uint32(client.KeyboardKeyStateReleased):
			kind = KeyUp
		case uint32(client.KeyboardKeyStateRepeated):
			kind = KeyHold
		default: // KeyboardKeyStatePressed
			kind = KeyDown
		}
		send(keyEvent(kind, e.Key))
	})
}

// attachPointer wires wl_pointer motion / button / axis into mouse events.
func attachPointer(st *waylandState, p *client.Pointer) {
	p.SetMotionHandler(func(e client.PointerMotionEvent) {
		lck.Lock()
		st.x = int16(e.SurfaceX)
		st.y = int16(e.SurfaceY)
		x, y := st.x, st.y
		lck.Unlock()

		send(Event{Kind: MouseMove, X: x, Y: y})
	})

	p.SetButtonHandler(func(e client.PointerButtonEvent) {
		lck.Lock()
		x, y := st.x, st.y
		lck.Unlock()

		kind := MouseUp
		if e.State == uint32(client.PointerButtonStatePressed) {
			kind = MouseDown
		}

		send(Event{
			Kind:   uint8(kind),
			Button: mouseButton(e.Button),
			Clicks: 1,
			X:      x,
			Y:      y,
		})
	})

	p.SetAxisHandler(func(e client.PointerAxisEvent) {
		lck.Lock()
		x, y := st.x, st.y
		lck.Unlock()

		dir := uint8(wheelVertical)
		if e.Axis == axisHorizontalScroll {
			dir = wheelHorizontal
		}

		rot := int32(e.Value)
		amt := rot
		if amt < 0 {
			amt = -amt
		}

		send(Event{
			Kind:      MouseWheel,
			X:         x,
			Y:         y,
			Amount:    uint16(amt),
			Rotation:  rot, // >0 down/right, <0 up/left (Wayland convention)
			Direction: dir,
		})
	})
}

// keyEvent builds a gohook key Event from a raw Linux evdev keycode.
//
// vcaesar/keycode's Keycode map already uses Linux evdev codes for the main
// keys (esc=1, a=30, space=57 ...), so for those Rawcode == Keycode. For keys
// where the name map yields a vcaesar entry we use it (this is what Register()
// matches against); otherwise we fall back to the raw evdev code.
func keyEvent(kind uint8, evcode uint32) Event {
	e := Event{
		Kind:    kind,
		Rawcode: uint16(evcode),
		Keychar: CharUndefined,
	}

	name, ok := waylandKeyName[evcode]
	if !ok {
		e.Keycode = uint16(evcode)
		return e
	}

	if kc, ok := Keycode[name]; ok {
		e.Keycode = kc
	} else {
		e.Keycode = uint16(evcode)
	}

	// A single-rune name is a printable character (e.g. "a", "1", ";").
	if utf8.RuneCountInString(name) == 1 {
		r, _ := utf8.DecodeRuneInString(name)
		e.Keychar = r

		// Keep RawcodetoKeychar() in sync, mirroring the CGo backend.
		lck.Lock()
		raw2keyLinux[e.Rawcode] = name
		lck.Unlock()
	}

	return e
}

// mouseButton maps a Linux evdev button code to a gohook MouseMap value.
func mouseButton(code uint32) uint16 {
	switch code {
	case btnLeft:
		return MouseMap["left"]
	case btnRight:
		return MouseMap["right"]
	case btnMiddle:
		return MouseMap["center"]
	case btnSide:
		return 4
	case btnExtra:
		return 5
	default:
		return uint16(code - btnLeft + 1)
	}
}

// send timestamps and pushes an event onto the global channel. It drops the
// event (rather than blocking the compositor dispatch loop) if no consumer is
// keeping up and the buffer is full. The recover guards the small shutdown
// window where End() may have closed ev while a handler is still in flight.
func send(e Event) {
	if !asyncon {
		return
	}
	defer func() { _ = recover() }() // ev closed by End(): drop silently

	e.When = time.Now()
	select {
	case ev <- e:
	default:
		// channel full: drop to avoid stalling the Wayland event loop.
	}
}

// waylandKeyName maps Linux evdev keycodes (linux/input-event-codes.h) to the
// gohook/vcaesar key-name strings. Keep this in sync with vcaesar/keycode so
// Keycode[name] resolves for hotkey matching via Register().
var waylandKeyName = map[uint32]string{
	1:  "esc",
	2:  "1",
	3:  "2",
	4:  "3",
	5:  "4",
	6:  "5",
	7:  "6",
	8:  "7",
	9:  "8",
	10: "9",
	11: "0",
	12: "-",
	13: "=",
	14: "delete", // KEY_BACKSPACE (vcaesar maps "delete"=14)
	15: "tab",
	16: "q",
	17: "w",
	18: "e",
	19: "r",
	20: "t",
	21: "y",
	22: "u",
	23: "i",
	24: "o",
	25: "p",
	26: "[",
	27: "]",
	28: "enter",
	29: "ctrl", // KEY_LEFTCTRL
	30: "a",
	31: "s",
	32: "d",
	33: "f",
	34: "g",
	35: "h",
	36: "j",
	37: "k",
	38: "l",
	39: ";",
	40: "'",
	41: "`",
	42: "shift", // KEY_LEFTSHIFT
	43: "\\",
	44: "z",
	45: "x",
	46: "c",
	47: "v",
	48: "b",
	49: "n",
	50: "m",
	51: ",",
	52: ".",
	53: "/",
	54: "shiftr", // KEY_RIGHTSHIFT
	55: "num_asterisk",
	56: "alt", // KEY_LEFTALT
	57: "space",
	58: "capslock",
	59: "f1",
	60: "f2",
	61: "f3",
	62: "f4",
	63: "f5",
	64: "f6",
	65: "f7",
	66: "f8",
	67: "f9",
	68: "f10",
	69: "numlock",
	70: "scrolllock",
	71: "num7",
	72: "num8",
	73: "num9",
	74: "num_minus",
	75: "num4",
	76: "num5",
	77: "num6",
	78: "num_plus",
	79: "num1",
	80: "num2",
	81: "num3",
	82: "num0",
	83: "num_period",
	87: "f11",
	88: "f12",
	96: "num_enter",
	97: "ctrl", // KEY_RIGHTCTRL
	98: "num_slash",

	100: "altr", // KEY_RIGHTALT
	102: "home",
	103: "up",
	104: "pageup",
	105: "left",
	106: "right",
	107: "end",
	108: "down",
	109: "pagedown",
	110: "insert",
	111: "delete", // KEY_DELETE
	119: "pause",

	125: "cmd",  // KEY_LEFTMETA  (super/win)
	126: "cmdr", // KEY_RIGHTMETA
}
