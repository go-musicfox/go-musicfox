//go:build enable_global_hotkey && purego

package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

func (h *EventHandler) RegisterGlobalHotkeys(opts *model.Options) {
	h.registerGlobalHotkeyBindings(opts)
	// With the purego backend, gohook manages its own logging;
	// no CGo logger override is needed.
}

// CloseGohookLogger is a no-op when using the purego backend.
func CloseGohookLogger() {
}
