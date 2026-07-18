// Copyright 2016 The go-vgo Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

//go:build windows && purego

// Package hook (pure-Go Windows backend).
//
// This is a *CGo-free* event source built directly on the Win32 low-level
// hook API (SetWindowsHookEx with WH_KEYBOARD_LL / WH_MOUSE_LL) via
// golang.org/x/sys/windows. It is selected at build time with the "purego"
// tag (matching the macOS pure-Go backend):
//
//	go build -tags purego .
//
// It replaces the default CGo/libuiohook backend (hook_cgo.go) so the package
// can be built with `CGO_ENABLED=0` and no C toolchain on Windows. The event
// stream is wire-compatible with the CGo backend: the same Event.Kind ids and
// the same Keycode (libuiohook VC_* "virtual code") / Rawcode (Win32 VK code)
// values are produced, so Register/Process hotkeys and RawcodeToKeychar behave
// identically.
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │  Threading note                                                          │
// │                                                                          │
// │  A WH_*_LL hook is dispatched through the message queue of the thread    │
// │  that installed it, so that thread MUST install the hook and pump a      │
// │  GetMessage loop. winLoop() therefore pins itself with LockOSThread and  │
// │  End() tears the loop down by posting WM_QUIT to it.                     │
// └───────────────────────────────────────────────────────────────────────────┘
package hook

