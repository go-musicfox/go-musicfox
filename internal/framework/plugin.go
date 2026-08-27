package framework

import (
	"errors"
	"sync"
)

// Plugin is the plugin lifecycle interface: Start starts the plugin, Stop stops
// it and Dispose performs recursive cleanup. Following cordis semantics, a
// plugin may hold child scopes (sub-plugins); Dispose must recursively clean up
// child scopes as well.
type Plugin interface {
	Start(ctx *Context) error
	Stop() error
	Dispose() error
}

// PluginWithDeps is the optional interface for plugins that need dependency
// injection. When a plugin implements it, its Deps method is invoked with the
// context before Start so the plugin can resolve and wire its dependencies.
type PluginWithDeps interface {
	Plugin
	Deps(ctx *Context) error
}

// Scope groups plugins and child scopes under a unified lifecycle. Plugins
// start in registration order and stop/dispose in reverse order; child scopes
// follow the parent scope lifecycle and are disposed before the parent.
//
// Lifecycle contract:
//
//   - Start is not idempotent: calling it on an already-started scope returns
//     an explicit error. A scope whose Start failed partway stays unstarted
//     (rollback already stopped the started subset) and may be started again.
//   - Stop is safe to defer: stopping a scope that was never started is a
//     no-op returning nil.
//   - Dispose is final and idempotent: once disposed, a scope rejects Start
//     and Add with an explicit error. A scope that is still started is
//     implicitly stopped before cleanup, so child scopes and plugins receive
//     Stop then Dispose (cordis dispose semantics; prevents leaking plugins
//     that only implement Stop).
//
// Plugins registered with AddWithEnabled(..., false) are skipped by Start but
// stay in the plugin slice: Stop and Dispose still process them in order (a
// disabled plugin is never started yet still finalized), and Dispose stays
// idempotent.
//
// Concurrency: a simple mutex protects the started/disposed flags and the
// plugin/child slices so plugins can be added while the scope is running
// (AddAndStart, e.g. dynamic/WASM plugins). Plugin lifecycle calls
// (Deps/Start/Stop/Dispose) are always made outside the lock, so a plugin may
// safely re-enter its owning scope from its own lifecycle methods.
type Scope struct {
	parent   *Scope
	children []*Scope
	plugins  []Plugin
	enabled  []bool // parallel to plugins: whether the plugin starts with the scope
	started  bool
	disposed bool

	startupHooks []StartupHook // scope-collected hooks (RegisterStartupHookWithScope)

	mu sync.Mutex
}

// NewScope creates a root scope. Root scopes have no parent; child scopes are
// created with (Scope).NewScope.
func NewScope() *Scope {
	return &Scope{}
}

// addPlugin appends plugin with its enabled flag, rejecting registration on a
// disposed scope (a disposed scope is final and rejects new plugins). It also
// records the flag onto plugins embedding PluginBase (via EnabledSetter) so
// plugins can observe their own disabled state.
func (s *Scope) addPlugin(p Plugin, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.disposed {
		return errors.New("framework: scope is disposed")
	}
	s.plugins = append(s.plugins, p)
	s.enabled = append(s.enabled, enabled)
	if setter, ok := p.(EnabledSetter); ok {
		setter.SetEnabled(enabled)
	}
	return nil
}

// Add registers a plugin into the scope. Plugins start in the order they are
// added and stop/dispose in reverse order. Add returns an error when the scope
// has been disposed (a disposed scope is final and rejects new plugins).
func (s *Scope) Add(plugin Plugin) error {
	return s.addPlugin(plugin, true)
}

// AddWithEnabled registers a plugin and remembers whether it should start with
// the scope. A plugin registered with enabled=false stays in the scope's slice
// (it still receives Stop and Dispose in order) but its Start is skipped by
// (Scope).Start. Dispose stays idempotent: disabled plugins are finalized
// exactly once.
func (s *Scope) AddWithEnabled(p Plugin, enabled bool) error {
	return s.addPlugin(p, enabled)
}

