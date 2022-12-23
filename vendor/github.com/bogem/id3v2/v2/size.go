// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import "errors"

const (
	// id3SizeLen is length of ID3v2 size format (4 * 0bxxxxxxxx).
	id3SizeLen = 4

	synchSafeMaxSize  = 268435455            // == 0b00001111 11111111 11111111 11111111
	synchSafeSizeBase = 7                    // == 0b01111111
	synchSafeMask     = uint(254 << (3 * 8)) // 11111110 000000000 000000000 000000000

	synchUnsafeMaxSize  = 4294967295           // == 0b11111111 11111111 11111111 11111111
	synchUnsafeSizeBase = 8                    // == 0b11111111
	synchUnsafeMask     = uint(255 << (3 * 8)) // 11111111 000000000 000000000 000000000
)

var ErrInvalidSizeFormat = errors.New("invalid format of tag's/frame's size")
var ErrSizeOverflow = errors.New("size of tag/frame is greater than allowed in id3 tag")

func writeBytesSize(bw *bufWriter, size uint, synchSafe bool) error {
	if synchSafe {
		return writeSynchSafeBytesSize(bw, size)
	}
	return writeSynchUnsafeBytesSize(bw, size)
}

func writeSynchSafeBytesSize(bw *bufWriter, size uint) error {
	if size > synchSafeMaxSize {
		return ErrSizeOverflow
	}

	// First 4 bits of size are always "0", because size should be smaller
	// as maxSize. So skip them.
	size <<= 4

	// Let's explain the algorithm on example.
	// E.g. size is 32-bit integer and after the skip of first 4 bits
	// its value is "10100111 01110101 01010010 11110000".
	// In loop we should write every first 7 bits to bw.
	for i := 0; i < id3SizeLen; i++ {
		// To take first 7 bits we should do `size&mask`.
		firstBits := size & synchSafeMask
		// firstBits is "10100110 00000000 00000000 00000000" now.
		// firstBits has int type, but we should have a byte.
		// To have a byte we should move first 7 bits to the end of firstBits,
		// because by converting int to byte only last 8 bits are taken.
		firstBits >>= (3*8 + 1)
		// firstBits is "00000000 00000000 00000000 01010011" now.
		bSize := byte(firstBits)
		// Now in bSize we have only "01010011". We can write it to bw.
		bw.WriteByte(bSize)
		// Do the same with next 7 bits.
		size <<= synchSafeSizeBase
	}

	return nil
}

func writeSynchUnsafeBytesSize(bw *bufWriter, size uint) error {
	if size > synchUnsafeMaxSize {
		return ErrSizeOverflow
	}

	// See the explanation of algorithm in writeSynchSafeBytesSize.
	for i := 0; i < id3SizeLen; i++ {
		firstBits := size & synchUnsafeMask
		firstBits >>= (3 * 8)
		bw.WriteByte(byte(firstBits))
		size <<= synchUnsafeSizeBase
	}

	return nil
}

func parseSize(data []byte, synchSafe bool) (int64, error) {
	if len(data) > id3SizeLen {
		return 0, ErrInvalidSizeFormat
	}

	var sizeBase uint
	if synchSafe {
		sizeBase = synchSafeSizeBase
	} else {
		sizeBase = synchUnsafeSizeBase
	}

	var size int64
	for _, b := range data {
		if synchSafe && b&128 > 0 { // 128 = 0b1000_0000
			return 0, ErrInvalidSizeFormat
		}

		size = (size << sizeBase) | int64(b)
	}

	return size, nil
}
