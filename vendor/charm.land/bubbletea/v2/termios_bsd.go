//go:build dragonfly || freebsd
// +build dragonfly freebsd

package tea

import "github.com/charmbracelet/x/term"

func (p *Program) checkOptimizedMovements(s *term.State) {
	// Keep layout independent of terminal tab-stop configuration.
	p.useHardTabs = false
}
