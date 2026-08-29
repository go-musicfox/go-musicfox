// Track B (UI-agnostic command contribution) registration with plugin
// attribution.
//
// RegisterCommand is the ui-layer entry point for track-B commands: it stamps
// the current WithPlugin scope's plugin id onto cmd.PluginID (idempotent — an
// already-set PluginID is kept) and records the key under that plugin's
// PluginInfo.CommandKeys, so [plugins] disabled can later gate the command.
// Plugin-side code must always go through this function instead of
// frontend.RegisterCommand.
package ui

import (
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// RegisterCommand registers a track-B command and injects plugin attribution:
// when called inside a WithPlugin scope, currentPluginID is stamped onto
// cmd.PluginID (idempotent: a non-empty cmd.PluginID is never overwritten) and
// the key is recorded into PluginInfo.CommandKeys. Plugins must always use this
// function rather than frontend.RegisterCommand so [plugins] disabled can take
// effect. Registration semantics (empty Key / nil Run / duplicate Key panics)
// are delegated to frontend.RegisterCommand.
//
// Locking: currentPluginID is captured in a short pluginMu critical section
// (the record function re-locks pluginMu itself, so the lock must be released
// before calling it — sync.Mutex is not reentrant). frontend.RegisterCommand
// uses its own registry lock and is never called while holding pluginMu.
func RegisterCommand(cmd frontend.Command) {
	pluginMu.Lock()
	if cmd.PluginID == "" && currentPluginID != "" {
		cmd.PluginID = currentPluginID
	}
	pluginMu.Unlock()

	// frontend.RegisterCommand panics on a duplicate key before any side
	// effect, so the attribution record below is only made for registrations
	// that actually succeed (mirrors RegisterMenu's record-after-check order).
	frontend.RegisterCommand(cmd)
	recordPluginCommandKey(cmd.Key)
}
