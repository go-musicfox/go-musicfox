//go:build !enable_global_hotkey

package model

type GlobalKeyHandler interface{}

func ListenGlobalKeys(_ *App, _ map[string]GlobalKeyHandler) {
}
