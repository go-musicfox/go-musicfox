package meta

import "io"

// readBuf is the local buffer used by readBytes.
var readBuf = make([]byte, 4096)

// readBytes reads and returns exactly n bytes from the provided io.Reader. The
// local buffer is reused in between calls to reduce garbage generation. It is
// the callers responsibility to make a copy of the returned data. The error is
// io.EOF only if no bytes were read. If an io.EOF happens after reading some
// but not all the bytes, ReadFull returns io.ErrUnexpectedEOF. On return, n ==
// len(buf) if and only if err == nil.
//
// The local buffer is initially 4096 bytes and will grow automatically if so
// required.
func readBytes(r io.Reader, n int) ([]byte, error) {
	if n > len(readBuf) {
		readBuf = make([]byte, n)
	}
	_, err := io.ReadFull(r, readBuf[:n])
	if err != nil {
		return nil, err
	}
	return readBuf[:n:n], nil
}
