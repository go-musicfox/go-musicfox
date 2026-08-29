package framework

// PluginBase is a base for plugins that want to observe their own registered
// enabled state. Embed it (by value) and the scope registration methods
// (Add / AddWithEnabled / AddAndStart) write Enabled via the EnabledSetter
// method, so a plugin can check p.Enabled inside its lifecycle methods.
type PluginBase struct {
	Enabled bool
}

// EnabledSetter is implemented by plugins embedding PluginBase. Scope
// registration methods use it to record whether the plugin was added enabled.
type EnabledSetter interface {
	SetEnabled(enabled bool)
}

// SetEnabled records the enabled state. It is called by Scope.Add /
// AddWithEnabled / AddAndStart on plugins embedding PluginBase.
func (b *PluginBase) SetEnabled(enabled bool) { b.Enabled = enabled }

// The contributor interfaces below are lifecycle anchors plus assertion and
// documentation hooks. Real registration actions still happen inside a
// plugin's Start, because registration functions such as RegisterMenu are
// package-level generic functions and Go does not allow generic methods, so
// the interfaces cannot carry them.

// StartupHookContributor reports the plugin's startup hook, replacing the
// package-level startup-hook registry for scope-collected hooks
// ((Scope).StartupHooks collects them).
type StartupHookContributor interface {
	StartupHook() func()
}

// MenuContributor reports the menu keys a plugin contributes.
type MenuContributor interface {
	MenuKeys() []string
}

// PageContributor reports the page keys a plugin contributes.
type PageContributor interface {
	PageKeys() []string
}

// MainMenuContributor reports the main-menu entry keys a plugin contributes.
type MainMenuContributor interface {
	MainMenuKeys() []string
}

// CommandContributor reports the command keys a plugin contributes.
type CommandContributor interface {
	CommandKeys() []string
}

// ContextMenuContributor reports the context-menu keys a plugin contributes.
type ContextMenuContributor interface {
	ContextMenuKeys() []string
}

// PluginIdentity reports a plugin's stable id and display name for
// introspection (PluginInfos collection — the scope is the source of truth for
// the active plugin set, and this interface lets a scope plugin identify
// itself). It is optional: plugins may expose identity directly, or a
// frontend layer may wrap them with a decorator that provides it.
type PluginIdentity interface {
	PluginID() string
	PluginName() string
}
