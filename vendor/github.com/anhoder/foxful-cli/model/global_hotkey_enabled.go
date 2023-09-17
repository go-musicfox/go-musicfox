//go:build enable_global_hotkey

package model

import (
	"strings"

	hook "github.com/robotn/gohook"
)

type GlobalKeyHandler func(hook.Event) Page

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
	hook.Process(s)
}
