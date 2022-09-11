package meta

import (
	"errors"
	"io"
	"io/ioutil"
)

// verifyPadding verifies the body of a Padding metadata block. It should only
// contain zero-padding.
//
// ref: https://www.xiph.org/flac/format.html#metadata_block_padding
func (block *Block) verifyPadding() error {
	zr := zeros{r: block.lr}
	_, err := io.Copy(ioutil.Discard, zr)
	return err
}

// Errors returned by zeros.Read.
var (
	ErrInvalidPadding = errors.New("invalid padding")
)

// zeros implements an io.Reader, with a Read method which returns an error if
// any byte read isn't zero.
type zeros struct {
	r io.Reader
}

// Read returns an error if any byte read isn't zero.
func (zr zeros) Read(p []byte) (n int, err error) {
	n, err = zr.r.Read(p)
	for i := 0; i < n; i++ {
		if p[i] != 0 {
			return n, ErrInvalidPadding
		}
	}
	return n, err
}