// AddAndStart registers a plugin and, when the scope has already been started,
// immediately starts it (dynamic plugin/WASM hot-loading). When the scope has
// not been started yet, it degrades to Add: the plugin is registered and will
// start with the scope. On error the plugin is rolled back (Stop) and removed
// from the slice.
func (s *Scope) AddAndStart(ctx *Context, p Plugin) error {
	s.mu.Lock()
	if s.disposed {
		s.mu.Unlock()
		return errors.New("framework: scope is disposed")
	}
	s.plugins = append(s.plugins, p)
	s.enabled = append(s.enabled, true)
	if setter, ok := p.(EnabledSetter); ok {
		setter.SetEnabled(true)
	}
	started := s.started
	s.mu.Unlock()

	if !started {
		// Scope not started yet: degrade to Add.
		return nil
	}

	var startErr error
	if withDeps, ok := p.(PluginWithDeps); ok {
		startErr = withDeps.Deps(ctx)
	}
	if startErr == nil {
		startErr = p.Start(ctx)
	}
	if startErr != nil {
		return errors.Join(startErr, s.removeAndStop(p))
	}
	return nil
}

// removeAndStop removes p from the scope's plugin slice and stops it. It rolls
// back a failed AddAndStart.
func (s *Scope) removeAndStop(p Plugin) error {
	s.mu.Lock()
	for i, q := range s.plugins {
		if q == p {
			s.plugins = append(s.plugins[:i], s.plugins[i+1:]...)
			s.enabled = append(s.enabled[:i], s.enabled[i+1:]...)
			break
		}
	}
	s.mu.Unlock()
	return p.Stop()
}

// Plugins returns a snapshot of the registered plugins (debug/assert/PluginInfo
// collection). The returned slice does not alias the scope's live state.
func (s *Scope) Plugins() []Plugin {
	s.mu.Lock()
	defer s.mu.Unlock()
	plugins := make([]Plugin, len(s.plugins))
	copy(plugins, s.plugins)
	return plugins
}

// NewScope creates a child scope and attaches it to the receiver. Child scopes
// follow the receiver's lifecycle: they are started with it, stopped with it
// and disposed before it.
func (s *Scope) NewScope() *Scope {
	child := &Scope{parent: s}
	s.mu.Lock()
	s.children = append(s.children, child)
	s.mu.Unlock()
	return child
}

// Start starts all enabled plugins in registration order (plugins registered
// with AddWithEnabled(..., false) are skipped). For plugins implementing
// PluginWithDeps, Deps is invoked first to inject dependencies. Child scopes
// are started after the receiver's own plugins. When any plugin or child scope
// fails to start, the already-started plugins and child scopes are rolled back
// (stopped in reverse start order) and the scope stays unstarted.
//
// Start is not idempotent: it returns an explicit error when the scope is
// already started or has been disposed.
func (s *Scope) Start(ctx *Context) error {
	s.mu.Lock()
	if s.disposed {
		s.mu.Unlock()
		return errors.New("framework: scope is disposed")
	}
	if s.started {
		s.mu.Unlock()
		return errors.New("framework: scope is already started")
	}
	plugins := append([]Plugin(nil), s.plugins...)
	enabled := append([]bool(nil), s.enabled...)
	children := append([]*Scope(nil), s.children...)
	s.mu.Unlock()

	startedPlugins := 0
	for i, p := range plugins {
		if !enabled[i] {
			continue // disabled plugin: skip Start, stays in the slice
		}
		if withDeps, ok := p.(PluginWithDeps); ok {
			if err := withDeps.Deps(ctx); err != nil {
				return errors.Join(err, s.rollback(plugins, enabled, children, startedPlugins, 0))
			}
		}
		if err := p.Start(ctx); err != nil {
			return errors.Join(err, s.rollback(plugins, enabled, children, startedPlugins, 0))
		}
		startedPlugins = i + 1
	}
	startedChildren := 0
	for i, child := range children {
		if err := child.Start(ctx); err != nil {
			return errors.Join(err, s.rollback(plugins, enabled, children, startedPlugins, startedChildren))
		}
		startedChildren = i + 1
	}
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
	return nil
}

