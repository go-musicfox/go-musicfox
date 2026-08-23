package framework

import "errors"

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
type Scope struct {
	parent   *Scope
	children []*Scope
	plugins  []Plugin
	started  bool
	disposed bool
}

// NewScope creates a root scope. Root scopes have no parent; child scopes are
// created with (Scope).NewScope.
func NewScope() *Scope {
	return &Scope{}
}

// Add registers a plugin into the scope. Plugins start in the order they are
// added and stop/dispose in reverse order. Add returns an error when the scope
// has been disposed (a disposed scope is final and rejects new plugins).
func (s *Scope) Add(plugin Plugin) error {
	if s.disposed {
		return errors.New("framework: scope is disposed")
	}
	s.plugins = append(s.plugins, plugin)
	return nil
}

// NewScope creates a child scope and attaches it to the receiver. Child scopes
// follow the receiver's lifecycle: they are started with it, stopped with it
// and disposed before it.
func (s *Scope) NewScope() *Scope {
	child := &Scope{parent: s}
	s.children = append(s.children, child)
	return child
}

// Start starts all plugins in registration order. For plugins implementing
// PluginWithDeps, Deps is invoked first to inject dependencies. Child scopes
// are started after the receiver's own plugins. When any plugin or child scope
// fails to start, the already-started plugins and child scopes are rolled back
// (stopped in reverse start order) and the scope stays unstarted.
//
// Start is not idempotent: it returns an explicit error when the scope is
// already started or has been disposed.
func (s *Scope) Start(ctx *Context) error {
	if s.disposed {
		return errors.New("framework: scope is disposed")
	}
	if s.started {
		return errors.New("framework: scope is already started")
	}
	startedPlugins := 0
	for i, p := range s.plugins {
		if withDeps, ok := p.(PluginWithDeps); ok {
			if err := withDeps.Deps(ctx); err != nil {
				return errors.Join(err, s.rollback(startedPlugins, 0))
			}
		}
		if err := p.Start(ctx); err != nil {
			return errors.Join(err, s.rollback(startedPlugins, 0))
		}
		startedPlugins = i + 1
	}
	startedChildren := 0
	for i, child := range s.children {
		if err := child.Start(ctx); err != nil {
			return errors.Join(err, s.rollback(startedPlugins, startedChildren))
		}
		startedChildren = i + 1
	}
	s.started = true
	return nil
}

// rollback stops the successfully started child scopes (first startedChildren)
// and plugins (first startedPlugins) in reverse start order, mirroring Stop.
// Rollback is best-effort: stop failures are aggregated and returned so they
// can be surfaced alongside the original start error.
func (s *Scope) rollback(startedPlugins, startedChildren int) error {
	var errs []error
	for i := startedChildren - 1; i >= 0; i-- {
		if err := s.children[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	for i := startedPlugins - 1; i >= 0; i-- {
		if err := s.plugins[i].Stop(); err != nil {
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
func (s *Scope) Stop() error {
	if !s.started {
		return nil
	}
	var errs []error
	for i := len(s.children) - 1; i >= 0; i-- {
		if err := s.children[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	for i := len(s.plugins) - 1; i >= 0; i-- {
		if err := s.plugins[i].Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	s.started = false
	return errors.Join(errs...)
}

// Dispose recursively disposes child scopes before the receiver's own plugins
// (child before parent), then disposes plugins in reverse order, clears its
// state and detaches itself from its parent. A scope that is still started is
// implicitly stopped first (cordis dispose semantics), so child scopes and
// plugins receive Stop before Dispose; this prevents leaking plugins that only
// implement Stop.
//
// Dispose is idempotent and final. State cleanup (children/plugins cleared,
// parent detached) always happens even when individual Stop/Dispose calls fail
// (failures are aggregated with errors.Join and returned), and the disposed
// scope rejects later Start and Add calls.
func (s *Scope) Dispose() error {
	if s.disposed {
		return nil
	}
	var errs []error
	if s.started {
		if err := s.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	for i := len(s.children) - 1; i >= 0; i-- {
		if err := s.children[i].Dispose(); err != nil {
			errs = append(errs, err)
		}
	}
	s.children = nil
	for i := len(s.plugins) - 1; i >= 0; i-- {
		if err := s.plugins[i].Dispose(); err != nil {
			errs = append(errs, err)
		}
	}
	s.plugins = nil
	if s.parent != nil {
		s.parent.removeChild(s)
		s.parent = nil
	}
	s.disposed = true
	return errors.Join(errs...)
}

// removeChild detaches the given child scope from the receiver.
func (s *Scope) removeChild(child *Scope) {
	for i, c := range s.children {
		if c == child {
			s.children = append(s.children[:i], s.children[i+1:]...)
			return
		}
	}
}
