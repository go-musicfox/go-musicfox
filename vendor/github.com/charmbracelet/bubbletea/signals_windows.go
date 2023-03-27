//go:build windows
// +build windows

package tea

import (
	"time"
)

// listenForResize is not available on windows because windows does not
// implement syscall.SIGWINCH.
func (p *Program) listenForResize(done chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer close(done)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
		}
		p.checkResize()
	}
}
