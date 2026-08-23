// Package framework provides a lightweight, zero-dependency DI and plugin
// framework with cordis-like semantics: named service registration, resolution
// and overriding on a Context, scoped plugin lifecycle management with
// recursive child-scope cleanup, and an event chain dispatching through
// listener, middleware, parallel and serial handlers.
package framework

import "sort"

// Context is the Go degradation of the cordis Context: services are registered,
// resolved and overridden by name, stored as map[string]any and accessed via
// type assertion.
type Context struct {
	services map[string]any
}

// Names returns the sorted names of all registered services. It is mainly used
// by registration-completeness tests to assert the exact service set.
func (c *Context) Names() []string {
	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Service resolves the service registered under name; it returns nil when the
// service does not exist.
func (c *Context) Service(name string) any {
	if c == nil {
		return nil
	}
	return c.services[name]
}

// Provide registers a new service under name. It panics when the name is
// already registered; use Override to replace an existing service.
func (c *Context) Provide(name string, svc any) {
	if c.services == nil {
		c.services = make(map[string]any)
	}
	if _, ok := c.services[name]; ok {
		panic("framework: service " + name + " is already provided")
	}
	c.services[name] = svc
}

// Override replaces the service registered under name. It panics when the name
// has not been registered yet.
func (c *Context) Override(name string, svc any) {
	if _, ok := c.services[name]; !ok {
		panic("framework: cannot override unregistered service " + name)
	}
	c.services[name] = svc
}

// ServiceOf resolves the service registered under name and asserts it to type T.
// The boolean result reports whether the service exists and the type assertion
// succeeded.
func ServiceOf[T any](c *Context, name string) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}
	svc, ok := c.services[name]
	if !ok {
		return zero, false
	}
	typed, ok := svc.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}
