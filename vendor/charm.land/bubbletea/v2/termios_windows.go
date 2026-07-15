//go:build windows
// +build windows

package tea

import "github.com/charmbracelet/x/term"

func (p *Program) checkOptimizedMovements(*term.State) {
	// Keep layout independent of terminal tab-stop configuration.
	p.useHardTabs = false
	p.useBackspace = true
}
