//go:build !windows

package player

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

type winMediaPlayer struct{}

func NewWinMediaPlayer() *winMediaPlayer {
	return &winMediaPlayer{}
}

func (p *winMediaPlayer) Play(_ URLMusic) {
}

func (p *winMediaPlayer) CurMusic() URLMusic {
	return URLMusic{}
}

func (p *winMediaPlayer) Paused() {
}

func (p *winMediaPlayer) Resume() {
}

func (p *winMediaPlayer) Stop() {
}

func (p *winMediaPlayer) Toggle() {
}

func (p *winMediaPlayer) Seek(_ time.Duration) {
}

func (p *winMediaPlayer) PassedTime() time.Duration {
	return 0
}

func (p *winMediaPlayer) TimeChan() <-chan time.Duration {
	return nil
}

func (p *winMediaPlayer) State() types.State {
	return types.Unknown
}

func (p *winMediaPlayer) StateChan() <-chan types.State {
	return nil
}

func (p *winMediaPlayer) UpVolume() {
}

func (p *winMediaPlayer) DownVolume() {
}

func (p *winMediaPlayer) Volume() int {
	return 0
}

func (p *winMediaPlayer) SetVolume(volume int) {
}

func (p *winMediaPlayer) Close() {
}
