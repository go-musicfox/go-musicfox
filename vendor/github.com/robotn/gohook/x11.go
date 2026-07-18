// Copyright 2016 The go-vgo Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

//go:build linux && purego && !wayland

// Package hook (Linux pure-Go X11 backend).
//
// This is a *CGo-free* event source built on github.com/jezek/xgb (a pure-Go
// X11 protocol client). It uses the X RECORD extension to observe every
// keyboard/mouse event delivered by the server — i.e. a true global hook, the
// same mechanism the CGo/libuiohook backend uses on X11. Select it at build
// time with the "purego" tag:
//
//	go build -tags purego .
//
// It replaces the default CGo/libuiohook backend (hook_cgo.go) on Linux and
// needs no C toolchain and no X11/Xtst development libraries. (Build with the
// "wayland" tag instead for the focused-surface Wayland backend; see
// wayland.go.)
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │  How it works                                                            │
// │                                                                          │
// │  The X RECORD extension requires two connections to the display: one     │
// │  *control* connection (create/free the record context) and one *data*    │
// │  connection that streams the intercepted events. jezek/xgb's cookie/     │
// │  reply machinery delivers only the first reply of a streaming request,   │
// │  so the data connection is driven over a raw socket here: we perform the │
// │  X11 connection handshake, send RECORD EnableContext, and read the       │
// │  reply stream directly. The control connection uses jezek/xgb normally.  │
// └──────────────────────────────────────────────────────────────────────────┘
package hook

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/record"
	"github.com/jezek/xgb/xproto"
)

// X keycodes are offset by 8 from Linux evdev codes on modern servers. The
// vcaesar Keycode map (used by Register) is keyed on evdev codes, so we expose
// Event.Keycode/Event.Rawcode as evdev = X keycode - evdevOffset.
const evdevOffset = 8

// EnableContext minor opcode in the RECORD extension (record.xml).
const recordEnableContext = 5

// X11 keyboard modifier bits (KeyButMask*, xproto). Used to fill Event.Mask
// and to pick the shifted keysym column for Keychar.
const (
	xShiftMask   = 1 << 0
	xLockMask    = 1 << 1 // caps lock
	xControlMask = 1 << 2
	xMod1Mask    = 1 << 3 // typically Alt
	xMod4Mask    = 1 << 6 // typically Super/Meta
)

// gohook virtual modifier masks (mirrors hook/iohook.h MASK_* for Event.Mask).
const (
	maskShiftL   uint16 = 1 << 0
	maskCtrlL    uint16 = 1 << 1
	maskMetaL    uint16 = 1 << 2
	maskAltL     uint16 = 1 << 3
	maskCapsLock uint16 = 1 << 14
)

// libuiohook-compatible scroll directions (matches the CGo backend).
const (
	wheelVertical   uint8 = 3
	wheelHorizontal uint8 = 4
)

// x11State holds the live connection objects for the running session so End()
// can tear them down. Guarded by the package-level lck mutex.
type x11State struct {
	ctrl *xgb.Conn // control connection (record context lifecycle, keymap)
	data net.Conn  // raw data connection streaming RECORD replies
	ctx  record.Context

	// keyboard mapping snapshot for keysym -> Keychar resolution.
	keysyms    []xproto.Keysym
	perCode    int
	minKeycode int

	// per-X-keycode pressed state, used to distinguish KeyDown vs KeyHold
	// (X delivers auto-repeat as additional KeyPress events).
	down map[byte]bool
}

var xst *x11State

// Start adds the X11 RECORD listener and returns the event channel.
//
// The optional timeout argument is accepted for API parity with the CGo
// backend but is ignored: this backend is event-driven (it blocks on the
// data socket) rather than polled.
func Start(tm ...int) chan Event {
	_ = tm

	ev = make(chan Event, 1024)
	asyncon = true

	go x11Loop()

	return ev
}

