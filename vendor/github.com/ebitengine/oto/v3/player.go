// Copyright 2021 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oto

import (
	"github.com/ebitengine/oto/v3/internal/mux"
)

// Player is a PCM (pulse-code modulation) audio player.
type Player struct {
	player *mux.Player
}

// Pause pauses its playing.
func (p *Player) Pause() {
	p.player.Pause()
}

// Play starts its playing if it doesn't play.
func (p *Player) Play() {
	p.player.Play()
}

// IsPlaying reports whether this player is playing.
func (p *Player) IsPlaying() bool {
	return p.player.IsPlaying()
}

// Reset clears the underyling buffer and pauses its playing.
//
// Deprecated: use Pause or Seek instead.
func (p *Player) Reset() {
	p.player.Reset()
}

// Volume returns the current volume in the range of [0, 1].
// The default volume is 1.
func (p *Player) Volume() float64 {
	return p.player.Volume()
}

// SetVolume sets the current volume in the range of [0, 1].
func (p *Player) SetVolume(volume float64) {
	p.player.SetVolume(volume)
}

// BufferedSize returns the byte size of the buffer data that is not sent to the audio hardware yet.
func (p *Player) BufferedSize() int {
	return p.player.BufferedSize()
}

// Err returns an error if this player has an error.
func (p *Player) Err() error {
	return p.player.Err()
}

// SetBufferSize sets the buffer size.
// If 0 is specified, the default buffer size is used.
func (p *Player) SetBufferSize(bufferSize int) {
	p.player.SetBufferSize(bufferSize)
}

// Seek implements io.Seeker.
//
// Seek returns an error when the underlying source doesn't implement io.Seeker.
func (p *Player) Seek(offset int64, whence int) (int64, error) {
	return p.player.Seek(offset, whence)
}

// Close implements io.Closer.
func (p *Player) Close() error {
	return p.player.Close()
}
