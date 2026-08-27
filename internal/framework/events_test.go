package framework

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// eventRecorder records handler invocations in a concurrency-safe way.
type eventRecorder struct {
	mu      sync.Mutex
	entries []string
}

func (r *eventRecorder) add(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, s)
}

func (r *eventRecorder) list() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.entries))
	copy(out, r.entries)
	return out
}

func assertRecorder(t *testing.T, r *eventRecorder, want []string) {
	t.Helper()
	if got := r.list(); !reflect.DeepEqual(got, want) {
		t.Fatalf("recorded = %v, want %v", got, want)
	}
}

func TestListenerForwardOrder(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l1"); return nil })
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l2"); return nil })
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l3"); return nil })

	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	assertRecorder(t, r, []string{"l1", "l2", "l3"})
}

func TestListenerErrorInterruptsRemainingKinds(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l1"); return nil })
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l2"); return errors.New("boom") })
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l3"); return nil })
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error { r.add("mw"); return next() })
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("par"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("ser"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("Emit() error = %v, want %q", err, "boom")
	}
	assertRecorder(t, r, []string{"l1", "l2"})
}

func TestMiddlewareOnionOrder(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		r.add("in:mw1")
		if err := next(); err != nil {
			return err
		}
		r.add("out:mw1")
		return nil
	})
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		r.add("in:mw2")
		if err := next(); err != nil {
			return err
		}
		r.add("out:mw2")
		return nil
	})

	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	assertRecorder(t, r, []string{"in:mw1", "in:mw2", "out:mw2", "out:mw1"})
}

func TestMiddlewareLastNextIsNoop(t *testing.T) {
	emitter := NewEventEmitter()
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		if err := next(); err != nil {
			return err
		}
		return nil
	})
	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v, want nil", err)
	}
}

func TestMiddlewareErrorStopsOuterAfterReturn(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		r.add("in:mw1")
		if err := next(); err != nil {
			return err // propagate: the outer "out:mw1" must not run
		}
		r.add("out:mw1")
		return nil
	})
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		r.add("in:mw2")
		return errors.New("mw2 failed")
	})

	err := emitter.Emit(&Context{}, "evt", nil)
	if err == nil || err.Error() != "mw2 failed" {
		t.Fatalf("Emit() error = %v, want %q", err, "mw2 failed")
	}
	assertRecorder(t, r, []string{"in:mw1", "in:mw2"})
}

func TestMiddlewareErrorSkipsLaterKinds(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		return errors.New("mw failed")
	})
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("par"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("ser"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	if err == nil {
		t.Fatal("Emit() error = nil, want non-nil")
	}
	if got := r.list(); len(got) != 0 {
		t.Fatalf("parallel/serial must not run after middleware error, got %v", got)
	}
}

func TestParallelHandlersRunConcurrentlyAndEmitWaits(t *testing.T) {
	emitter := NewEventEmitter()
	var started atomic.Int32
	release := make(chan struct{})
	emitDone := make(chan error, 1)

	handler := func(ctx *Context, payload any) error {
		started.Add(1)
		<-release
		return nil
	}
	emitter.Parallel("evt", handler)
	emitter.Parallel("evt", handler)

	go func() { emitDone <- emitter.Emit(&Context{}, "evt", nil) }()

	select {
	case err := <-emitDone:
		t.Fatalf("Emit() returned before handlers released: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	if got := started.Load(); got != 2 {
		close(release)
		t.Fatalf("started = %d, want 2 (handlers must run concurrently)", got)
	}
	close(release)

	select {
	case err := <-emitDone:
		if err != nil {
			t.Fatalf("Emit() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Emit() did not return after handlers completed")
	}
}

func TestParallelCollectsFirstError(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("p1"); return nil })
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("p2"); return errors.New("p2 failed") })
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("p3"); return errors.New("p3 failed") })

	err := emitter.Emit(&Context{}, "evt", nil)
	if err == nil {
		t.Fatal("Emit() error = nil, want non-nil")
	}
	// All parallel handlers must have run despite errors.
	if got := r.list(); len(got) != 3 {
		t.Fatalf("recorded %v, want all 3 parallel handlers to run", got)
	}
}

// TestParallelReadOnlyServiceResolutionIsRaceSafe exercises the documented
// parallel concurrency contract: handlers share the Context read-only and only
// resolve services concurrently. Run with -race; a write from any handler
// would be reported as a data race.
func TestParallelReadOnlyServiceResolutionIsRaceSafe(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("greeting", "hello")
	ctx.Provide("answer", 42)

	emitter := NewEventEmitter()
	const handlers = 32
	for i := 0; i < handlers; i++ {
		emitter.Parallel("evt", func(ctx *Context, _ any) error {
			if got := ctx.Service("greeting"); got != "hello" {
				t.Errorf("Service(greeting) = %v, want %q", got, "hello")
			}
			if got, ok := ServiceOf[int](ctx, "answer"); !ok || got != 42 {
				t.Errorf("ServiceOf(answer) = %v (ok=%v), want 42", got, ok)
			}
			return nil
		})
	}
	if err := emitter.Emit(ctx, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
}

func TestSerialForwardOrder(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s1"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s2"); return nil })

	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	assertRecorder(t, r, []string{"s1", "s2"})
}

