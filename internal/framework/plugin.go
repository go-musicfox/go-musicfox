package framework

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
type Scope struct {
	parent   *Scope
	children []*Scope
	plugins  []Plugin
}

// Add registers a plugin into the scope. Plugins start in the order they are
// added and stop/dispose in reverse order.
func (s *Scope) Add(plugin Plugin) {
	s.plugins = append(s.plugins, plugin)
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
// are started after the receiver's own plugins.
func (s *Scope) Start(ctx *Context) error {
	for _, p := range s.plugins {
		if withDeps, ok := p.(PluginWithDeps); ok {
			if err := withDeps.Deps(ctx); err != nil {
				return err
			}
		}
		if err := p.Start(ctx); err != nil {
			return err
		}
	}
	for _, child := range s.children {
		if err := child.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the receiver's child scopes first, then its plugins, both in
// reverse order.
func (s *Scope) Stop() error {
	for i := len(s.children) - 1; i >= 0; i-- {
		if err := s.children[i].Stop(); err != nil {
			return err
		}
	}
	for i := len(s.plugins) - 1; i >= 0; i-- {
		if err := s.plugins[i].Stop(); err != nil {
			return err
		}
	}
	return nil
}

// Dispose recursively disposes child scopes before the receiver's own plugins
// (child before parent), then disposes plugins in reverse order, clears its
// state and detaches itself from its parent. Dispose is idempotent.
func (s *Scope) Dispose() error {
	for i := len(s.children) - 1; i >= 0; i-- {
		if err := s.children[i].Dispose(); err != nil {
			return err
		}
	}
	s.children = nil
	for i := len(s.plugins) - 1; i >= 0; i-- {
		if err := s.plugins[i].Dispose(); err != nil {
			return err
		}
	}
	s.plugins = nil
	if s.parent != nil {
		s.parent.removeChild(s)
		s.parent = nil
	}
	return nil
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
