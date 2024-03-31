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
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/ebitengine/oto/v3/internal/mux"
)

// Avoid goroutines on Windows (hajimehoshi/ebiten#1768).
// Apparently, switching contexts might take longer than other platforms.

const defaultHeaderBufferSize = 4096

type header struct {
	waveOut uintptr
	buffer  []float32
	waveHdr *_WAVEHDR
}

func newHeader(waveOut uintptr, bufferSizeInBytes int) (*header, error) {
	h := &header{
		waveOut: waveOut,
		buffer:  make([]float32, bufferSizeInBytes/4),
	}
	h.waveHdr = &_WAVEHDR{
		lpData:         uintptr(unsafe.Pointer(&h.buffer[0])),
		dwBufferLength: uint32(bufferSizeInBytes),
	}
	if err := waveOutPrepareHeader(waveOut, h.waveHdr); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *header) Write(data []float32) error {
	copy(h.buffer, data)
	if err := waveOutWrite(h.waveOut, h.waveHdr); err != nil {
		return err
	}
	return nil
}

func (h *header) IsQueued() bool {
	return h.waveHdr.dwFlags&_WHDR_INQUEUE != 0
}

func (h *header) Close() error {
	return waveOutUnprepareHeader(h.waveOut, h.waveHdr)
}

type winmmContext struct {
	sampleRate        int
	channelCount      int
	bufferSizeInBytes int

	waveOut uintptr
	headers []*header

	buf32 []float32

	mux       *mux.Mux
	err       atomicError
	loopEndCh chan error

	cond *sync.Cond

	suspended     bool
	suspendedCond *sync.Cond
}

var theWinMMContext *winmmContext

func newWinMMContext(sampleRate, channelCount int, mux *mux.Mux, bufferSizeInBytes int) (*winmmContext, error) {
	// winmm.dll is not available on Xbox.
	if err := winmm.Load(); err != nil {
		return nil, fmt.Errorf("oto: loading winmm.dll failed: %w", err)
	}

	c := &winmmContext{
		sampleRate:        sampleRate,
		channelCount:      channelCount,
		bufferSizeInBytes: bufferSizeInBytes,
		mux:               mux,
		cond:              sync.NewCond(&sync.Mutex{}),
		suspendedCond:     sync.NewCond(&sync.Mutex{}),
	}
	theWinMMContext = c

	if err := c.start(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *winmmContext) start() error {
	const bitsPerSample = 32
	nBlockAlign := c.channelCount * bitsPerSample / 8
	f := &_WAVEFORMATEX{
		wFormatTag:      _WAVE_FORMAT_IEEE_FLOAT,
		nChannels:       uint16(c.channelCount),
		nSamplesPerSec:  uint32(c.sampleRate),
		nAvgBytesPerSec: uint32(c.sampleRate * nBlockAlign),
		nBlockAlign:     uint16(nBlockAlign),
		wBitsPerSample:  bitsPerSample,
	}

	// TOOD: What about using an event instead of a callback? PortAudio and other libraries do that.
	w, err := waveOutOpen(f, waveOutOpenCallback)
	if errors.Is(err, windows.ERROR_NOT_FOUND) {
		// This can happen when no device is found (#77).
		return errDeviceNotFound
	}
	if errors.Is(err, _MMSYSERR_BADDEVICEID) {
		// This can happen when no device is found (hajimehoshi/ebiten#2316).
		return errDeviceNotFound
	}
	if err != nil {
		return err
	}

	headerBufferSize := defaultHeaderBufferSize
	if c.bufferSizeInBytes != 0 {
		headerBufferSize = c.bufferSizeInBytes
	}

	c.waveOut = w
	c.headers = make([]*header, 0, 6)
	for len(c.headers) < cap(c.headers) {
		h, err := newHeader(c.waveOut, headerBufferSize)
		if err != nil {
			return err
		}
		c.headers = append(c.headers, h)
	}

	c.buf32 = make([]float32, headerBufferSize/4)
	go c.loop()

	return nil
}

func (c *winmmContext) Suspend() error {
	c.suspendedCond.L.Lock()
	c.suspended = true
	c.suspendedCond.L.Unlock()
	c.suspendedCond.Signal()

	return nil
}

func (c *winmmContext) Resume() (ferr error) {
	c.suspendedCond.L.Lock()
	c.suspended = false
	c.suspendedCond.L.Unlock()
	c.suspendedCond.Signal()

	return nil
}

func (c *winmmContext) Err() error {
	if err := c.err.Load(); err != nil {
		return err.(error)
	}
	return nil
}

func (c *winmmContext) isHeaderAvailable() bool {
	for _, h := range c.headers {
		if !h.IsQueued() {
			return true
		}
	}
	return false
}

var waveOutOpenCallback = windows.NewCallback(func(hwo, uMsg, dwInstance, dwParam1, dwParam2 uintptr) uintptr {
	// Queuing a header in this callback might not work especially when a headset is connected or disconnected.
	// Just signal the condition vairable and don't do other things.
	const womDone = 0x3bd
	if uMsg != womDone {
		return 0
	}
	theWinMMContext.cond.Signal()
	return 0
})

func (c *winmmContext) waitUntilHeaderAvailable() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for !c.isHeaderAvailable() && c.err.Load() == nil && c.loopEndCh == nil {
		c.cond.Wait()
	}
	return c.err.Load() == nil && c.loopEndCh == nil
}

func (c *winmmContext) loop() {
	defer func() {
		if err := c.closeLoop(); err != nil {
			c.err.TryStore(err)
		}
	}()
	for {
		c.suspendedCond.L.Lock()
		for c.suspended {
			c.suspendedCond.Wait()
		}
		c.suspendedCond.L.Unlock()

		if !c.waitUntilHeaderAvailable() {
			return
		}
		c.appendBuffers()
	}
}

func (c *winmmContext) closeLoop() (ferr error) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	defer func() {
		if c.loopEndCh != nil {
			if ferr != nil {
				c.loopEndCh <- ferr
				ferr = nil
			}
			close(c.loopEndCh)
			c.loopEndCh = nil
		}
	}()

	for _, h := range c.headers {
		if err := h.Close(); err != nil {
			return err
		}
	}
	c.headers = nil

	if err := waveOutClose(c.waveOut); err != nil {
		return err
	}
	c.waveOut = 0
	return nil
}

func (c *winmmContext) appendBuffers() {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	if c.err.Load() != nil {
		return
	}

	c.mux.ReadFloat32s(c.buf32)

	for _, h := range c.headers {
		if h.IsQueued() {
			continue
		}

		if err := h.Write(c.buf32); err != nil {
			switch {
			case errors.Is(err, _MMSYSERR_NOMEM):
				continue
			case errors.Is(err, _MMSYSERR_NODRIVER):
				sleep := time.Duration(float64(time.Second) * float64(len(c.buf32)) / float64(c.channelCount) / float64(c.sampleRate))
				time.Sleep(sleep)
				return
			case errors.Is(err, windows.ERROR_NOT_FOUND):
				// This error can happen when e.g. a new HDMI connection is detected (#51).
				// TODO: Retry later.
			}
			c.err.TryStore(fmt.Errorf("oto: Queueing the header failed: %v", err))
		}
		return
	}
}
