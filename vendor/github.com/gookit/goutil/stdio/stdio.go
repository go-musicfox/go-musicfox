// Package stdio provide some standard IO util functions.
package stdio

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// DiscardReader anything from the reader
func DiscardReader(src io.Reader) {
	_, _ = io.Copy(io.Discard, src)
}

// ReadString read contents from io.Reader
func ReadString(r io.Reader) string {
	bs, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return string(bs)
}

// MustReadReader read contents from io.Reader, will panic on error
func MustReadReader(r io.Reader) []byte {
	bs, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return bs
}

// NewIOReader instance by input: string, bytes, io.Reader
func NewIOReader(in any) io.Reader {
	switch typIn := in.(type) {
	case []byte:
		return bytes.NewReader(typIn)
	case string:
		return strings.NewReader(typIn)
	case io.Reader:
		return typIn
	}
	panic("invalid input type for create reader")
}

// NewScanner instance by input data or reader
func NewScanner(in any) *bufio.Scanner {
	switch typIn := in.(type) {
	case io.Reader:
		return bufio.NewScanner(typIn)
	case []byte:
		return bufio.NewScanner(bytes.NewReader(typIn))
	case string:
		return bufio.NewScanner(strings.NewReader(typIn))
	case *bufio.Scanner:
		return typIn
	default:
		panic("invalid input type for create scanner")
	}
}
