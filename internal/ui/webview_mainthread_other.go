//go:build !linux

package ui

// ServeWebviewMainThread is a no-op on platforms where the GTK login window is
// not used.
func ServeWebviewMainThread() {}
