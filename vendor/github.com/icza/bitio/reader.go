/*

Reader implementation.

*/

package bitio

import (
	"bufio"
	"io"
)

// An io.Reader and io.ByteReader at the same time.
type readerAndByteReader interface {
	io.Reader
	io.ByteReader
}

// Reader is the bit reader implementation.
//
// If you need the number of processed bits, use CountReader.
//
// For convenience, it also implements io.Reader and io.ByteReader.
type Reader struct {
	in    readerAndByteReader
	cache byte // unread bits are stored here
	bits  byte // number of unread bits in cache

	// TryError holds the first error occurred in TryXXX() methods.
	TryError error
}

// NewReader returns a new Reader using the specified io.Reader as the input (source).
func NewReader(in io.Reader) *Reader {
	bin, ok := in.(readerAndByteReader)
	if !ok {
		bin = bufio.NewReader(in)
	}
	return &Reader{in: bin}
}

// Read reads up to len(p) bytes (8 * len(p) bits) from the underlying reader.
//
// Read implements io.Reader, and gives a byte-level view of the bit stream.
// This will give best performance if the underlying io.Reader is aligned
// to a byte boundary (else all the individual bytes are assembled from multiple bytes).
// Byte boundary can be ensured by calling Align().
func (r *Reader) Read(p []byte) (n int, err error) {
	// r.bits will be the same after reading 8 bits, so we don't need to update that.
	if r.bits == 0 {
		return r.in.Read(p)
	}

	for ; n < len(p); n++ {
		if p[n], err = r.readUnalignedByte(); err != nil {
			return
		}
	}

	return
}

// ReadBits reads n bits and returns them as the lowest n bits of u.
func (r *Reader) ReadBits(n uint8) (u uint64, err error) {
	// Some optimization, frequent cases
	if n < r.bits {
		// cache has all needed bits, and there are some extra which will be left in cache
		shift := r.bits - n
		u = uint64(r.cache >> shift)
		r.cache &= 1<<shift - 1
		r.bits = shift
		return
	}

	if n > r.bits {
		// all cache bits needed, and it's not even enough so more will be read
		if r.bits > 0 {
			u = uint64(r.cache)
			n -= r.bits
		}
		// Read whole bytes
		for n >= 8 {
			b, err2 := r.in.ReadByte()
			if err2 != nil {
				return 0, err2
			}
			u = u<<8 + uint64(b)
			n -= 8
		}
		// Read last fraction, if any
		if n > 0 {
			if r.cache, err = r.in.ReadByte(); err != nil {
				return 0, err
			}
			shift := 8 - n
			u = u<<n + uint64(r.cache>>shift)
			r.cache &= 1<<shift - 1
			r.bits = shift
		} else {
			r.bits = 0
		}
		return u, nil
	}

	// cache has exactly as many as needed
	r.bits = 0 // no need to clear cache, will be overwritten on next read
	return uint64(r.cache), nil
}

// ReadByte reads the next 8 bits and returns them as a byte.
//
// ReadByte implements io.ByteReader.
func (r *Reader) ReadByte() (b byte, err error) {
	// r.bits will be the same after reading 8 bits, so we don't need to update that.
	if r.bits == 0 {
		return r.in.ReadByte()
	}
	return r.readUnalignedByte()
}

// readUnalignedByte reads the next 8 bits which are (may be) unaligned and returns them as a byte.
func (r *Reader) readUnalignedByte() (b byte, err error) {
	// r.bits will be the same after reading 8 bits, so we don't need to update that.
	bits := r.bits
	b = r.cache << (8 - bits)
	r.cache, err = r.in.ReadByte()
	if err != nil {
		return 0, err
	}
	b |= r.cache >> bits
	r.cache &= 1<<bits - 1
	return
}

// ReadBool reads the next bit, and returns true if it is 1.
func (r *Reader) ReadBool() (b bool, err error) {
	if r.bits == 0 {
		r.cache, err = r.in.ReadByte()
		if err != nil {
			return
		}
		b = (r.cache & 0x80) != 0
		r.cache, r.bits = r.cache&0x7f, 7
		return
	}

	r.bits--
	b = (r.cache & (1 << r.bits)) != 0
	r.cache &= 1<<r.bits - 1
	return
}

// Align aligns the bit stream to a byte boundary,
// so next read will read/use data from the next byte.
// Returns the number of unread / skipped bits.
func (r *Reader) Align() (skipped uint8) {
	skipped = r.bits
	r.bits = 0 // no need to clear cache, will be overwritten on next read
	return
}

// TryRead tries to read up to len(p) bytes (8 * len(p) bits) from the underlying reader.
//
// If there was a previous TryError, it does nothing. Else it calls Read(),
// returns the data it provides and stores the error in the TryError field.
func (r *Reader) TryRead(p []byte) (n int) {
	if r.TryError == nil {
		n, r.TryError = r.Read(p)
	}
	return
}

// TryReadBits tries to read n bits.
//
// If there was a previous TryError, it does nothing. Else it calls ReadBits(),
// returns the data it provides and stores the error in the TryError field.
func (r *Reader) TryReadBits(n uint8) (u uint64) {
	if r.TryError == nil {
		u, r.TryError = r.ReadBits(n)
	}
	return
}

// TryReadByte tries to read the next 8 bits and return them as a byte.
//
// If there was a previous TryError, it does nothing. Else it calls ReadByte(),
// returns the data it provides and stores the error in the TryError field.
func (r *Reader) TryReadByte() (b byte) {
	if r.TryError == nil {
		b, r.TryError = r.ReadByte()
	}
	return
}

// TryReadBool tries to read the next bit, and return true if it is 1.
//
// If there was a previous TryError, it does nothing. Else it calls ReadBool(),
// returns the data it provides and stores the error in the TryError field.
func (r *Reader) TryReadBool() (b bool) {
	if r.TryError == nil {
		b, r.TryError = r.ReadBool()
	}
	return
}
