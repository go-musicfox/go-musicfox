package framework

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// EventEmitter dispatches named events through four kinds of handlers,
// mirroring the cordis event API semantics:
//
//   - listener: simple listeners called in registration order; an error stops
//     the remaining listeners.
//   - middleware: onion-style middlewares called in registration order inward
//     and reverse order outward; an error interrupts and propagates upward.
//   - parallel: all handlers are invoked concurrently in goroutines and the
//     first error is collected.
//   - serial: handlers are invoked one by one and the first non-nil error is
//     returned immediately.
//
// Handler panics are isolated: every invocation is wrapped in a recover that
// converts the panic into a returned error carrying a short stack trace, so
// the emitter itself never panics because of a handler. A panic follows the
// same semantics as a returned error: listener/middleware/serial panics abort
// the chain, while a parallel panic is delivered through the error channel and
// all parallel handlers still run to completion.
type EventEmitter struct {
	listeners  map[string][]listenerFunc
	middleware map[string][]middlewareFunc
	parallel   map[string][]parallelFunc
	serial     map[string][]serialFunc
}

type listenerFunc func(ctx *Context, payload any) error
type middlewareFunc func(ctx *Context, payload any, next func() error) error
type parallelFunc func(ctx *Context, payload any) error
type serialFunc func(ctx *Context, payload any) error

// handlerPanicError converts a recovered panic from a handler into a framework
// error carrying a short stack trace (truncated to ~1KB) for diagnosability.
func handlerPanicError(name string, r any) error {
	const maxStackBytes = 1024
	stack := debug.Stack()
	if len(stack) > maxStackBytes {
		stack = stack[:maxStackBytes]
	}
	return fmt.Errorf("framework: %s handler panicked: %v\n%s", name, r, stack)
}

// invokeHandler runs fn, converting any panic it raises into a returned error
// so a misbehaving handler can never crash the emitter.
func invokeHandler(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = handlerPanicError(name, r)
		}
	}()
	return fn()
}

// NewEventEmitter creates an empty event emitter.
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{}
}

// Listener registers a simple listener under name.
func (e *EventEmitter) Listener(name string, fn func(ctx *Context, payload any) error) {
	if e.listeners == nil {
		e.listeners = make(map[string][]listenerFunc)
	}
	e.listeners[name] = append(e.listeners[name], fn)
}

// Middleware registers an onion-style middleware under name.
func (e *EventEmitter) Middleware(name string, fn func(ctx *Context, payload any, next func() error) error) {
	if e.middleware == nil {
		e.middleware = make(map[string][]middlewareFunc)
	}
	e.middleware[name] = append(e.middleware[name], fn)
}

// Parallel registers a concurrently-invoked handler under name. Parallel
// handlers run in their own goroutines and share the emitting Context
// read-only: resolving services (Context.Service / ServiceOf) is safe
// concurrently, but mutating the Context (Provide / Override) from a parallel
// handler races against the other handlers and is unsupported.
func (e *EventEmitter) Parallel(name string, fn func(ctx *Context, payload any) error) {
	if e.parallel == nil {
		e.parallel = make(map[string][]parallelFunc)
	}
	e.parallel[name] = append(e.parallel[name], fn)
}

// Serial registers a serially-invoked handler under name.
func (e *EventEmitter) Serial(name string, fn func(ctx *Context, payload any) error) {
	if e.serial == nil {
		e.serial = make(map[string][]serialFunc)
	}
	e.serial[name] = append(e.serial[name], fn)
}

// Emit dispatches the event named name to all four kinds of registered
// handlers in order: listeners (forward), middlewares (onion), parallel
// (concurrent) and serial (forward). The first error returned by any kind
// stops the remaining kinds and is returned. Emitting an unregistered event is
// a no-op that returns nil. A panicking handler is converted into an error and
// follows the same abort semantics as a returned error; Emit never panics
// because of a handler.
func (e *EventEmitter) Emit(ctx *Context, name string, payload any) error {
	for _, fn := range e.listeners[name] {
		if err := invokeHandler(name, func() error { return fn(ctx, payload) }); err != nil {
			return err
		}
	}

	if mws := e.middleware[name]; len(mws) > 0 {
		var run func(i int) error
		run = func(i int) error {
			if i >= len(mws) {
				return nil
			}
			return invokeHandler(name, func() error {
				return mws[i](ctx, payload, func() error { return run(i + 1) })
			})
		}
		if err := run(0); err != nil {
			return err
		}
	}

	if fns := e.parallel[name]; len(fns) > 0 {
		// Concurrency contract: each handler runs in its own goroutine and they
		// all share the same ctx read-only. Resolution (Service/ServiceOf) is
		// safe; Provide/Override from a parallel handler is a race and
		// unsupported. Errors are best-effort collected (the first one read
		// back is returned); all handlers still run to completion.
		var wg sync.WaitGroup
		errCh := make(chan error, len(fns))
		for _, fn := range fns {
			wg.Add(1)
			go func(fn parallelFunc) {
				defer wg.Done()
				if err := invokeHandler(name, func() error { return fn(ctx, payload) }); err != nil {
					errCh <- err
				}
			}(fn)
		}
		wg.Wait()
		close(errCh)
		if err := <-errCh; err != nil {
			return err
		}
	}

	for _, fn := range e.serial[name] {
		if err := invokeHandler(name, func() error { return fn(ctx, payload) }); err != nil {
			return err
		}
	}

	return nil
}
