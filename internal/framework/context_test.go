package framework

import "testing"

func TestServiceResolvesRegisteredService(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("player", "beep")
	if got := ctx.Service("player"); got != "beep" {
		t.Fatalf("Service() = %v, want %q", got, "beep")
	}
}

func TestServiceMissingReturnsNil(t *testing.T) {
	ctx := &Context{}
	if got := ctx.Service("missing"); got != nil {
		t.Fatalf("Service() = %v, want nil", got)
	}
}

func TestProvideRegistersMultipleServices(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("a", 1)
	ctx.Provide("b", "two")
	if ctx.Service("a") != 1 {
		t.Fatalf("Service(a) = %v, want 1", ctx.Service("a"))
	}
	if ctx.Service("b") != "two" {
		t.Fatalf("Service(b) = %v, want %q", ctx.Service("b"), "two")
	}
}

func TestProvideDuplicatePanics(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("svc", 1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate Provide")
		}
	}()
	ctx.Provide("svc", 2)
}

func TestOverrideReplacesExistingService(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("svc", "old")
	ctx.Override("svc", "new")
	if got := ctx.Service("svc"); got != "new" {
		t.Fatalf("Service() = %v, want %q", got, "new")
	}
}

func TestOverrideUnregisteredPanics(t *testing.T) {
	ctx := &Context{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on Override of unregistered service")
		}
	}()
	ctx.Override("missing", 1)
}

func TestServiceOfTypeAssertionSuccess(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("greeting", "hello")
	got, ok := ServiceOf[string](ctx, "greeting")
	if !ok {
		t.Fatal("ServiceOf: expected ok=true")
	}
	if got != "hello" {
		t.Fatalf("ServiceOf: got %q, want %q", got, "hello")
	}
}

func TestServiceOfTypeAssertionFailure(t *testing.T) {
	ctx := &Context{}
	ctx.Provide("greeting", "hello")
	if got, ok := ServiceOf[int](ctx, "greeting"); ok {
		t.Fatalf("ServiceOf: expected ok=false, got %v", got)
	}
}

func TestServiceOfMissingService(t *testing.T) {
	ctx := &Context{}
	if got, ok := ServiceOf[string](ctx, "missing"); ok {
		t.Fatalf("ServiceOf: expected ok=false, got %q", got)
	}
}
