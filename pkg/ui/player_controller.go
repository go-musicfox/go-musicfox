package ui

import (
	"time"
)

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlPaused() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPaused}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlResume() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlResume}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlStop() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlStop}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlToggle() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlToggle}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlNext() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlNext}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlPrevious() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPrevious}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSeek(duration time.Duration) {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlSeek, Duration: duration}
}

// Deprecated: Only state_handler.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSetVolume(volume int) {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.Player.SetVolume(volume)
}