import (
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows hook ids.
const (
	whKeyboardLL = 13
	whMouseLL    = 14
)

// Window messages (subset relevant to the LL hooks).
const (
	wmQuit = 0x0012

	wmKeyDown    = 0x0100
	wmKeyUp      = 0x0101
	wmSysKeyDown = 0x0104
	wmSysKeyUp   = 0x0105

	wmMouseMove   = 0x0200
	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmRButtonDown = 0x0204
	wmRButtonUp   = 0x0205
	wmMButtonDown = 0x0207
	wmMButtonUp   = 0x0208
	wmMouseWheel  = 0x020A
	wmXButtonDown = 0x020B
	wmXButtonUp   = 0x020C
	wmMouseHWheel = 0x020E
)

// KBDLLHOOKSTRUCT flag bits and misc constants.
const (
	llkhfExtended = 0x01
	wheelDelta    = 120
	xbutton1      = 0x0001
	xbutton2      = 0x0002
)

// Virtual-key codes used for modifier bookkeeping / char translation.
const (
	vkShift    = 0x10
	vkControl  = 0x11
	vkMenu     = 0x12
	vkCapital  = 0x14
	vkLShift   = 0xA0
	vkRShift   = 0xA1
	vkLControl = 0xA2
	vkRControl = 0xA3
	vkLMenu    = 0xA4
	vkRMenu    = 0xA5
	vkLWin     = 0x5B
	vkRWin     = 0x5C
	vkNumlock  = 0x90
	vkScroll   = 0x91
)

// libuiohook modifier mask bits (iohook.h). Tracked so Event.Mask matches the
// CGo backend.
const (
	maskShiftL = 1 << 0
	maskCtrlL  = 1 << 1
	maskMetaL  = 1 << 2
	maskAltL   = 1 << 3
	maskShiftR = 1 << 4
	maskCtrlR  = 1 << 5
	maskMetaR  = 1 << 6
	maskAltR   = 1 << 7

	maskButton1 = 1 << 8
	maskButton2 = 1 << 9
	maskButton3 = 1 << 10
	maskButton4 = 1 << 11
	maskButton5 = 1 << 12

	maskNumLock    = 1 << 13
	maskCapsLock   = 1 << 14
	maskScrollLock = 1 << 15

	maskShift = maskShiftL | maskShiftR
	maskCtrl  = maskCtrlL | maskCtrlR
	maskMeta  = maskMetaL | maskMetaR
	maskAlt   = maskAltL | maskAltR

	maskButtons = maskButton1 | maskButton2 | maskButton3 | maskButton4 | maskButton5
)

// Mouse wheel scroll directions (iohook.h).
const (
	wheelVerticalDir   = 3
	wheelHorizontalDir = 4
)

// POINT mirrors the Win32 POINT struct.
type point struct {
	x, y int32
}

// kbdLLHookStruct mirrors KBDLLHOOKSTRUCT.
type kbdLLHookStruct struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

// msLLHookStruct mirrors MSLLHOOKSTRUCT.
type msLLHookStruct struct {
	pt          point
	mouseData   uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

// msg mirrors the Win32 MSG struct.
type msg struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procSetWindowsHookEx   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHook  = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx     = user32.NewProc("CallNextHookEx")
	procGetMessage         = user32.NewProc("GetMessageW")
	procPostThreadMessage  = user32.NewProc("PostThreadMessageW")
	procToUnicodeEx        = user32.NewProc("ToUnicodeEx")
	procGetKeyboardLayout  = user32.NewProc("GetKeyboardLayout")
	procGetKeyState        = user32.NewProc("GetKeyState")
	procGetDoubleClickTime = user32.NewProc("GetDoubleClickTime")

	procGetModuleHandle  = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentThread = kernel32.NewProc("GetCurrentThreadId")
)

// winState holds the live hook session so End() can tear it down. Guarded by
// the package-level lck mutex.
type winState struct {
	keyboardHook uintptr
	mouseHook    uintptr
	threadID     uint32
}

var (
	win *winState

	// Persistent C-callable trampolines for the two hooks. Allocated once via
	// NewCallback (which never frees) so they are created lazily and reused.
	keyboardCallback uintptr
	mouseCallback    uintptr

	// Hook-thread-local state (only touched on the single locked OS thread that
	// runs winLoop, so no extra locking is required between callbacks).
	winModifiers uint16

	clickCount  uint16
	clickTime   uint32
	clickButton uint16
	lastClickX  int32
	lastClickY  int32
	lastMoveX   int32
	lastMoveY   int32
)

// Start installs the Win32 low-level keyboard/mouse hooks and returns the
// event channel. The optional timeout argument is accepted for API parity with
// the CGo backend but ignored: this backend is event-driven (it blocks in a
// GetMessage loop) rather than polled.
func Start(tm ...int) chan Event {
	_ = tm

	ev = make(chan Event, 1024)
	asyncon = true

	go winLoop()

	return ev
}

// End removes the hooks and resets all hook state. The optional timeout is the
// grace period (ms) to let the message loop drain before the channel closes.
func End(tm ...int) {
	tm1 := 10
	if len(tm) > 0 {
		tm1 = tm[0]
	}

	asyncon = false

	lck.Lock()
	tid := uint32(0)
	if win != nil {
		tid = win.threadID
	}
	lck.Unlock()

	// Posting WM_QUIT unblocks GetMessage on the hook thread, which then
	// unhooks and returns.
	if tid != 0 {
		procPostThreadMessage.Call(uintptr(tid), wmQuit, 0, 0)
	}

	time.Sleep(time.Millisecond * time.Duration(tm1))

	for len(ev) != 0 {
		<-ev
	}
	close(ev)

	resetState()
}

// addEvent: the single-shot *blocking* listener (AddEvent/StopEvent) is a
// CGo/libuiohook-only feature with no pure-Go equivalent here. The supported
// path is the channel API (Start + Register/Process). Returning a non-zero code
// makes the public AddEvent report failure rather than silently pretending to
// register a hook.
func addEvent(key string) int {
	_ = key
	return -1
}

// StopEvent is a no-op on the pure-Go Windows backend (see addEvent).
func StopEvent() {}

// winLoop installs the hooks on a pinned OS thread and pumps the message loop
// until End() posts WM_QUIT.
func winLoop() {
	// LL hooks are delivered on the installing thread's message queue, so this
	// goroutine must stay on one OS thread for the whole session.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if keyboardCallback == 0 {
		keyboardCallback = windows.NewCallback(keyboardProc)
		mouseCallback = windows.NewCallback(mouseProc)
	}

	hInst, _, _ := procGetModuleHandle.Call(0)
	tid, _, _ := procGetCurrentThread.Call()

	kbHook, _, _ := procSetWindowsHookEx.Call(whKeyboardLL, keyboardCallback, hInst, 0)
	msHook, _, _ := procSetWindowsHookEx.Call(whMouseLL, mouseCallback, hInst, 0)
	if kbHook == 0 || msHook == 0 {
		if kbHook != 0 {
			procUnhookWindowsHook.Call(kbHook)
		}
		if msHook != 0 {
			procUnhookWindowsHook.Call(msHook)
		}
		send(Event{Kind: HookDisabled})
		return
	}

	lck.Lock()
	win = &winState{keyboardHook: kbHook, mouseHook: msHook, threadID: uint32(tid)}
	lck.Unlock()

	// Reset the per-session modifier/click bookkeeping.
	winModifiers = 0
	clickCount, clickTime, clickButton = 0, 0, 0

	send(Event{Kind: HookEnabled})

	// Windows has no native "hook start" callback; the loop blocks here until
	// WM_QUIT (posted by End()) or an error.
	var m msg
	for asyncon {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 { // 0 == WM_QUIT, -1 == error
			break
		}
	}

	procUnhookWindowsHook.Call(kbHook)
	procUnhookWindowsHook.Call(msHook)

	lck.Lock()
	win = nil
	lck.Unlock()
}

