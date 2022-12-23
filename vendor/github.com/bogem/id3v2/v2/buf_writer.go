package id3v2

import (
	"bufio"
	"io"
)

// bufWriter is used for convenient writing of frames.
type bufWriter struct {
	err     error
	w       *bufio.Writer
	written int
}

func newBufWriter(w io.Writer) *bufWriter {
	return &bufWriter{w: bufio.NewWriter(w)}
}

func (bw *bufWriter) EncodeAndWriteText(src string, to Encoding) {
	if bw.err != nil {
		return
	}

	bw.err = encodeWriteText(bw, src, to)
}

func (bw *bufWriter) Flush() error {
	if bw.err != nil {
		return bw.err
	}
	return bw.w.Flush()
}

func (bw *bufWriter) Reset(w io.Writer) {
	bw.err = nil
	bw.written = 0
	bw.w.Reset(w)
}

func (bw *bufWriter) WriteByte(c byte) {
	if bw.err != nil {
		return
	}
	bw.err = bw.w.WriteByte(c)
	if bw.err == nil {
		bw.written++
	}
}

func (bw *bufWriter) WriteBytesSize(size uint, synchSafe bool) {
	if bw.err != nil {
		return
	}
	bw.err = writeBytesSize(bw, size, synchSafe)
}

func (bw *bufWriter) WriteString(s string) {
	if bw.err != nil {
		return
	}
	var n int
	n, bw.err = bw.w.WriteString(s)
	bw.written += n
}

func (bw *bufWriter) Write(p []byte) (n int, err error) {
	if bw.err != nil {
		return 0, bw.err
	}
	n, err = bw.w.Write(p)
	bw.written += n
	bw.err = err
	return n, err
}

func (bw *bufWriter) Written() int {
	return bw.written
}

func useBufWriter(w io.Writer, f func(*bufWriter)) (int64, error) {
	var writtenBefore int
	bw, ok := w.(*bufWriter)
	if ok {
		writtenBefore = bw.Written()
	} else {
		bw = getBufWriter(w)
		defer putBufWriter(bw)
	}

	f(bw)

	return int64(bw.Written() - writtenBefore), bw.Flush()
}
