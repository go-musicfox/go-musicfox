// Copyright 2017 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"bytes"
	"io"
	"sync"
)

// bsPool is a pool of byte slices.
var bsPool = sync.Pool{
	New: func() interface{} { return nil },
}

// getByteSlice returns []byte with len == size.
func getByteSlice(size int) []byte {
	fromPool := bsPool.Get()
	if fromPool == nil {
		return make([]byte, size)
	}
	bs := fromPool.([]byte)
	if cap(bs) < size {
		bs = make([]byte, size)
	}
	return bs[0:size]
}

// putByteSlice puts b to pool.
func putByteSlice(b []byte) {
	bsPool.Put(b)
}

var bwPool = sync.Pool{
	New: func() interface{} { return newBufWriter(nil) },
}

func getBufWriter(w io.Writer) *bufWriter {
	bw := bwPool.Get().(*bufWriter)
	bw.Reset(w)
	return bw
}

func putBufWriter(bw *bufWriter) {
	bwPool.Put(bw)
}

var lrPool = sync.Pool{
	New: func() interface{} { return new(io.LimitedReader) },
}

func getLimitedReader(rd io.Reader, n int64) *io.LimitedReader {
	r := lrPool.Get().(*io.LimitedReader)
	r.R = rd
	r.N = n
	return r
}

func putLimitedReader(r *io.LimitedReader) {
	r.N = 0
	r.R = nil
	lrPool.Put(r)
}

var rdPool = sync.Pool{
	New: func() interface{} { return newBufReader(nil) },
}

func getBufReader(rd io.Reader) *bufReader {
	reader := rdPool.Get().(*bufReader)
	reader.Reset(rd)
	return reader
}

func putBufReader(rd *bufReader) {
	rdPool.Put(rd)
}

var bbPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func getBytesBuffer() *bytes.Buffer {
	return bbPool.Get().(*bytes.Buffer)
}

func putBytesBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bbPool.Put(buf)
}
