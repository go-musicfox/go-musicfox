// Copyright 2019 The Oto Authors
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

package mux

import (
	"bufio"
	"io"
	"runtime"
	"sync"
)

// Mux is a multiplexer for multiple io.Reader objects.
type Mux struct {
	channelNum      int
	bitDepthInBytes int
	readers         map[io.Reader]*bufio.Reader
	closed          bool

	m sync.RWMutex
}

// New creates a new Mux with the specified number of channels and bit depth.
func New(channelNum, bitDepthInBytes int) *Mux {
	m := &Mux{
		channelNum:      channelNum,
		bitDepthInBytes: bitDepthInBytes,
		readers:         map[io.Reader]*bufio.Reader{},
	}
	runtime.SetFinalizer(m, (*Mux).Close)
	return m
}

// Read reads data from all of its readers, interprets it as samples with the bit depth
// specified during its creation, then adds all of the samples together and fills the buf
// slice with the result of this.
//
// If there are no readers, Read fills in some zeros to prevent a program from freezing.
func (m *Mux) Read(buf []byte) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.closed {
		return 0, io.EOF
	}

	if len(m.readers) == 0 {
		// When there is no reader, Read should return with 0s or Read caller can block forever.
		// See https://github.com/hajimehoshi/go-mp3/issues/28
		n := 256
		if len(buf) < 256 {
			n = len(buf)
		}

		switch m.bitDepthInBytes {
		case 1:
			const offset = 128
			for i := 0; i < n; i++ {
				buf[i] = offset
			}
			return n, nil
		case 2:
			for i := 0; i < n; i++ {
				buf[i] = 0
			}
			return n, nil
		default:
			panic("not reached")
		}
	}

	bs := m.channelNum * m.bitDepthInBytes
	l := len(buf)
	l = l / bs * bs // Adjust the length in order not to mix different channels.

	var bufs [][]byte
	for _, p := range m.readers {
		peeked, err := p.Peek(l)
		if err != nil && err != bufio.ErrBufferFull && err != io.EOF {
			return 0, err
		}
		if l > len(peeked) {
			l = len(peeked)
			l = l / bs * bs
		}
		bufs = append(bufs, peeked[:l])
	}

	if l == 0 {
		// Returning 0 without error can block the caller of Read forever. Call Gosched to encourage context switching.
		runtime.Gosched()
		return 0, nil
	}

	for _, p := range m.readers {
		if _, err := p.Discard(l); err != nil && err != io.EOF {
			return 0, err
		}
	}

	switch m.bitDepthInBytes {
	case 1:
		const (
			max    = 127
			min    = -128
			offset = 128
		)
		for i := 0; i < l; i++ {
			x := 0
			for _, b := range bufs {
				x += int(b[i]) - offset
			}
			if x > max {
				x = max
			}
			if x < min {
				x = min
			}
			buf[i] = byte(x + offset)
		}
	case 2:
		const (
			max = (1 << 15) - 1
			min = -(1 << 15)
		)
		for i := 0; i < l/2; i++ {
			x := 0
			for _, b := range bufs {
				x += int(int16(b[2*i]) | (int16(b[2*i+1]) << 8))
			}
			if x > max {
				x = max
			}
			if x < min {
				x = min
			}
			buf[2*i] = byte(x)
			buf[2*i+1] = byte(x >> 8)
		}
	default:
		panic("not reached")
	}

	return l, nil
}

// Close invalidates the Mux. It doesn't close its readers.
func (m *Mux) Close() error {
	m.m.Lock()
	runtime.SetFinalizer(m, nil)
	m.readers = nil
	m.closed = true
	m.m.Unlock()
	return nil
}

// AddSource adds a reader to the Mux.
func (m *Mux) AddSource(source io.Reader) {
	m.m.Lock()
	if m.closed {
		panic("mux: already closed")
	}
	if _, ok := m.readers[source]; ok {
		panic("mux: the io.Reader cannot be added multiple times")
	}
	m.readers[source] = bufio.NewReaderSize(source, 256)
	m.m.Unlock()
}

// RemoveSource removes a reader from the Mux.
func (m *Mux) RemoveSource(source io.Reader) {
	m.m.Lock()
	if m.closed {
		panic("mux: already closed")
	}
	if _, ok := m.readers[source]; !ok {
		panic("mux: the io.Reader is already removed")
	}
	delete(m.readers, source)
	m.m.Unlock()
}

// Sources returns all the registered readers.
func (m *Mux) Sources() []io.Reader {
	m.m.Lock()
	defer m.m.Unlock()

	var rs []io.Reader
	for r := range m.readers {
		rs = append(rs, r)
	}
	return rs
}
