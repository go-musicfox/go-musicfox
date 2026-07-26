//go:build enable_global_hotkey

package model

import (
	"strings"

	hook "github.com/robotn/gohook"
)

type GlobalKeyHandler func(hook.Event) Page

// comboEntry holds the parsed keycodes and handler for a single hotkey combo.
// The combo is split into a trigger key (the last, non-modifier key) and the
// modifier keys that must be held while the trigger is pressed. A combo fires
// only when the trigger key's KeyDown arrives with all modifiers already held,
// so partial input (e.g. holding ctrl+shift without the trigger) never fires.
type comboEntry struct {
	trigger   uint16
	modifiers []uint16
	handler   GlobalKeyHandler
}

func ListenGlobalKeys(app *App, handlers map[string]GlobalKeyHandler) {
	// Parse combos: split "ctrl+shift+space" into individual keycodes. The last
	// key is the trigger; the preceding keys are modifiers held alongside it.
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
		combos = append(combos, comboEntry{
			trigger:   keycodes[len(keycodes)-1],
			modifiers: keycodes[:len(keycodes)-1],
			handler:   handler,
		})
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
				// Require the pressed key to be this combo's trigger. Relying on
				// the actual event key (rather than the persistent pressed map)
				// prevents a stale entry from a missed KeyUp from firing the combo
				// when only the modifiers are held.
				if ev.Keycode != c.trigger {
					continue
				}
				if allKeysDown(pressed, c.modifiers) {
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