// keyboardProc is the WH_KEYBOARD_LL callback. NewCallback delivers lParam as
// the typed struct pointer directly, which keeps the (vet-flagged)
// uintptr->unsafe.Pointer conversion out of our code.
func keyboardProc(nCode int, wParam uintptr, kb *kbdLLHookStruct) uintptr {
	if nCode >= 0 && kb != nil {
		switch wParam {
		case wmKeyDown, wmSysKeyDown:
			processKeyPressed(kb)
		case wmKeyUp, wmSysKeyUp:
			processKeyReleased(kb)
		}
	}

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, uintptr(unsafe.Pointer(kb)))
	return ret
}

func processKeyPressed(kb *kbdLLHookStruct) {
	setKeyModifier(kb.vkCode, true)

	vk := uint16(kb.vkCode)
	send(Event{
		Kind:    KeyDown,
		Mask:    winModifiers,
		Keycode: vkToKeycode(vk, kb.flags),
		Rawcode: vk,
		Keychar: CharUndefined,
	})

	// Emit a separate "typed" event (Kind == KeyHold, matching libuiohook's
	// EVENT_KEY_TYPED) carrying the translated unicode char, if any.
	if r, ok := keychar(kb); ok {
		lck.Lock()
		raw2keyWin[vk] = string([]rune{r})
		lck.Unlock()

		send(Event{
			Kind:    KeyHold,
			Mask:    winModifiers,
			Keycode: 0, // VC_UNDEFINED, as in the CGo backend
			Rawcode: vk,
			Keychar: r,
		})
	}
}

func processKeyReleased(kb *kbdLLHookStruct) {
	setKeyModifier(kb.vkCode, false)

	vk := uint16(kb.vkCode)
	send(Event{
		Kind:    KeyUp,
		Mask:    winModifiers,
		Keycode: vkToKeycode(vk, kb.flags),
		Rawcode: vk,
		Keychar: CharUndefined,
	})
}

// mouseProc is the WH_MOUSE_LL callback (see keyboardProc on the pointer arg).
func mouseProc(nCode int, wParam uintptr, ms *msLLHookStruct) uintptr {
	if nCode >= 0 && ms != nil {
		switch wParam {
		case wmLButtonDown:
			winModifiers |= maskButton1
			processButtonPressed(ms, MouseMap["left"])
		case wmRButtonDown:
			winModifiers |= maskButton2
			processButtonPressed(ms, MouseMap["right"])
		case wmMButtonDown:
			winModifiers |= maskButton3
			processButtonPressed(ms, MouseMap["center"])
		case wmXButtonDown:
			processButtonPressed(ms, xButton(ms, true))
		case wmLButtonUp:
			winModifiers &^= maskButton1
			processButtonReleased(ms, MouseMap["left"])
		case wmRButtonUp:
			winModifiers &^= maskButton2
			processButtonReleased(ms, MouseMap["right"])
		case wmMButtonUp:
			winModifiers &^= maskButton3
			processButtonReleased(ms, MouseMap["center"])
		case wmXButtonUp:
			processButtonReleased(ms, xButton(ms, false))
		case wmMouseMove:
			processMouseMoved(ms)
		case wmMouseWheel:
			processMouseWheel(ms, wheelVerticalDir)
		case wmMouseHWheel:
			processMouseWheel(ms, wheelHorizontalDir)
		}
	}

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, uintptr(unsafe.Pointer(ms)))
	return ret
}

