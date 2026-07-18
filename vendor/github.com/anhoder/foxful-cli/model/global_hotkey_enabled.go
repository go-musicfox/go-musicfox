//go:build enable_global_hotkey

package model

import (
	"strings"

	hook "github.com/robotn/gohook"
)

type GlobalKeyHandler func(hook.Event) Page

// comboEntry holds the parsed keycodes and handler for a single hotkey combo.
type comboEntry struct {
	keycodes []uint16
	handler  GlobalKeyHandler
}

func ListenGlobalKeys(app *App, handlers map[string]GlobalKeyHandler) {
	// Parse combos: split "ctrl+shift+space" into individual keycodes.
	combos := make([]comboEntry, 0, len(handlers))
	for global, handler := range handlers {
		keys := strings.Split(global, "+")
		keycodes := make([]uint16, 0, len(keys))
		for _, k := range keys {
			if kc, ok := hook.Keycode[k]; ok {
				keycodes = append(keycodes, kc)
			}
		}
		if len(keycodes) == 0 {
			continue
		}
		combos = append(combos, comboEntry{keycodes: keycodes, handler: handler})
	}

	// Use the raw channel API directly to avoid gohook's buggy
	// keyRegistered check in Process() that skips combo callbacks.
	evChan := hook.Start()

	go func() {
		pressed := make(map[uint16]bool)
		for ev := range evChan {
			switch ev.Kind {
			case hook.KeyDown, hook.KeyHold:
				pressed[ev.Keycode] = true
			case hook.KeyUp:
				pressed[ev.Keycode] = false
			}

			// Only trigger on KeyDown to match the original Register(hook.KeyDown, ...) semantics.
			if ev.Kind != hook.KeyDown {
				continue
			}

			for _, c := range combos {
				if allKeysDown(pressed, c.keycodes) {
					page := c.handler(ev)
					if page == nil {
						page = app.page
					}
					app.program.Send(page.Msg())
				}
			}
		}
	}()
}

func allKeysDown(pressed map[uint16]bool, keycodes []uint16) bool {
	for _, kc := range keycodes {
		if !pressed[kc] {
			return false
		}
	}
	return true
}
