package meta

import "io"

// readString reads and returns exactly n bytes from the provided io.Reader.
//
// The error is io.EOF only if no bytes were read. If an io.EOF happens after
// reading some but not all the bytes, ReadFull returns io.ErrUnexpectedEOF. On
// return, n == len(buf) if and only if err == nil.
func readString(r io.Reader, n int) (string, error) {
	// readBuf is the local buffer used by readBytes.
	var backingArray [4096]byte // hopefully allocated on stack.
	readBuf := backingArray[:]
	if n > len(readBuf) {
		// The local buffer is initially 4096 bytes and will grow automatically if
		// so required.
		readBuf = make([]byte, n)
	}
	_, err := io.ReadFull(r, readBuf[:n])
	if err != nil {
		return "", err
	}
	return string(readBuf[:n]), nil
}
