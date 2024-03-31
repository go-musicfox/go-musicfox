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

// Package mux offers APIs for a low-level multiplexer of audio players.
// Usually you don't have to use this.
package mux

import (
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"time"
)

// Format must sync with oto's Format.
type Format int

const (
	FormatFloat32LE Format = iota
	FormatUnsignedInt8
	FormatSignedInt16LE
)

func (f Format) ByteLength() int {
	switch f {
	case FormatFloat32LE:
		return 4
	case FormatUnsignedInt8:
		return 1
	case FormatSignedInt16LE:
		return 2
	}
	panic(fmt.Sprintf("mux: unexpected format: %d", f))
}

// Mux is a low-level multiplexer of audio players.
type Mux struct {
	sampleRate   int
	channelCount int
	format       Format

	players map[*playerImpl]struct{}
	buf     []float32
	cond    *sync.Cond
}

// New creates a new Mux.
func New(sampleRate int, channelCount int, format Format) *Mux {
	m := &Mux{
		sampleRate:   sampleRate,
		channelCount: channelCount,
		format:       format,
		cond:         sync.NewCond(&sync.Mutex{}),
	}
	go m.loop()
	return m
}

func (m *Mux) shouldWait() bool {
	for p := range m.players {
		if p.canReadSourceToBuffer() {
			return false
		}
	}
	return true
}

func (m *Mux) wait() {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()

	for m.shouldWait() {
		m.cond.Wait()
	}
}

func (m *Mux) loop() {
	var players []*playerImpl
	for {
		m.wait()

		m.cond.L.Lock()
		for i := range players {
			players[i] = nil
		}
		players = players[:0]
		for p := range m.players {
			players = append(players, p)
		}
		m.cond.L.Unlock()

		allZero := true
		for _, p := range players {
			n := p.readSourceToBuffer()
			if n != 0 {
				allZero = false
			}
		}

		// Sleeping is necessary especially on browsers.
		// Sometimes a player continues to read 0 bytes from the source and this loop can be a busy loop in such case.
		if allZero {
			time.Sleep(time.Millisecond)
		}
	}
}

func (m *Mux) addPlayer(player *playerImpl) {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()

	if m.players == nil {
		m.players = map[*playerImpl]struct{}{}
	}
	m.players[player] = struct{}{}
	m.cond.Signal()
}

func (m *Mux) removePlayer(player *playerImpl) {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()

	delete(m.players, player)
	m.cond.Signal()
}

// ReadFloat32s fills buf with the multiplexed data of the players as float32 values.
func (m *Mux) ReadFloat32s(buf []float32) {
	m.cond.L.Lock()
	players := make([]*playerImpl, 0, len(m.players))
	for p := range m.players {
		players = append(players, p)
	}
	m.cond.L.Unlock()

	for i := range buf {
		buf[i] = 0
	}
	for _, p := range players {
		p.readBufferAndAdd(buf)
	}
	m.cond.Signal()
}

type Player struct {
	p *playerImpl
}

type playerState int

const (
	playerPaused playerState = iota
	playerPlay
	playerClosed
)

type playerImpl struct {
	mux        *Mux
	src        io.Reader
	prevVolume float64
	volume     float64
	err        error
	state      playerState
	tmpbuf     []byte
	buf        []byte
	eof        bool
	bufferSize int

	m sync.Mutex
}

func (m *Mux) NewPlayer(src io.Reader) *Player {
	pl := &Player{
		p: &playerImpl{
			mux:        m,
			src:        src,
			prevVolume: 1,
			volume:     1,
			bufferSize: m.defaultBufferSize(),
		},
	}
	runtime.SetFinalizer(pl, (*Player).Close)
	return pl
}

func (p *Player) Err() error {
	return p.p.Err()
}

func (p *playerImpl) Err() error {
	p.m.Lock()
	defer p.m.Unlock()

	return p.err
}

func (p *Player) Play() {
	p.p.Play()
}

func (p *playerImpl) Play() {
	// Goroutines don't work effiently on Windows. Avoid using them (hajimehoshi/ebiten#1768).
	if runtime.GOOS == "windows" {
		p.m.Lock()
		defer p.m.Unlock()

		p.playImpl()
	} else {
		ch := make(chan struct{})
		go func() {
			p.m.Lock()
			defer p.m.Unlock()

			close(ch)
			p.playImpl()
		}()
		<-ch
	}
}

