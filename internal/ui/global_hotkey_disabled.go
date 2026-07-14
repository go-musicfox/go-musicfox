//go:build !enable_global_hotkey

package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

func (h *EventHandler) RegisterGlobalHotkeys(_ *model.Options) {

}

// CloseGohookLogger is a no-op when global hotkey is disabled.
func CloseGohookLogger() {
}