func TestSerialErrorInterruptsRemaining(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s1"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s2"); return errors.New("s2 failed") })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s3"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	if err == nil || err.Error() != "s2 failed" {
		t.Fatalf("Emit() error = %v, want %q", err, "s2 failed")
	}
	assertRecorder(t, r, []string{"s1", "s2"})
}

func TestEmitDispatchesKindsInOrder(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("listener"); return nil })
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		r.add("mw:in")
		if err := next(); err != nil {
			return err
		}
		r.add("mw:out")
		return nil
	})
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("parallel"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("serial"); return nil })

	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	assertRecorder(t, r, []string{"listener", "mw:in", "mw:out", "parallel", "serial"})
}

func TestEmitPassesPayload(t *testing.T) {
	emitter := NewEventEmitter()
	var got any
	emitter.Listener("evt", func(ctx *Context, payload any) error { got = payload; return nil })

	payload := map[string]int{"track": 42}
	if err := emitter.Emit(&Context{}, "evt", payload); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	if !reflect.DeepEqual(got, payload) {
		t.Fatalf("payload = %v, want %v", got, payload)
	}
}

func TestEmitUnregisteredEventIsNoop(t *testing.T) {
	emitter := NewEventEmitter()
	if err := emitter.Emit(&Context{}, "never-registered", nil); err != nil {
		t.Fatalf("Emit() error = %v, want nil", err)
	}
}

// --- Unregister ---

func TestUnregisterRemovesAllHandlerKinds(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l"); return nil })
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error { r.add("mw"); return next() })
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("par"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("ser"); return nil })

	emitter.Unregister("evt")
	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() after Unregister error = %v, want nil", err)
	}
	if got := r.list(); len(got) != 0 {
		t.Fatalf("recorded %v after Unregister, want no handler to run", got)
	}
}

func TestUnregisterUnknownNameIsNoop(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l"); return nil })
	emitter.Unregister("never-registered")
	if err := emitter.Emit(&Context{}, "evt", nil); err != nil {
		t.Fatalf("Emit() error = %v, want nil", err)
	}
	assertRecorder(t, r, []string{"l"})
}

func TestUnregisterOnlyRemovesTargetName(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("a", func(ctx *Context, payload any) error { r.add("a"); return nil })
	emitter.Listener("b", func(ctx *Context, payload any) error { r.add("b"); return nil })
	emitter.Unregister("a")
	if err := emitter.Emit(&Context{}, "b", nil); err != nil {
		t.Fatalf("Emit(b) error = %v", err)
	}
	assertRecorder(t, r, []string{"b"})
}

// --- panic isolation ---

// assertPanicError asserts err is non-nil and reports the panicking handler
// kind for event name "evt".
func assertPanicError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Emit() error = nil, want non-nil (handler panicked)")
	}
	if !strings.Contains(err.Error(), "evt handler panicked") {
		t.Fatalf("Emit() error = %v, want %q", err, "evt handler panicked")
	}
}

func TestListenerPanicIsIsolated(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l1"); return nil })
	emitter.Listener("evt", func(ctx *Context, payload any) error { panic("listener boom") })
	emitter.Listener("evt", func(ctx *Context, payload any) error { r.add("l3"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	assertPanicError(t, err)
	// The panic follows returned-error semantics: the chain is aborted and the
	// remaining listeners must not run.
	assertRecorder(t, r, []string{"l1"})
}

func TestMiddlewarePanicIsIsolated(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		r.add("in:mw1")
		if err := next(); err != nil {
			return err
		}
		r.add("out:mw1")
		return nil
	})
	emitter.Middleware("evt", func(ctx *Context, payload any, next func() error) error {
		panic("middleware boom")
	})
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("par"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	assertPanicError(t, err)
	// The onion is aborted: the outer middleware's out phase and the parallel
	// kind must not run.
	assertRecorder(t, r, []string{"in:mw1"})
}

func TestSerialPanicIsIsolated(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s1"); return nil })
	emitter.Serial("evt", func(ctx *Context, payload any) error { panic("serial boom") })
	emitter.Serial("evt", func(ctx *Context, payload any) error { r.add("s3"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	assertPanicError(t, err)
	assertRecorder(t, r, []string{"s1"})
}

func TestParallelPanicIsIsolatedAndOthersRun(t *testing.T) {
	emitter := NewEventEmitter()
	r := &eventRecorder{}
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("p1"); return nil })
	emitter.Parallel("evt", func(ctx *Context, payload any) error { panic("parallel boom") })
	emitter.Parallel("evt", func(ctx *Context, payload any) error { r.add("p3"); return nil })

	err := emitter.Emit(&Context{}, "evt", nil)
	assertPanicError(t, err)
	// All parallel handlers still run to completion: both non-panicking
	// handlers must have been invoked.
	if got := r.list(); len(got) != 2 {
		t.Fatalf("recorded %v, want both non-panicking parallel handlers to run", got)
	}
}
