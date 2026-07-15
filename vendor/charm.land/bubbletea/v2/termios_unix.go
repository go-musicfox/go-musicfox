//go:build darwin || linux || solaris || aix
// +build darwin linux solaris aix

package tea

import (
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

func (p *Program) checkOptimizedMovements(s *term.State) {
	// Hard tabs turn precise space-based layout into terminal-dependent cursor moves.
	// Disable them so TUI column positions remain stable across tab-stop settings.
	p.useHardTabs = false
	p.useBackspace = s.Lflag&unix.BSDLY == unix.BS0
}
