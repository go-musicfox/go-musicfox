// Package kitty implements terminal image detection, encoding and output.
package kitty

import (
	"bytes"
	"io"
	"os"

	"github.com/charmbracelet/x/ansi"
)

// TmuxOutput batches final TUI writes containing Unicode image placeholders.
// Notifications and popups discard control sequences embedded in their input,
// so synchronization must be applied after composition and renderer line diffs.
// Embedding File preserves the terminal descriptor used by Bubble Tea.
type TmuxOutput struct {
	*os.File
}

func (w *TmuxOutput) Write(p []byte) (int, error) {
	if !bytes.Contains(p, []byte("\U0010eeee")) {
		return w.File.Write(p)
	}
	// tmux 3.7c omits the pane x offset when testing combining-character
	// visibility. Sync mode redraws complete cells instead of those updates.
	prefix, suffix := ansi.SetModeSynchronizedOutput, ansi.ResetModeSynchronizedOutput
	frame := make([]byte, 0, len(prefix)+len(p)+len(suffix))
	frame = append(frame, prefix...)
	frame = append(frame, p...)
	frame = append(frame, suffix...)
	n, err := w.File.Write(frame)
	if n < len(frame) && err == nil {
		err = io.ErrShortWrite
	}
	return min(len(p), max(0, n-len(prefix))), err
}

// WriteString must not use the promoted os.File method, which bypasses Write.
func (w *TmuxOutput) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