// End removes the X11 RECORD listener and resets all hook state.
func End(tm ...int) {
	tm1 := 10
	if len(tm) > 0 {
		tm1 = tm[0]
	}

	asyncon = false

	lck.Lock()
	st := xst
	xst = nil
	lck.Unlock()

	if st != nil {
		x11Teardown(st)
	}

	time.Sleep(time.Millisecond * time.Duration(tm1))

	for len(ev) != 0 {
		<-ev
	}
	close(ev)

	resetState()
}

// addEvent: the single-shot *blocking* listener (AddEvent/StopEvent) is a
// CGo/libuiohook-only feature with no direct equivalent here. The supported
// path on this backend is the channel API (Start + Register/Process). A
// non-zero code makes the public AddEvent report failure rather than silently
// pretending to register a hook.
func addEvent(key string) int {
	_ = key
	return -1
}

// StopEvent is a no-op on the pure-Go X11 backend (see addEvent).
func StopEvent() {}

// x11Loop opens the control connection, creates the RECORD context, opens the
// raw data connection and pumps the intercepted-event stream until End() tears
// the connections down.
func x11Loop() {
	ctrl, err := xgb.NewConn()
	if err != nil {
		// No X server / not an X session: report disabled and bail.
		send(Event{Kind: HookDisabled})
		return
	}

	if err := record.Init(ctrl); err != nil {
		send(Event{Kind: HookDisabled})
		ctrl.Close()
		return
	}

	ctx, err := record.NewContextId(ctrl)
	if err != nil {
		send(Event{Kind: HookDisabled})
		ctrl.Close()
		return
	}

	// Capture device events 2..6: KeyPress, KeyRelease, ButtonPress,
	// ButtonRelease, MotionNotify — from all clients (current and future).
	rng := record.Range{}
	rng.DeviceEvents = record.Range8{First: xproto.KeyPress, Last: xproto.MotionNotify}
	specs := []record.ClientSpec{record.ClientSpec(record.CsAllClients)}
	ranges := []record.Range{rng}

	if err := record.CreateContextChecked(ctrl, ctx, 0,
		uint32(len(specs)), uint32(len(ranges)), specs, ranges).Check(); err != nil {
		send(Event{Kind: HookDisabled})
		ctrl.Close()
		return
	}

	st := &x11State{ctrl: ctrl, ctx: ctx, down: make(map[byte]bool)}
	loadKeymap(st)

	data, err := x11DialAuth()
	if err != nil {
		send(Event{Kind: HookDisabled})
		record.FreeContext(ctrl, ctx)
		ctrl.Close()
		return
	}
	st.data = data

	lck.Lock()
	xst = st
	lck.Unlock()

	if err := x11EnableContext(data, recordOpcode(ctrl), ctx); err != nil {
		send(Event{Kind: HookDisabled})
		x11Teardown(st)
		return
	}

	send(Event{Kind: HookEnabled})

	x11ReadLoop(st)
}

// x11Teardown disables/frees the record context (over the control connection)
// and closes both connections. Closing the data socket unblocks x11ReadLoop's
// pending socket read.
func x11Teardown(st *x11State) {
	if st.ctrl != nil {
		record.DisableContext(st.ctrl, st.ctx)
		record.FreeContext(st.ctrl, st.ctx)
	}
	if st.data != nil {
		_ = st.data.Close()
	}
	if st.ctrl != nil {
		st.ctrl.Close()
	}
}

// recordOpcode returns the negotiated major opcode of the RECORD extension on
// the given connection (set by record.Init).
func recordOpcode(c *xgb.Conn) byte {
	c.ExtLock.RLock()
	defer c.ExtLock.RUnlock()
	return c.Extensions["RECORD"]
}

// loadKeymap snapshots the server keyboard mapping so key events can resolve a
// Keychar without a round-trip per keystroke. Best-effort: on failure Keychar
// is simply left undefined.
func loadKeymap(st *x11State) {
	setup := xproto.Setup(st.ctrl)
	if setup == nil {
		return
	}

	count := int(setup.MaxKeycode) - int(setup.MinKeycode) + 1
	if count <= 0 {
		return
	}

	reply, err := xproto.GetKeyboardMapping(st.ctrl, setup.MinKeycode, byte(count)).Reply()
	if err != nil || reply == nil {
		return
	}

	st.keysyms = reply.Keysyms
	st.perCode = int(reply.KeysymsPerKeycode)
	st.minKeycode = int(setup.MinKeycode)
}

