// Package utf8 implements encoding and decoding of UTF-8 coded numbers.
package utf8

import (
	"errors"
	"fmt"
	"io"

	"github.com/mewkiz/flac/internal/ioutilx"
)

const (
	tx = 0x80 // 1000 0000
	t2 = 0xC0 // 1100 0000
	t3 = 0xE0 // 1110 0000
	t4 = 0xF0 // 1111 0000
	t5 = 0xF8 // 1111 1000
	t6 = 0xFC // 1111 1100
	t7 = 0xFE // 1111 1110
	t8 = 0xFF // 1111 1111

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111
	mask5 = 0x03 // 0000 0011
	mask6 = 0x01 // 0000 0001

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1
	rune4Max = 1<<21 - 1
	rune5Max = 1<<26 - 1
	rune6Max = 1<<31 - 1
	rune7Max = 1<<36 - 1
)

// Decode decodes a "UTF-8" coded number and returns it.
//
// ref: http://permalink.gmane.org/gmane.comp.audio.compression.flac.devel/3033
//
// Algorithm description:
//    - read one byte B0 from the stream
//    - if B0 = 0xxxxxxx then the read value is B0 -> end
//    - if B0 = 10xxxxxx, the encoding is invalid
//    - if B0 = 11xxxxxx, set L to the number of leading binary 1s minus 1:
//         B0 = 110xxxxx -> L = 1
//         B0 = 1110xxxx -> L = 2
//         B0 = 11110xxx -> L = 3
//         B0 = 111110xx -> L = 4
//         B0 = 1111110x -> L = 5
//         B0 = 11111110 -> L = 6
//    - assign the bits following the encoding (the x bits in the examples) to
//      a variable R with a magnitude of at least 36 bits
//    - loop from 1 to L
//         - left shift R 6 bits
//         - read B from the stream
//         - if B does not match 10xxxxxx, the encoding is invalid
//         - set R = R or <the lower 6 bits from B>
//    - the read value is R
func Decode(r io.Reader) (x uint64, err error) {
	c0, err := ioutilx.ReadByte(r)
	if err != nil {
		return 0, err
	}

	// 1-byte, 7-bit sequence?
	if c0 < tx {
		// if c0 == 0xxxxxxx
		// total: 7 bits (7)
		return uint64(c0), nil
	}

	// unexpected continuation byte?
	if c0 < t2 {
		// if c0 == 10xxxxxx
		return 0, errors.New("frame.decodeUTF8Int: unexpected continuation byte")
	}

	// get number of continuation bytes and store bits from c0.
	var l int
	switch {
	case c0 < t3:
		// if c0 == 110xxxxx
		// total: 11 bits (5 + 6)
		l = 1
		x = uint64(c0 & mask2)
	case c0 < t4:
		// if c0 == 1110xxxx
		// total: 16 bits (4 + 6 + 6)
		l = 2
		x = uint64(c0 & mask3)
	case c0 < t5:
		// if c0 == 11110xxx
		// total: 21 bits (3 + 6 + 6 + 6)
		l = 3
		x = uint64(c0 & mask4)
	case c0 < t6:
		// if c0 == 111110xx
		// total: 26 bits (2 + 6 + 6 + 6 + 6)
		l = 4
		x = uint64(c0 & mask5)
	case c0 < t7:
		// if c0 == 1111110x
		// total: 31 bits (1 + 6 + 6 + 6 + 6 + 6)
		l = 5
		x = uint64(c0 & mask6)
	case c0 < t8:
		// if c0 == 11111110
		// total: 36 bits (0 + 6 + 6 + 6 + 6 + 6 + 6)
		l = 6
		x = 0
	}

	// store bits from continuation bytes.
	for i := 0; i < l; i++ {
		x <<= 6
		c, err := ioutilx.ReadByte(r)
		if err != nil {
			if err == io.EOF {
				return 0, io.ErrUnexpectedEOF
			}
			return 0, err
		}
		if c < tx || t2 <= c {
			// if c != 10xxxxxx
			return 0, errors.New("frame.decodeUTF8Int: expected continuation byte")
		}
		x |= uint64(c & maskx)
	}

	// check if number representation is larger than necessary.
	switch l {
	case 1:
		if x <= rune1Max {
			return 0, fmt.Errorf("frame.decodeUTF8Int: larger number representation than necessary; x (%d) stored in %d bytes, could be stored in %d bytes", x, l+1, l)
		}
	case 2:
		if x <= rune2Max {
			return 0, fmt.Errorf("frame.decodeUTF8Int: larger number representation than necessary; x (%d) stored in %d bytes, could be stored in %d bytes", x, l+1, l)
		}
	case 3:
		if x <= rune3Max {
			return 0, fmt.Errorf("frame.decodeUTF8Int: larger number representation than necessary; x (%d) stored in %d bytes, could be stored in %d bytes", x, l+1, l)
		}
	case 4:
		if x <= rune4Max {
			return 0, fmt.Errorf("frame.decodeUTF8Int: larger number representation than necessary; x (%d) stored in %d bytes, could be stored in %d bytes", x, l+1, l)
		}
	case 5:
		if x <= rune5Max {
			return 0, fmt.Errorf("frame.decodeUTF8Int: larger number representation than necessary; x (%d) stored in %d bytes, could be stored in %d bytes", x, l+1, l)
		}
	case 6:
		if x <= rune6Max {
			return 0, fmt.Errorf("frame.decodeUTF8Int: larger number representation than necessary; x (%d) stored in %d bytes, could be stored in %d bytes", x, l+1, l)
		}
	}
	return x, nil
}