func (p *Player) SetBufferSize(bufferSize int) {
	p.p.setBufferSize(bufferSize)
}

func (p *playerImpl) setBufferSize(bufferSize int) {
	p.m.Lock()
	defer p.m.Unlock()

	orig := p.bufferSize
	p.bufferSize = bufferSize
	if bufferSize == 0 {
		p.bufferSize = p.mux.defaultBufferSize()
	}
	if orig != p.bufferSize {
		p.tmpbuf = nil
	}
}

func (p *playerImpl) ensureTmpBuf() []byte {
	if p.tmpbuf == nil {
		p.tmpbuf = make([]byte, p.bufferSize)
	}
	return p.tmpbuf
}

// read reads the source to buf.
// read unlocks the mutex temporarily and locks when reading finishes.
// This avoids locking during an external function call Read (#188).
//
// When read is called, the mutex m must be locked.
func (p *playerImpl) read(buf []byte) (int, error) {
	p.m.Unlock()
	defer p.m.Lock()
	return p.src.Read(buf)
}

// addToPlayers adds p to the players set.
//
// When addToPlayers is called, the mutex m must be locked.
func (p *playerImpl) addToPlayers() {
	p.m.Unlock()
	defer p.m.Lock()
	p.mux.addPlayer(p)
}

// removeFromPlayers removes p from the players set.
//
// When removeFromPlayers is called, the mutex m must be locked.
func (p *playerImpl) removeFromPlayers() {
	p.m.Unlock()
	defer p.m.Lock()
	p.mux.removePlayer(p)
}

func (p *playerImpl) playImpl() {
	if p.err != nil {
		return
	}
	if p.state != playerPaused {
		return
	}
	p.state = playerPlay

	if !p.eof {
		buf := p.ensureTmpBuf()
		for len(p.buf) < p.bufferSize {
			n, err := p.read(buf)
			if err != nil && err != io.EOF {
				p.setErrorImpl(err)
				return
			}
			p.buf = append(p.buf, buf[:n]...)
			if err == io.EOF {
				p.eof = true
				break
			}
		}
	}

	if p.eof && len(p.buf) == 0 {
		p.state = playerPaused
	}

	p.addToPlayers()
}

func (p *Player) Pause() {
	p.p.Pause()
}

func (p *playerImpl) Pause() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.state != playerPlay {
		return
	}
	p.state = playerPaused
}

func (p *Player) Seek(offset int64, whence int) (int64, error) {
	return p.p.Seek(offset, whence)
}

func (p *playerImpl) Seek(offset int64, whence int) (int64, error) {
	p.m.Lock()
	defer p.m.Unlock()

	// If a player is playing, keep playing even after this seeking.
	if p.state == playerPlay {
		defer p.playImpl()
	}

	// Reset the internal buffer.
	p.resetImpl()

	// Check if the source implements io.Seeker.
	s, ok := p.src.(io.Seeker)
	if !ok {
		return 0, errors.New("mux: the source must implement io.Seeker")
	}
	return s.Seek(offset, whence)
}

func (p *Player) Reset() {
	p.p.Reset()
}

func (p *playerImpl) Reset() {
	p.m.Lock()
	defer p.m.Unlock()
	p.resetImpl()
}

func (p *playerImpl) resetImpl() {
	if p.state == playerClosed {
		return
	}
	p.state = playerPaused
	p.buf = p.buf[:0]
	p.eof = false
}

func (p *Player) IsPlaying() bool {
	return p.p.IsPlaying()
}

func (p *playerImpl) IsPlaying() bool {
	p.m.Lock()
	defer p.m.Unlock()
	return p.state == playerPlay
}

func (p *Player) Volume() float64 {
	return p.p.Volume()
}

func (p *playerImpl) Volume() float64 {
	p.m.Lock()
	defer p.m.Unlock()
	return p.volume
}

func (p *Player) SetVolume(volume float64) {
	p.p.SetVolume(volume)
}