// x11EnableContext writes a RECORD EnableContext request onto the raw data
// connection. The server then streams intercepted events back as a series of
// replies until the context is disabled.
func x11EnableContext(conn net.Conn, opcode byte, ctx record.Context) error {
	buf := make([]byte, 8)
	buf[0] = opcode
	buf[1] = recordEnableContext
	xgb.Put16(buf[2:], uint16(len(buf)/4)) // request length, 4-byte units
	xgb.Put32(buf[4:], uint32(ctx))

	_, err := conn.Write(buf)
	return err
}

// x11ReadLoop reads RECORD reply records off the raw data connection and
// dispatches the device events they carry. It returns when the connection is
// closed (by End()) or a read fails.
func x11ReadLoop(st *x11State) {
	header := make([]byte, 32)
	for asyncon {
		if _, err := io.ReadFull(st.data, header); err != nil {
			return
		}

		// We only expect replies (1) on the data connection. Errors (0) and
		// any stray events are exactly 32 bytes and need no extra reads.
		if header[0] != 1 {
			continue
		}

		// Reply: 32-byte header + Length*4 bytes of trailing data.
		length := int(xgb.Get32(header[4:]))
		var data []byte
		if length > 0 {
			data = make([]byte, length*4)
			if _, err := io.ReadFull(st.data, data); err != nil {
				return
			}
		}

		// Category 0 (FromServer) carries the intercepted device events.
		// StartOfData (4) and EndOfData (5) are skipped.
		if header[1] != 0 {
			continue
		}

		x11Dispatch(st, data)
	}
}

// x11Dispatch walks the 32-byte X event records packed in a RECORD data block
// and emits the corresponding gohook events.
func x11Dispatch(st *x11State, data []byte) {
	for i := 0; i+32 <= len(data); i += 32 {
		buf := data[i : i+32]
		switch buf[0] & 0x7f {
		case xproto.KeyPress:
			x11OnKey(st, buf, true)
		case xproto.KeyRelease:
			x11OnKey(st, buf, false)
		case xproto.ButtonPress:
			x11OnButton(st, buf, true)
		case xproto.ButtonRelease:
			x11OnButton(st, buf, false)
		case xproto.MotionNotify:
			x11OnMotion(buf)
		}
	}
}

// x11OnKey emits KeyDown/KeyHold/KeyUp from a recorded key event.
func x11OnKey(st *x11State, buf []byte, press bool) {
	ke := xproto.KeyPressEventNew(buf).(xproto.KeyPressEvent)
	xkc := byte(ke.Detail)

	var evdev uint16
	if int(xkc) >= evdevOffset {
		evdev = uint16(xkc) - evdevOffset
	}

	// CGo backend parity (hook/x11/hook_c.h): Rawcode carries the X keysym
	// resolved under the event's modifier state, while Keycode carries the
	// evdev code that Register()/Keycode matching relies on. This keeps
	// KeycharToRawcode()/RawcodeToKeychar() consistent across both backends.
	ks := st.keysymFor(xkc, ke.State)

	e := Event{
		Rawcode: uint16(ks),
		Keycode: evdev,
		Keychar: CharUndefined,
		Mask:    maskFromState(ke.State),
	}

	if press {
		lck.Lock()
		if st.down[xkc] {
			e.Kind = KeyHold
		} else {
			e.Kind = KeyDown
		}
		st.down[xkc] = true
		lck.Unlock()
	} else {
		e.Kind = KeyUp
		lck.Lock()
		delete(st.down, xkc)
		lck.Unlock()
	}

	if r := keysymToRune(ks); r != CharUndefined {
		e.Keychar = r
		if press {
			// Keep RawcodeToKeychar() in sync, mirroring the CGo backend
			// (extern.go go_send keys the map by the keysym rawcode).
			lck.Lock()
			raw2keyLinux[e.Rawcode] = string([]rune{r})
			lck.Unlock()
		}
	}

	send(e)
}

