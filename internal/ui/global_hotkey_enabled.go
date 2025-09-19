//go:build enable_global_hotkey

package ui

import (
	"fmt"
	"log/slog"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	hook "github.com/robotn/gohook"
)

func (h *EventHandler) RegisterGlobalHotkeys(opts *model.Options) {
	opts.GlobalKeyHandlers = make(map[string]model.GlobalKeyHandler)
	for global, operate := range configs.AppConfig.Keybindings.Global {
		ot, ok := keybindings.GetOperationFromName(operate)
		if !ok {
			slog.Warn(fmt.Sprintf("无效的操作：'%s'，忽略全局快捷键 '%s'", operate, global))
			continue
		}
		opts.GlobalKeyHandlers[global] = func(event hook.Event) model.Page {
			_, page, _ := h.handle(ot)
			return page
		}
	}
}