// xButton resolves an X (extra) mouse button to its gohook button number and
// sets (press) or clears (release) the matching button modifier bit, so the
// mask does not stay stuck after XButtonUp.
func xButton(ms *msLLHookStruct, down bool) uint16 {
	var btn, bit uint16
	switch uint16(ms.mouseData >> 16) {
	case xbutton1:
		btn, bit = 4, maskButton4
	case xbutton2:
		btn, bit = 5, maskButton5
	default:
		return uint16(ms.mouseData >> 16)
	}

	if down {
		winModifiers |= bit
	} else {
		winModifiers &^= bit
	}
	return btn
}

func processButtonPressed(ms *msLLHookStruct, button uint16) {
	// Track consecutive clicks of the same button within the double-click time.
	if button == clickButton && ms.time-clickTime <= doubleClickTime() {
		if clickCount < 0xFFFF {
			clickCount++
		}
	} else {
		clickCount = 1
		clickButton = button
	}
	clickTime = ms.time
	lastClickX, lastClickY = ms.pt.x, ms.pt.y

	send(Event{
		Kind:   MouseDown, // EVENT_MOUSE_PRESSED
		Mask:   winModifiers,
		Button: button,
		Clicks: clickCount,
		X:      int16(ms.pt.x),
		Y:      int16(ms.pt.y),
	})
}

func processButtonReleased(ms *msLLHookStruct, button uint16) {
	send(Event{
		Kind:   MouseHold, // EVENT_MOUSE_RELEASED
		Mask:   winModifiers,
		Button: button,
		Clicks: clickCount,
		X:      int16(ms.pt.x),
		Y:      int16(ms.pt.y),
	})

	// A press+release at the same point is also a "click".
	if lastClickX == ms.pt.x && lastClickY == ms.pt.y {
		send(Event{
			Kind:   MouseUp, // EVENT_MOUSE_CLICKED
			Mask:   winModifiers,
			Button: button,
			Clicks: clickCount,
			X:      int16(ms.pt.x),
			Y:      int16(ms.pt.y),
		})
	}

	if button == clickButton && ms.time-clickTime > doubleClickTime() {
		clickCount = 0
	}
}

func processMouseMoved(ms *msLLHookStruct) {
	// Drop duplicate positions to avoid flooding the channel.
	if ms.pt.x == lastMoveX && ms.pt.y == lastMoveY {
		return
	}
	lastMoveX, lastMoveY = ms.pt.x, ms.pt.y

	kind := uint8(MouseMove)
	if winModifiers&maskButtons != 0 {
		kind = MouseDrag
	}

	send(Event{
		Kind:   kind,
		Mask:   winModifiers,
		Button: 0, // MOUSE_NOBUTTON
		X:      int16(ms.pt.x),
		Y:      int16(ms.pt.y),
	})
}

func processMouseWheel(ms *msLLHookStruct, direction uint8) {
	clickCount = 1
	clickButton = 0

	// HIWORD(mouseData) is a signed wheel delta: +120 forward (away from the
	// user), -120 backward. libuiohook reports rotation in clicks, inverted.
	delta := int16(uint16(ms.mouseData >> 16))
	rotation := int32(delta/wheelDelta) * -1

	send(Event{
		Kind:      MouseWheel,
		Mask:      winModifiers,
		Clicks:    clickCount,
		X:         int16(ms.pt.x),
		Y:         int16(ms.pt.y),
		Amount:    wheelAmount(),
		Rotation:  rotation,
		Direction: direction,
	})
}

// setKeyModifier maintains winModifiers for the modifier/lock keys, mirroring
// the CGo backend's set/unset_modifier_mask logic.
func setKeyModifier(vkCode uint32, down bool) {
	var bit uint16
	switch vkCode {
	case vkLShift:
		bit = maskShiftL
	case vkRShift:
		bit = maskShiftR
	case vkLControl:
		bit = maskCtrlL
	case vkRControl:
		bit = maskCtrlR
	case vkLMenu:
		bit = maskAltL
	case vkRMenu:
		bit = maskAltR
	case vkLWin:
		bit = maskMetaL
	case vkRWin:
		bit = maskMetaR
	case vkNumlock:
		bit = maskNumLock
	case vkCapital:
		bit = maskCapsLock
	case vkScroll:
		bit = maskScrollLock
	default:
		return
	}

	if down {
		winModifiers |= bit
	} else {
		winModifiers &^= bit
	}
}