// x11OnButton emits MouseDown/MouseUp, or MouseWheel for the scroll-wheel
// pseudo-buttons (X buttons 4..7).
func x11OnButton(st *x11State, buf []byte, press bool) {
	be := xproto.ButtonPressEventNew(buf).(xproto.ButtonPressEvent)
	btn := byte(be.Detail)
	x, y := be.RootX, be.RootY
	mask := maskFromState(be.State)

	// X delivers wheel scrolls as button 4/5 (vertical) and 6/7 (horizontal)
	// press+release pairs. Emit a single MouseWheel on press; drop the release.
	if btn >= 4 && btn <= 7 {
		if !press {
			return
		}
		send(x11Wheel(btn, x, y, mask))
		return
	}

	kind := uint8(MouseDown)
	if !press {
		kind = MouseUp
	}

	send(Event{
		Kind:   kind,
		Button: x11Button(btn),
		Clicks: 1,
		X:      x,
		Y:      y,
		Mask:   mask,
	})
}

// x11OnMotion emits MouseMove using the absolute root-window coordinates that
// the recorded core motion event carries.
func x11OnMotion(buf []byte) {
	me := xproto.MotionNotifyEventNew(buf).(xproto.MotionNotifyEvent)
	send(Event{Kind: MouseMove, X: me.RootX, Y: me.RootY})
}

// x11Button maps an X core button number to a gohook MouseMap code.
func x11Button(btn byte) uint16 {
	switch btn {
	case 1:
		return MouseMap["left"]
	case 2:
		return MouseMap["center"] // X middle button
	case 3:
		return MouseMap["right"]
	default:
		return uint16(btn)
	}
}

// x11Wheel builds a MouseWheel Event for an X scroll pseudo-button.
func x11Wheel(btn byte, x, y int16, mask uint16) Event {
	e := Event{Kind: MouseWheel, X: x, Y: y, Clicks: 1, Amount: 1, Mask: mask}
	switch btn {
	case 4: // wheel up
		e.Direction = wheelVertical
		e.Rotation = WheelUp
	case 5: // wheel down
		e.Direction = wheelVertical
		e.Rotation = WheelDown
	case 6: // wheel left
		e.Direction = wheelHorizontal
		e.Rotation = WheelUp
	case 7: // wheel right
		e.Direction = wheelHorizontal
		e.Rotation = WheelDown
	}
	return e
}

// keysymFor resolves the keysym for an X keycode under the given modifier
// state (shifted column when Shift is held, with fallback to the unshifted
// column), or 0 (NoSymbol) when the keymap is unavailable.
func (st *x11State) keysymFor(xkc byte, state uint16) xproto.Keysym {
	col := 0
	if state&xShiftMask != 0 {
		col = 1
	}

	ks := st.keysymAt(xkc, col)
	if ks == 0 && col != 0 {
		ks = st.keysymAt(xkc, 0)
	}

	return ks
}

// keychar resolves the printable rune for an X keycode under the given
// modifier state, or CharUndefined when the key has no character.
func (st *x11State) keychar(xkc byte, state uint16) rune {
	return keysymToRune(st.keysymFor(xkc, state))
}

// keysymAt returns the keysym for an X keycode at the given column (0 =
// unshifted, 1 = shifted) from the cached keyboard mapping.
func (st *x11State) keysymAt(xkc byte, col int) xproto.Keysym {
	if st.perCode == 0 {
		return 0
	}
	idx := (int(xkc)-st.minKeycode)*st.perCode + col
	if idx < 0 || idx >= len(st.keysyms) {
		return 0
	}
	return st.keysyms[idx]
}

