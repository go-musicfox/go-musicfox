//go:build enable_global_hotkey

package model

import (
	"strings"
	"sync/atomic"

	hook "github.com/robotn/gohook"
)

type GlobalKeyHandler func(hook.Event) Page

var globalKeysStarted atomic.Bool

func ListenGlobalKeys(app *App, handlers map[string]GlobalKeyHandler) {
	for global, handler := range handlers {
		h := handler
		keys := strings.Split(global, "+")
		hook.Register(hook.KeyDown, keys, func(e hook.Event) {
			page := h(e)
			if page == nil {
				page = app.page
			}
			app.program.Send(page.Msg())
		})
	}
	s := hook.Start()
	globalKeysStarted.Store(true)
	hook.Process(s)
}

func stopGlobalKeys() {
	if globalKeysStarted.Swap(false) {
		hook.End()
	}
}