// vkToKeycode converts a Win32 virtual-key code to the libuiohook VC_* code,
// mirroring keycode_to_scancode() in hook/windows/input_c.h. Unlike the C
// helper we deliberately do NOT OR in the 0xEE00 "extended" bits for the
// navigation cluster: the plain table value already equals the
// github.com/vcaesar/keycode Keycode value (e.g. VC_LEFT == 0xE04B), which is
// what Register/Process match against.
func vkToKeycode(vk uint16, flags uint32) uint16 {
	_ = flags
	return winVKToKeycode[vk]
}

// keychar translates a key press into its unicode character using ToUnicodeEx
// against the active keyboard layout. The modifier state is synthesized from
// winModifiers (GetKeyboardState is unreliable inside a low-level hook).
//
// NOTE: ToUnicodeEx mutates the kernel's dead-key buffer; in the rare case the
// user is composing a diacritic this can disturb that state. Keychar is best
// effort and informational.
func keychar(kb *kbdLLHookStruct) (rune, bool) {
	var state [256]byte
	if winModifiers&maskShift != 0 {
		state[vkShift] = 0x80
	}
	if winModifiers&maskCtrl != 0 {
		state[vkControl] = 0x80
	}
	if winModifiers&maskAlt != 0 {
		state[vkMenu] = 0x80
	}
	if caps, _, _ := procGetKeyState.Call(vkCapital); caps&0x0001 != 0 {
		state[vkCapital] = 0x01
	}

	hkl, _, _ := procGetKeyboardLayout.Call(0)

	var buf [8]uint16
	r, _, _ := procToUnicodeEx.Call(
		uintptr(kb.vkCode),
		uintptr(kb.scanCode),
		uintptr(unsafe.Pointer(&state[0])),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0,
		hkl,
	)

	n := int32(r)
	if n < 1 {
		// 0 == no translation, -1 == dead key.
		return 0, false
	}

	ch := rune(buf[0])
	// Filter control characters (e.g. esc, backspace, enter) which are not
	// meaningful "typed" characters.
	if ch < 0x20 || ch == 0x7F {
		return 0, false
	}
	return ch, true
}

// doubleClickTime returns the system double-click interval in milliseconds.
func doubleClickTime() uint32 {
	d, _, _ := procGetDoubleClickTime.Call()
	if d == 0 {
		return 500
	}
	return uint32(d)
}

// wheelAmount returns the configured wheel scroll amount (lines per notch).
func wheelAmount() uint16 {
	// The CGo backend queries SPI_GETWHEELSCROLLLINES; the common default is 3.
	return 3
}

// send timestamps and pushes an event onto the global channel without blocking
// the hook callback (a blocked LL hook stalls all system input and is silently
// torn down by Windows). The recover guards the small shutdown window where
// End() may have closed ev while a callback is still in flight.
func send(e Event) {
	if !asyncon {
		return
	}
	defer func() { _ = recover() }() // ev closed by End(): drop silently

	e.When = time.Now()
	select {
	case ev <- e:
	default:
		// channel full: drop rather than stalling the Windows input queue.
	}
}