// keysymToRune converts an X11 keysym to a Unicode rune, returning
// CharUndefined for keysyms that are not printable characters (function keys,
// modifiers, etc.). See Appendix A of the X protocol and the keysymdef rules.
func keysymToRune(ks xproto.Keysym) rune {
	switch {
	case ks == 0:
		return CharUndefined
	case ks >= 0x20 && ks <= 0x7e: // ASCII printable
		return rune(ks)
	case ks >= 0xa0 && ks <= 0xff: // Latin-1 high range
		return rune(ks)
	case ks&0xff000000 == 0x01000000: // direct Unicode keysym
		return rune(ks & 0x00ffffff)
	default:
		return CharUndefined
	}
}

// maskFromState maps X11 modifier state bits to gohook's virtual mask.
func maskFromState(state uint16) uint16 {
	var m uint16
	if state&xShiftMask != 0 {
		m |= maskShiftL
	}
	if state&xControlMask != 0 {
		m |= maskCtrlL
	}
	if state&xMod4Mask != 0 {
		m |= maskMetaL
	}
	if state&xMod1Mask != 0 {
		m |= maskAltL
	}
	if state&xLockMask != 0 {
		m |= maskCapsLock
	}
	return m
}

// send timestamps and pushes an event onto the global channel. It drops the
// event (rather than blocking the reader) if no consumer keeps up and the
// buffer is full. The recover guards the small shutdown window where End() may
// have closed ev while a read is still in flight.
func send(e Event) {
	if !asyncon {
		return
	}
	defer func() { _ = recover() }() // ev closed by End(): drop silently

	e.When = time.Now()
	select {
	case ev <- e:
	default:
		// channel full: drop to avoid stalling the reader.
	}
}

// ---------------------------------------------------------------------------
// Raw X11 data connection: dial + handshake + MIT-MAGIC-COOKIE-1 auth.
//
// jezek/xgb performs this internally but does not export it, and its cookie
// machinery cannot stream a RECORD reply sequence. So the data connection is
// opened and authenticated here and read directly.
// ---------------------------------------------------------------------------

