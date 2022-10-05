/*

CountReader implementation.

*/

package bitio

import (
	"io"
)

// CountReader is an improved version of Reader that also keeps track
// of the number of processed bits. If you don't need the number
// of processed bits, use the faster Reader.
//
// For convenience, it also implements io.Reader and io.ByteReader.
type CountReader struct {
	*Reader
	BitsCount int64 // Total number of bits read
}

// NewCountReader returns a new CountReader using the specified io.Reader as
// the input (source).
func NewCountReader(in io.Reader) *CountReader {
	return &CountReader{NewReader(in), 0}
}

// Read reads up to len(p) bytes (8 * len(p) bits) from the underlying reader,
// and counts the number of bits read.
//
// Read implements io.Reader, and gives a byte-level view of the bit stream.
// This will give best performance if the underlying io.Reader is aligned
// to a byte boundary (else all the individual bytes are assembled from multiple bytes).
// Byte boundary can be ensured by calling Align().
func (r *CountReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.BitsCount += int64(n) * 8
	return
}

// ReadBits reads n bits and returns them as the lowest n bits of u.
func (r *CountReader) ReadBits(n uint8) (u uint64, err error) {
	u, err = r.Reader.ReadBits(n)
	if err == nil {
		r.BitsCount += int64(n)
	}
	return
}

// ReadByte reads the next 8 bits and returns them as a byte.
//
// ReadByte implements io.ByteReader.
func (r *CountReader) ReadByte() (b byte, err error) {
	b, err = r.Reader.ReadByte()
	if err == nil {
		r.BitsCount += 8
	}
	return
}

// ReadBool reads the next bit, and returns true if it is 1.
func (r *CountReader) ReadBool() (b bool, err error) {
	b, err = r.Reader.ReadBool()
	if err == nil {
		r.BitsCount += 1
	}
	return
}

// Align aligns the bit stream to a byte boundary,
// so next read will read/use data from the next byte.
// Returns the number of unread / skipped bits.
func (r *CountReader) Align() (skipped uint8) {
	skipped = r.Reader.Align()
	r.BitsCount += int64(skipped)
	return
}

// TryRead tries to read up to len(p) bytes (8 * len(p) bits) from the underlying reader.
//
// If there was a previous TryError, it does nothing. Else it calls Read(),
// returns the data it provides and stores the error in the TryError field.
func (r *CountReader) TryRead(p []byte) (n int) {
	if r.TryError == nil {
		n, r.TryError = r.Read(p)
	}
	return
}

// TryReadBits tries to read n bits.
//
// If there was a previous TryError, it does nothing. Else it calls ReadBits(),
// returns the data it provides and stores the error in the TryError field.
func (r *CountReader) TryReadBits(n uint8) (u uint64) {
	if r.TryError == nil {
		u, r.TryError = r.ReadBits(n)
	}
	return
}

// TryReadByte tries to read the next 8 bits and return them as a byte.
//
// If there was a previous TryError, it does nothing. Else it calls ReadByte(),
// returns the data it provides and stores the error in the TryError field.
func (r *CountReader) TryReadByte() (b byte) {
	if r.TryError == nil {
		b, r.TryError = r.ReadByte()
	}
	return
}

// TryReadBool tries to read the next bit, and return true if it is 1.
//
// If there was a previous TryError, it does nothing. Else it calls ReadBool(),
// returns the data it provides and stores the error in the TryError field.
func (r *CountReader) TryReadBool() (b bool) {
	if r.TryError == nil {
		b, r.TryError = r.ReadBool()
	}
	return
}
