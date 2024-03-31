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
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3/internal/mux"
)

var (
	contextCreated       bool
	contextCreationMutex sync.Mutex
)

// Context is the main object in Oto. It interacts with the audio drivers.
//
// To play sound with Oto, first create a context. Then use the context to create
// an arbitrary number of players. Then use the players to play sound.
//
// Creating multiple contexts is NOT supported.
type Context struct {
	context *context
}

// Format is the format of sources.
type Format int

const (
	// FormatFloat32LE is the format of 32 bits floats little endian.
	FormatFloat32LE Format = iota

	// FormatUnsignedInt8 is the format of 8 bits integers.
	FormatUnsignedInt8

	//FormatSignedInt16LE is the format of 16 bits integers little endian.
	FormatSignedInt16LE
)

// NewContextOptions represents options for NewContext.
type NewContextOptions struct {
	// SampleRate specifies the number of samples that should be played during one second.
	// Usual numbers are 44100 or 48000. One context has only one sample rate. You cannot play multiple audio
	// sources with different sample rates at the same time.
	SampleRate int

	// ChannelCount specifies the number of channels. One channel is mono playback. Two
	// channels are stereo playback. No other values are supported.
	ChannelCount int

	// Format specifies the format of sources.
	Format Format

	// BufferSize specifies a buffer size in the underlying device.
	//
	// If 0 is specified, the driver's default buffer size is used.
	// Set BufferSize to adjust the buffer size if you want to adjust latency or reduce noises.
	// Too big buffer size can increase the latency time.
	// On the other hand, too small buffer size can cause glitch noises due to buffer shortage.
	BufferSize time.Duration
}

// NewContext creates a new context with given options.
// A context creates and holds ready-to-use Player objects.
// NewContext returns a context, a channel that is closed when the context is ready, and an error if it exists.
//
// Creating multiple contexts is NOT supported.
func NewContext(options *NewContextOptions) (*Context, chan struct{}, error) {
	contextCreationMutex.Lock()
	defer contextCreationMutex.Unlock()

	if contextCreated {
		return nil, nil, fmt.Errorf("oto: context is already created")
	}
	contextCreated = true

	var bufferSizeInBytes int
	if options.BufferSize != 0 {
		// The underying driver always uses 32bit floats.
		bytesPerSample := options.ChannelCount * 4
		bytesPerSecond := options.SampleRate * bytesPerSample
		bufferSizeInBytes = int(int64(options.BufferSize) * int64(bytesPerSecond) / int64(time.Second))
		bufferSizeInBytes = bufferSizeInBytes / bytesPerSample * bytesPerSample
	}
	ctx, ready, err := newContext(options.SampleRate, options.ChannelCount, mux.Format(options.Format), bufferSizeInBytes)
	if err != nil {
		return nil, nil, err
	}
	return &Context{context: ctx}, ready, nil
}

// NewPlayer creates a new, ready-to-use Player belonging to the Context.
// It is safe to create multiple players.
//
// The format of r is as follows:
//
//	[data]      = [sample 1] [sample 2] [sample 3] ...
//	[sample *]  = [channel 1] [channel 2] ...
//	[channel *] = [byte 1] [byte 2] ...
//
// Byte ordering is little endian.
//
// A player has some amount of an underlying buffer.
// Read data from r is queued to the player's underlying buffer.
// The underlying buffer is consumed by its playing.
// Then, r's position and the current playing position don't necessarily match.
// If you want to clear the underlying buffer for some reasons e.g., you want to seek the position of r,
// call the player's Reset function.
//
// You cannot share r by multiple players.
//
// The returned player implements Player, BufferSizeSetter, and io.Seeker.
// You can modify the buffer size of a player by the SetBufferSize function.
// A small buffer size is useful if you want to play a real-time PCM for example.
// Note that the audio quality might be affected if you modify the buffer size.
//
// If r does not implement io.Seeker, the returned player's Seek returns an error.
//
// NewPlayer is concurrent-safe.
//
// All the functions of a Player returned by NewPlayer are concurrent-safe.
func (c *Context) NewPlayer(r io.Reader) *Player {
	return &Player{
		player: c.context.mux.NewPlayer(r),
	}
}

// Suspend suspends the entire audio play.
//
// Suspend is concurrent-safe.
func (c *Context) Suspend() error {
	return c.context.Suspend()
}

// Resume resumes the entire audio play, which was suspended by Suspend.
//
// Resume is concurrent-safe.
func (c *Context) Resume() error {
	return c.context.Resume()
}

// Err returns the current error.
//
// Err is concurrent-safe.
func (c *Context) Err() error {
	return c.context.Err()
}

type atomicError struct {
	err error
	m   sync.Mutex
}

func (a *atomicError) TryStore(err error) {
	a.m.Lock()
	defer a.m.Unlock()
	if a.err == nil {
		a.err = err
	}
}

func (a *atomicError) Load() error {
	a.m.Lock()
	defer a.m.Unlock()
	return a.err
}
