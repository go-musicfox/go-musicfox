// Copyright 2017 Hajime Hoshi
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

// Package oto offers io.Writer to play sound on multiple platforms.
package oto

import (
	"io"
	"runtime"
)

// Player is a PCM (pulse-code modulation) audio player.
// Player implements io.WriteCloser.
// Use Write method to play samples.
type Player struct {
	context *Context
	r       io.ReadCloser
	w       io.WriteCloser
}

func newPlayer(context *Context) *Player {
	r, w := pipe()
	p := &Player{
		context: context,
		r:       r,
		w:       w,
	}
	context.mux.AddSource(r)
	runtime.SetFinalizer(p, (*Player).Close)
	return p
}

// Write writes PCM samples to the Player.
//
// The format is as follows:
//   [data]      = [sample 1] [sample 2] [sample 3] ...
//   [sample *]  = [channel 1] ...
//   [channel *] = [byte 1] [byte 2] ...
// Byte ordering is little endian.
//
// The data is first put into the Player's buffer. Once the buffer is full, Player starts playing
// the data and empties the buffer.
//
// If the supplied data doesn't fit into the Player's buffer, Write block until a sufficient amount
// of data has been played (or at least started playing) and the remaining unplayed data fits into
// the buffer.
//
// Note, that the Player won't start playing anything until the buffer is full.
func (p *Player) Write(buf []byte) (int, error) {
	select {
	case err := <-p.context.errCh:
		return 0, err
	default:
	}
	n, err := p.w.Write(buf)
	// When the error is io.ErrClosedPipe, the context is already closed.
	if err == io.ErrClosedPipe {
		select {
		case err := <-p.context.errCh:
			return n, err
		default:
		}
	}
	return n, err
}

// Close closes the Player and frees any resources associated with it. The Player is no longer
// usable after calling Close.
func (p *Player) Close() error {
	runtime.SetFinalizer(p, nil)

	// Already closed
	if p.context == nil {
		return nil
	}

	select {
	case err := <-p.context.errCh:
		return err
	default:
	}

	// Close the pipe writer before RemoveSource, or Read-ing in the mux takes forever.
	if err := p.w.Close(); err != nil {
		return err
	}

	p.context.mux.RemoveSource(p.r)
	p.context = nil

	// Close the pipe reader after RemoveSource, or ErrClosedPipe happens at Read-ing.
	return p.r.Close()
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