// rollback stops the successfully started child scopes (first startedChildren)
// and plugins (first startedPlugins) in reverse start order, mirroring Stop.
// Plugins registered disabled are skipped — they were never started. Rollback
// is best-effort: stop failures are aggregated and returned so they can be
// surfaced alongside the original start error.
func (s *Scope) rollback(plugins []Plugin, enabled []bool, children []*Scope, startedPlugins, startedChildren int) error {
	var errs []error
	for i := startedChildren - 1; i >= 0; i-- {
		if err := children[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	for i := startedPlugins - 1; i >= 0; i-- {
		if !enabled[i] {
			continue
		}
		if err := plugins[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Stop stops the receiver's child scopes first, then its plugins, both in
// reverse order. Stop is safe to defer: stopping a scope that was never
// started is a no-op returning nil. Stop failures are aggregated with
// errors.Join while the remaining children and plugins are still stopped, and
// the scope is marked as stopped regardless of failures.
//
// Stop processes every registered plugin, including ones registered disabled:
// a disabled plugin is never started but is still finalized in order by
// Stop/Dispose (its Stop is a no-op by contract).
func (s *Scope) Stop() error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	children := append([]*Scope(nil), s.children...)
	plugins := append([]Plugin(nil), s.plugins...)
	s.mu.Unlock()

	var errs []error
	for i := len(children) - 1; i >= 0; i-- {
		if err := children[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	for i := len(plugins) - 1; i >= 0; i-- {
		if err := plugins[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	s.mu.Lock()
	s.started = false
	s.mu.Unlock()
	return errors.Join(errs...)
}

// Dispose recursively disposes child scopes before the receiver's own plugins
// (child before parent), then disposes plugins in reverse order, clears its
// state and detaches itself from its parent. A scope that is still started is
// implicitly stopped first (cordis dispose semantics), so child scopes and
// plugins receive Stop before Dispose; this prevents leaking plugins that only
// implement Stop. Plugins registered disabled are disposed too: they were
// never started but are still finalized in order.
//
// Dispose is idempotent and final. State cleanup (children/plugins cleared,
// parent detached) always happens even when individual Stop/Dispose calls fail
// (failures are aggregated with errors.Join and returned), and the disposed
// scope rejects later Start and Add calls.
func (s *Scope) Dispose() error {
	s.mu.Lock()
	if s.disposed {
		s.mu.Unlock()
		return nil
	}
	started := s.started
	s.mu.Unlock()

	var errs []error
	if started {
		if err := s.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	s.mu.Lock()
	children := append([]*Scope(nil), s.children...)
	plugins := append([]Plugin(nil), s.plugins...)
	s.mu.Unlock()

	for i := len(children) - 1; i >= 0; i-- {
		if err := children[i].Dispose(); err != nil {
			errs = append(errs, err)
		}
	}
	s.mu.Lock()
	s.children = nil
	s.mu.Unlock()

	for i := len(plugins) - 1; i >= 0; i-- {
		if err := plugins[i].Dispose(); err != nil {
			errs = append(errs, err)
		}
	}
	s.mu.Lock()
	s.plugins = nil
	s.enabled = nil
	s.startupHooks = nil
	if s.parent != nil {
		s.parent.removeChild(s)
		s.parent = nil
	}
	s.disposed = true
	s.mu.Unlock()
	return errors.Join(errs...)
}

// removeChild detaches the given child scope from the receiver.
func (s *Scope) removeChild(child *Scope) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.children {
		if c == child {
			s.children = append(s.children[:i], s.children[i+1:]...)
			return
		}
	}
}
