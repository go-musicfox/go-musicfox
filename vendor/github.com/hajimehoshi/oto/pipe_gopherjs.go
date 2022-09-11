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

//go:build js && !wasm
// +build js,!wasm

package oto

import (
	"io"
)

const pipeBufSize = 4096

// pipe returns a set of an io.ReadCloser and an io.WriteCloser.
//
// This is basically same as io.Pipe, but is implemented in more effient way under the assumption that
// this works on a single thread environment so that locks are not required.
func pipe() (io.ReadCloser, io.WriteCloser) {
	w := &pipeWriter{
		consumed: make(chan struct{}),
		provided: make(chan struct{}),
		closed:   make(chan struct{}),
	}
	r := &pipeReader{
		w:      w,
		closed: make(chan struct{}),
	}
	w.r = r
	return r, w
}

type pipeReader struct {
	w      *pipeWriter
	closed chan struct{}
}

func (r *pipeReader) Read(buf []byte) (int, error) {
	// If this returns 0 with no errors, the caller might block forever on browsers.
	// For example, bufio.Reader tries to Read until any byte can be read, but context switch never happens on browsers.
	for len(r.w.buf) == 0 {
		select {
		case <-r.w.provided:
		case <-r.w.closed:
			if len(r.w.buf) == 0 {
				return 0, io.EOF
			}
		case <-r.closed:
			return 0, io.ErrClosedPipe
		}
	}
	notify := len(r.w.buf) >= pipeBufSize && len(buf) > 0
	n := copy(buf, r.w.buf)
	r.w.buf = r.w.buf[n:]
	if notify {
		go func() {
			r.w.consumed <- struct{}{}
		}()
	}
	return n, nil
}

func (r *pipeReader) Close() error {
	close(r.closed)
	return nil
}

type pipeWriter struct {
	r        *pipeReader
	buf      []byte
	closed   chan struct{}
	consumed chan struct{}
	provided chan struct{}
}

func (w *pipeWriter) Write(buf []byte) (int, error) {
	for len(w.buf) >= pipeBufSize {
		select {
		case <-w.consumed:
		case <-w.r.closed:
			return 0, io.ErrClosedPipe
		case <-w.closed:
			return 0, io.ErrClosedPipe
		}
	}
	notify := len(w.buf) == 0 && len(buf) > 0
	w.buf = append(w.buf, buf...)
	if notify {
		w.provided <- struct{}{}
	}
	return len(buf), nil
}

func (w *pipeWriter) Close() error {
	close(w.closed)
	return nil
}