func (p *playerImpl) SetVolume(volume float64) {
	p.m.Lock()
	defer p.m.Unlock()
	p.volume = volume
	if p.state != playerPlay {
		p.prevVolume = volume
	}
}

func (p *Player) BufferedSize() int {
	return p.p.BufferedSize()
}

func (p *playerImpl) BufferedSize() int {
	p.m.Lock()
	defer p.m.Unlock()
	return len(p.buf)
}

func (p *Player) Close() error {
	runtime.SetFinalizer(p, nil)
	return p.p.Close()
}

func (p *playerImpl) Close() error {
	p.m.Lock()
	defer p.m.Unlock()
	return p.closeImpl()
}

func (p *playerImpl) closeImpl() error {
	p.removeFromPlayers()

	if p.state == playerClosed {
		return p.err
	}
	p.state = playerClosed
	p.buf = nil
	return p.err
}

func (p *playerImpl) readBufferAndAdd(buf []float32) int {
	p.m.Lock()
	defer p.m.Unlock()

	if p.state != playerPlay {
		return 0
	}

	format := p.mux.format
	bitDepthInBytes := format.ByteLength()
	n := len(p.buf) / bitDepthInBytes
	if n > len(buf) {
		n = len(buf)
	}

	prevVolume := float32(p.prevVolume)
	volume := float32(p.volume)

	channelCount := p.mux.channelCount
	rateDenom := float32(n / channelCount)

	src := p.buf[:n*bitDepthInBytes]

	for i := 0; i < n; i++ {
		var v float32
		switch format {
		case FormatFloat32LE:
			v = math.Float32frombits(uint32(src[4*i]) | uint32(src[4*i+1])<<8 | uint32(src[4*i+2])<<16 | uint32(src[4*i+3])<<24)
		case FormatUnsignedInt8:
			v8 := src[i]
			v = float32(v8-(1<<7)) / (1 << 7)
		case FormatSignedInt16LE:
			v16 := int16(src[2*i]) | (int16(src[2*i+1]) << 8)
			v = float32(v16) / (1 << 15)
		default:
			panic(fmt.Sprintf("mux: unexpected format: %d", format))
		}
		if volume == prevVolume {
			buf[i] += v * volume
		} else {
			rate := float32(i/channelCount) / rateDenom
			if rate > 1 {
				rate = 1
			}
			buf[i] += v * (volume*rate + prevVolume*(1-rate))
		}
	}

	p.prevVolume = p.volume

	copy(p.buf, p.buf[n*bitDepthInBytes:])
	p.buf = p.buf[:len(p.buf)-n*bitDepthInBytes]

	if p.eof && len(p.buf) == 0 {
		p.state = playerPaused
	}

	return n
}

func (p *playerImpl) canReadSourceToBuffer() bool {
	p.m.Lock()
	defer p.m.Unlock()

	if p.eof {
		return false
	}
	return len(p.buf) < p.bufferSize
}

func (p *playerImpl) readSourceToBuffer() int {
	p.m.Lock()
	defer p.m.Unlock()

	if p.err != nil {
		return 0
	}
	if p.state == playerClosed {
		return 0
	}

	if len(p.buf) >= p.bufferSize {
		return 0
	}

	buf := p.ensureTmpBuf()
	n, err := p.read(buf)

	if err != nil && err != io.EOF {
		p.setErrorImpl(err)
		return 0
	}

	p.buf = append(p.buf, buf[:n]...)
	if err == io.EOF {
		p.eof = true
		if len(p.buf) == 0 {
			p.state = playerPaused
		}
	}
	return n
}

func (p *playerImpl) setErrorImpl(err error) {
	p.err = err
	p.closeImpl()
}

// TODO: The term 'buffer' is confusing. Name each buffer with good terms.

// defaultBufferSize returns the default size of the buffer for the audio source.
// This buffer is used when unreading on pausing the player.
func (m *Mux) defaultBufferSize() int {
	bytesPerSample := m.channelCount * m.format.ByteLength()
	s := m.sampleRate * bytesPerSample / 2 // 0.5[s]
	// Align s in multiples of bytes per sample, or a buffer could have extra bytes.
	return s / bytesPerSample * bytesPerSample
}
