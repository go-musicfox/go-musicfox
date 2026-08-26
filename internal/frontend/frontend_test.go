package frontend

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// fakeFrontend is a minimal Frontend implementation for tests.
type fakeFrontend struct {
	id   string
	name string
}

func (f fakeFrontend) ID() string   { return f.id }
func (f fakeFrontend) Name() string { return f.name }
func (f fakeFrontend) Run(_ context.Context, _ LaunchOptions) error {
	return errors.New("fake frontend run")
}

// resetRegistry clears the package-level registry so each test controls its
// own registration state.
func resetRegistry() {
	mu.Lock()
	defer mu.Unlock()
	registry = map[string]Frontend{}
	order = nil
}

func TestRegisterByID(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	f := fakeFrontend{id: "byid-a", name: "ByID A"}
	Register(f)

	got, ok := ByID("byid-a")
	if !ok {
		t.Fatal("ByID miss after Register")
	}
	if got != f {
		t.Fatalf("ByID() = %v, want %v", got, f)
	}

	if _, ok := ByID("missing"); ok {
		t.Fatal("ByID hit for unregistered ID")
	}
}

func TestRegisteredInRegistrationOrder(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	ids := []string{"order-a", "order-b", "order-c"}
	for _, id := range ids {
		Register(fakeFrontend{id: id, name: id})
	}

	got := Registered()
	if !reflect.DeepEqual(got, ids) {
		t.Fatalf("Registered() = %v, want %v", got, ids)
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	f := fakeFrontend{id: "dup-a", name: "Dup A"}
	Register(f)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	Register(f)
}

func TestRegisterNilPanics(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil registration")
		}
	}()
	Register(nil)
}
