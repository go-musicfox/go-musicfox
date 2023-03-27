package termenv

import (
	"github.com/aymanbagabas/go-osc52"
)

func (o Output) osc52Output() *osc52.Output {
	return osc52.NewOutput(o.tty, o.environ.Environ())
}

// Copy copies text to clipboard using OSC 52 escape sequence.
func (o Output) Copy(str string) {
	o.osc52Output().Copy(str)
}

// CopyPrimary copies text to primary clipboard (X11) using OSC 52 escape
// sequence.
func (o Output) CopyPrimary(str string) {
	o.osc52Output().CopyPrimary(str)
}

// Copy copies text to clipboard using OSC 52 escape sequence.
func Copy(str string) {
	output.Copy(str)
}

// CopyPrimary copies text to primary clipboard (X11) using OSC 52 escape
// sequence.
func CopyPrimary(str string) {
	output.CopyPrimary(str)
}
