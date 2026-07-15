package client

import (
	"errors"
	"fmt"
	"net"
	"os"
)

type Context struct {
	conn      *net.UnixConn
	objects   map[uint32]Proxy
	currentID uint32
}

func (ctx *Context) Register(p Proxy) {
	ctx.currentID++
	p.SetID(ctx.currentID)
	p.SetContext(ctx)
	ctx.objects[ctx.currentID] = p
}

func (ctx *Context) Unregister(p Proxy) {
	delete(ctx.objects, p.ID())
}

func (ctx *Context) GetProxy(id uint32) Proxy {
	return ctx.objects[id]
}

func (ctx *Context) Close() error {
	return ctx.conn.Close()
}

// Dispatch reads and processes incoming messages and calls [client.Dispatcher.Dispatch] on the
// respective wayland protocol.
// Dispatch must be called on the same goroutine as other interactions with the Context.
// If a multi goroutine approach is desired, use [Context.GetDispatch] instead.
// Dispatch blocks if there are no incoming messages.
// A Dispatch loop is usually used to handle incoming messages.
func (ctx *Context) Dispatch() error {
	return ctx.GetDispatch()()
}

var ErrDispatchSenderNotFound = errors.New("dispatch: unable to find sender")
var ErrDispatchSenderUnsupported = errors.New("dispatch: sender does not implement Dispatch method")
var ErrDispatchUnableToReadMsg = errors.New("dispatch: unable to read msg")

// GetDispatch reads incoming messages and returns the dispatch function which calls
// [client.Dispatcher.Dispatch] on the respective wayland protocol.
// While GetDispatch is usually called in a loop in a separate goroutine, the dispatch function it
// returns must be called in the same goroutine as other interactions with the Context.
// GetDispatch blocks if there are no incoming messages.
func (ctx *Context) GetDispatch() func() error {
	senderID, opcode, fd, data, err := ctx.ReadMsg() // Blocks if there are no incoming messages
	if err != nil {
		return func() error {
			return fmt.Errorf("%w: %w", ErrDispatchUnableToReadMsg, err)
		}
	}

	return func() error {
		sender, ok := ctx.objects[senderID]
		if ok {
			if sender, ok := sender.(Dispatcher); ok {
				sender.Dispatch(opcode, fd, data)
			} else {
				return fmt.Errorf("%w (senderID=%d)", ErrDispatchSenderUnsupported, senderID)
			}
		} else {
			return fmt.Errorf("%w (senderID=%d)", ErrDispatchSenderNotFound, senderID)
		}

		return nil
	}
}

func Connect(addr string) (*Display, error) {
	if addr == "" {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			return nil, errors.New("env XDG_RUNTIME_DIR not set")
		}
		if addr == "" {
			addr = os.Getenv("WAYLAND_DISPLAY")
		}
		if addr == "" {
			addr = "wayland-0"
		}
		addr = runtimeDir + "/" + addr
	}

	ctx := &Context{
		objects: map[uint32]Proxy{},
	}

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
	if err != nil {
		return nil, err
	}
	ctx.conn = conn

	return NewDisplay(ctx), nil
}
