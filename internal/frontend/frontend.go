// Package frontend provides the UI-agnostic frontend registration mechanism.
// It mirrors the internal/plugins/plugins.go aggregator pattern: each frontend
// (TUI, headless, WebUI, ...) registers itself in its own init(), and the
// binary blank-imports the aggregator subpackage (registration) to trigger
// registration.
//
// The frontend package itself must never import internal/ui or
// internal/headless: both of them import frontend, so importing them here
// would create a compile-time cycle. The aggregator therefore lives in the
// registration subpackage.
package frontend

import (
	"context"
	"sync"
)

// LaunchOptions carries the CLI flags; a frontend builds its own engine inside
// Run (no Start(engine) shape).
type LaunchOptions struct {
	Once  string // only meaningful for headless
	Debug bool
	Pprof bool
}

// Frontend is a runnable frontend implementation.
type Frontend interface {
	ID() string
	Name() string
	Run(ctx context.Context, opts LaunchOptions) error
}

var (
	mu       sync.RWMutex
	registry = map[string]Frontend{}
	order    []string
)

// Register registers a frontend. It panics on a nil frontend or on a duplicate
// ID.
func Register(f Frontend) {
	if f == nil {
		panic("frontend: Register(nil)")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := registry[f.ID()]; ok {
		panic("frontend: duplicate registration for ID " + f.ID())
	}
	registry[f.ID()] = f
	order = append(order, f.ID())
}

// ByID returns the frontend registered under id.
func ByID(id string) (Frontend, bool) {
	mu.RLock()
	defer mu.RUnlock()
	f, ok := registry[id]
	return f, ok
}

// Registered returns the IDs of all registered frontends in registration order.
func Registered() []string {
	mu.RLock()
	defer mu.RUnlock()
	ids := make([]string, len(order))
	copy(ids, order)
	return ids
}