// winVKToKeycode maps a Windows virtual-key code to the libuiohook
// VC_* "virtual code" (== github.com/vcaesar/keycode Keycode values),
// mirroring keycode_to_scancode() in hook/windows/input_c.h. Generated by
// test/gen; do not edit by hand. Zero (VC_UNDEFINED) entries are omitted.
var winVKToKeycode = map[uint16]uint16{
	0x01: 0x0001,
	0x02: 0x0002,
	0x04: 0x0003,
	0x05: 0x0004,
	0x06: 0x0005,
	0x08: 0x000E,
	0x09: 0x000F,
	0x0C: 0xE04C,
	0x0D: 0x001C,
	0x10: 0x002A,
	0x11: 0x001D,
	0x12: 0x0038,
	0x13: 0x0E45,
	0x14: 0x003A,
	0x15: 0x0070,
	0x19: 0x0079,
	0x1B: 0x0001,
	0x20: 0x0039,
	0x21: 0x0E49,
	0x22: 0x0E51,
	0x23: 0x0E4F,
	0x24: 0x0E47,
	0x25: 0xE04B,
	0x26: 0xE048,
	0x27: 0xE04D,
	0x28: 0xE050,
	0x2C: 0x0E37,
	0x2D: 0x0E52,
	0x2E: 0x0E53,
	0x30: 0x000B,
	0x31: 0x0002,
	0x32: 0x0003,
	0x33: 0x0004,
	0x34: 0x0005,
	0x35: 0x0006,
	0x36: 0x0007,
	0x37: 0x0008,
	0x38: 0x0009,
	0x39: 0x000A,
	0x41: 0x001E,
	0x42: 0x0030,
	0x43: 0x002E,
	0x44: 0x0020,
	0x45: 0x0012,
	0x46: 0x0021,
	0x47: 0x0022,
	0x48: 0x0023,
	0x49: 0x0017,
	0x4A: 0x0024,
	0x4B: 0x0025,
	0x4C: 0x0026,
	0x4D: 0x0032,
	0x4E: 0x0031,
	0x4F: 0x0018,
	0x50: 0x0019,
	0x51: 0x0010,
	0x52: 0x0013,
	0x53: 0x001F,
	0x54: 0x0014,
	0x55: 0x0016,
	0x56: 0x002F,
	0x57: 0x0011,
	0x58: 0x002D,
	0x59: 0x0015,
	0x5A: 0x002C,
	0x5B: 0x0E5B,
	0x5C: 0x0E5C,
	0x5D: 0x0E5D,
	0x5F: 0xE05F,
	0x60: 0x0052,
	0x61: 0x004F,
	0x62: 0x0050,
	0x63: 0x0051,
	0x64: 0x004B,
	0x65: 0x004C,
	0x66: 0x004D,
	0x67: 0x0047,
	0x68: 0x0048,
	0x69: 0x0049,
	0x6A: 0x0037,
	0x6B: 0x004E,
	0x6D: 0x004A,
	0x6E: 0x0053,
	0x6F: 0x0E35,
	0x70: 0x003B,
	0x71: 0x003C,
	0x72: 0x003D,
	0x73: 0x003E,
	0x74: 0x003F,
	0x75: 0x0040,
	0x76: 0x0041,
	0x77: 0x0042,
	0x78: 0x0043,
	0x79: 0x0044,
	0x7A: 0x0057,
	0x7B: 0x0058,
	0x7C: 0x005B,
	0x7D: 0x005C,
	0x7E: 0x005D,
	0x7F: 0x0063,
	0x80: 0x0064,
	0x81: 0x0065,
	0x82: 0x0066,
	0x83: 0x0067,
	0x84: 0x0068,
	0x85: 0x0069,
	0x86: 0x006A,
	0x87: 0x006B,
	0x90: 0x0045,
	0x91: 0x0046,
	0xA0: 0x002A,
	0xA1: 0x0036,
	0xA2: 0x001D,
	0xA3: 0x0E1D,
	0xA4: 0x0038,
	0xA5: 0x0E38,
	0xA6: 0xE06A,
	0xA7: 0xE069,
	0xA8: 0xE067,
	0xA9: 0xE068,
	0xAA: 0xE065,
	0xAB: 0xE066,
	0xAC: 0xE032,
	0xAD: 0xE020,
	0xAE: 0xE02E,
	0xAF: 0xE030,
	0xB0: 0xE019,
	0xB1: 0xE010,
	0xB2: 0xE024,
	0xB3: 0xE022,
	0xB5: 0xE06D,
	0xB6: 0xE06C,
	0xB7: 0xE021,
	0xBA: 0x0027,
	0xBB: 0x000D,
	0xBC: 0x0033,
	0xBD: 0x000C,
	0xBE: 0x0034,
	0xBF: 0x0035,
	0xC0: 0x0029,
	0xDB: 0x001A,
	0xDC: 0x002B,
	0xDD: 0x001B,
	0xDE: 0x0028,
	0xDF: 0x007D,
	0xE5: 0xE064,
	0xE6: 0xE03C,
	0xFE: 0xE04C,
}
