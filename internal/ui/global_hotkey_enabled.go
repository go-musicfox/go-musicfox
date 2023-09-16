//go:build enable_global_hotkey

package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	hook "github.com/robotn/gohook"
)

func (h *EventHandler) RegisterGlobalHotkeys(opts *model.Options) {
	opts.GlobalKeyHandlers = make(map[string]model.GlobalKeyHandler)
	for global, operate := range configs.ConfigRegistry.GlobalHotkeys {
		opts.GlobalKeyHandlers[global] = func(event hook.Event) model.Page {
			_, page, _ := h.handle(OperateType(operate))
			return page
		}
	}
}
