//go:build windows
// +build windows

package tea

import "github.com/charmbracelet/x/term"

func (p *Program) checkOptimizedMovements(*term.State) {
	if !p.useHardTabsSet {
		p.useHardTabs = true
	}
	p.useBackspace = true
}