// x11DialAuth dials the X server named by $DISPLAY and completes the X11
// connection setup handshake, returning a ready-to-use socket.
func x11DialAuth() (net.Conn, error) {
	display := os.Getenv("DISPLAY")
	if display == "" {
		return nil, errors.New("hook: DISPLAY is not set")
	}

	conn, host, dispNum, err := x11Dial(display)
	if err != nil {
		return nil, err
	}

	if err := x11Handshake(conn, host, dispNum); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// x11Dial parses a DISPLAY string and connects to the X server, returning the
// socket along with the host and display number for authority lookup.
func x11Dial(display string) (conn net.Conn, host, dispNum string, err error) {
	colon := strings.LastIndex(display, ":")
	if colon < 0 {
		return nil, "", "", errors.New("hook: bad DISPLAY: " + display)
	}

	var socket, protocol string
	if display[0] == '/' {
		socket = display[:colon]
	} else if slash := strings.LastIndex(display, "/"); slash >= 0 {
		protocol = display[:slash]
		host = display[slash+1 : colon]
	} else {
		host = display[:colon]
	}

	rest := display[colon+1:]
	if rest == "" {
		return nil, "", "", errors.New("hook: bad DISPLAY: " + display)
	}
	dispNum = rest
	if dot := strings.LastIndex(rest, "."); dot >= 0 {
		dispNum = rest[:dot]
	}

	num, err := strconv.Atoi(dispNum)
	if err != nil || num < 0 {
		return nil, "", "", errors.New("hook: bad DISPLAY: " + display)
	}

	switch {
	case socket != "":
		conn, err = net.Dial("unix", socket+":"+dispNum)
	case host != "" && host != "unix":
		if protocol == "" {
			protocol = "tcp"
		}
		conn, err = net.Dial(protocol, host+":"+strconv.Itoa(6000+num))
	default:
		host = ""
		conn, err = net.Dial("unix", "/tmp/.X11-unix/X"+dispNum)
	}
	if err != nil {
		return nil, "", "", err
	}

	return conn, host, dispNum, nil
}

// x11Handshake performs the X11 connection setup: it sends the client setup
// (little-endian) with MIT-MAGIC-COOKIE-1 auth and verifies the server accepts
// it. The setup reply body is read and discarded — the control connection
// already exposes everything else we need.
func x11Handshake(conn net.Conn, host, dispNum string) error {
	name, data, err := x11Auth(host, dispNum)
	if err != nil {
		// Fall back to no authentication; the server may allow it.
		name, data = "", nil
	}

	buf := make([]byte, 12+xgb.Pad(len(name))+xgb.Pad(len(data)))
	buf[0] = 0x6c          // little-endian byte order
	xgb.Put16(buf[2:], 11) // protocol major version
	xgb.Put16(buf[4:], 0)  // protocol minor version
	xgb.Put16(buf[6:], uint16(len(name)))
	xgb.Put16(buf[8:], uint16(len(data)))
	copy(buf[12:], name)
	copy(buf[12+xgb.Pad(len(name)):], data)

	if _, err := conn.Write(buf); err != nil {
		return err
	}

	head := make([]byte, 8)
	if _, err := io.ReadFull(conn, head); err != nil {
		return err
	}

	code := head[0]
	reasonLen := int(head[1])
	extra := int(xgb.Get16(head[6:])) // additional data length, 4-byte units

	body := make([]byte, extra*4)
	if _, err := io.ReadFull(conn, body); err != nil {
		return err
	}

	// Setup status: 0 = Failed, 1 = Success, 2 = Authenticate (the server
	// wants an additional negotiation this client does not implement) —
	// anything but Success leaves the connection unusable.
	switch code {
	case 1:
		return nil
	case 2:
		return errors.New("hook: X server requires further authentication")
	default: // 0: Failed
		if reasonLen > len(body) {
			reasonLen = len(body)
		}
		return errors.New("hook: X authentication refused: " + string(body[:reasonLen]))
	}
}

// x11Auth reads $XAUTHORITY and returns the MIT-MAGIC-COOKIE-1 entry matching
// the host/display.
func x11Auth(host, dispNum string) (name string, data []byte, err error) {
	const (
		familyLocal = 256
		familyWild  = 65535
	)

	hostname := host
	if hostname == "" || hostname == "localhost" {
		hostname, err = os.Hostname()
		if err != nil {
			return "", nil, err
		}
	}

	fname := os.Getenv("XAUTHORITY")
	if fname == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", nil, errors.New("hook: neither XAUTHORITY nor HOME is set")
		}
		fname = home + "/.Xauthority"
	}

	f, err := os.Open(fname)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	scratch := make([]byte, 256)
	for {
		var family uint16
		if err := binary.Read(f, binary.BigEndian, &family); err != nil {
			return "", nil, err // io.EOF: no matching entry
		}

		addr, err := authString(f, scratch)
		if err != nil {
			return "", nil, err
		}
		disp, err := authString(f, scratch)
		if err != nil {
			return "", nil, err
		}
		entryName, err := authString(f, scratch)
		if err != nil {
			return "", nil, err
		}
		entryData, err := authBytes(f, scratch)
		if err != nil {
			return "", nil, err
		}

		addrMatch := family == familyWild ||
			(family == familyLocal && addr == hostname)
		dispMatch := disp == "" || disp == dispNum

		if addrMatch && dispMatch && entryName == "MIT-MAGIC-COOKIE-1" {
			cookie := make([]byte, len(entryData))
			copy(cookie, entryData)
			return entryName, cookie, nil
		}
	}
}

// authBytes reads a uint16-length-prefixed byte field from an Xauthority file
// into the scratch buffer (the returned slice aliases scratch).
func authBytes(r io.Reader, scratch []byte) ([]byte, error) {
	var n uint16
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	if int(n) > len(scratch) {
		return nil, errors.New("hook: Xauthority field too long")
	}
	if _, err := io.ReadFull(r, scratch[:n]); err != nil {
		return nil, err
	}
	return scratch[:n], nil
}

// authString reads a length-prefixed string field from an Xauthority file.
func authString(r io.Reader, scratch []byte) (string, error) {
	b, err := authBytes(r, scratch)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
