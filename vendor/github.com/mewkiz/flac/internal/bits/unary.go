package bits

import (
	"github.com/icza/bitio"
)

// ReadUnary decodes and returns an unary coded integer, whose value is
// represented by the number of leading zeros before a one.
//
// Examples of unary coded binary on the left and decoded decimal on the right:
//
//    1       => 0
//    01      => 1
//    001     => 2
//    0001    => 3
//    00001   => 4
//    000001  => 5
//    0000001 => 6
func (br *Reader) ReadUnary() (x uint64, err error) {
	for {
		bit, err := br.Read(1)
		if err != nil {
			return 0, err
		}
		if bit == 1 {
			break
		}
		x++
	}
	return x, nil
}

// WriteUnary encodes x as an unary coded integer, whose value is represented by
// the number of leading zeros before a one.
//
// Examples of unary coded binary on the left and decoded decimal on the right:
//
//    0 => 1
//    1 => 01
//    2 => 001
//    3 => 0001
//    4 => 00001
//    5 => 000001
//    6 => 0000001
func WriteUnary(bw *bitio.Writer, x uint64) error {
	bits := uint64(1)
	n := byte(1)
	for ; x > 0; x-- {
		n++
	}
	if err := bw.WriteBits(bits, n); err != nil {
		return err
	}
	return nil
}
