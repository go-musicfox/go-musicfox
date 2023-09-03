//go:build !darwin

package player

import (
	"time"
)

type osxPlayer struct {
}

func NewOsxPlayer() Player {
	return &osxPlayer{}
}

func (p *osxPlayer) Play(_ UrlMusic) {
}

func (p *osxPlayer) CurMusic() UrlMusic {
	return UrlMusic{}
}

func (p *osxPlayer) Paused() {
}

func (p *osxPlayer) Resume() {
}

func (p *osxPlayer) Stop() {
}

func (p *osxPlayer) Toggle() {
}

func (p *osxPlayer) Seek(_ time.Duration) {
}

func (p *osxPlayer) PassedTime() time.Duration {
	return 0
}

func (p *osxPlayer) TimeChan() <-chan time.Duration {
	return nil
}

func (p *osxPlayer) State() State {
	return Unknown
}

func (p *osxPlayer) StateChan() <-chan State {
	return nil
}

func (p *osxPlayer) UpVolume() {
}

func (p *osxPlayer) DownVolume() {
}

func (p *osxPlayer) Volume() int {
	return 0
}

func (p *osxPlayer) SetVolume(volume int) {
}

func (p *osxPlayer) Close() {
}
