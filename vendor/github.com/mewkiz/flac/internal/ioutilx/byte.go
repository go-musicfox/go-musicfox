// Package ioutilx implements extended input/output utility functions.
package ioutilx

import (
	"io"
)

// ReadByte reads and returns the next byte from r.
func ReadByte(r io.Reader) (byte, error) {
	var buf [1]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

// WriteByte writes the given byte to w.
func WriteByte(w io.Writer, b byte) error {
	buf := [1]byte{b}
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}
	return nil
}
