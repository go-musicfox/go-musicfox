package framework

// NoopPlugin is a zero-boilerplate base for pure menu/command plugins: the
// lifecycle trio is a no-op. Embed or use it to satisfy the Plugin interface
// without writing empty Start/Stop/Dispose methods.
type NoopPlugin struct{}

// Start is a no-op satisfying the Plugin interface.
func (NoopPlugin) Start(*Context) error { return nil }

// Stop is a no-op satisfying the Plugin interface.
func (NoopPlugin) Stop() error { return nil }

// Dispose is a no-op satisfying the Plugin interface.
func (NoopPlugin) Dispose() error { return nil }
