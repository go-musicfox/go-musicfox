package framework

import (
	"errors"
	"reflect"
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
