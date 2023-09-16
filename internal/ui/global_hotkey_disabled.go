//go:build !enable_global_hotkey

package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

func (h *EventHandler) RegisterGlobalHotkeys(_ *model.Options) {

}
