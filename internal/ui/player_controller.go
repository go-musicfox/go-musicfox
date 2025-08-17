package ui

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/remote_control"
)

var _ remote_control.Controller = (*Player)(nil)

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlPause() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPaused}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlResume() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlResume}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlStop() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlStop}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlToggle() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlToggle}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlNext() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlNext}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlPrevious() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPrevious}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSeek(duration time.Duration) {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlSeek, Duration: duration}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSetVolume(volume int) {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.Player.SetVolume(volume)
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlLikeNowPlaying() {
	
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlDislikeNowPlaying() {
	
}
