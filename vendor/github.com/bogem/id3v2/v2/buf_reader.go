// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"bufio"
	"bytes"
	"io"
)

// bufReader is used for convenient parsing of frames.
type bufReader struct {
	buf *bufio.Reader
	err error
}

// newBufReader returns *bufReader with specified rd.
func newBufReader(rd io.Reader) *bufReader {
	return &bufReader{buf: bufio.NewReader(rd)}
}

func (br *bufReader) Discard(n int) {
	if br.err != nil {
		return
	}
	_, br.err = br.buf.Discard(n)
}

func (br *bufReader) Err() error {
	return br.err
}

// Read calls br.buf.Read(p) and returns the results of it.
// It does nothing, if br.err != nil.
//
// NOTE: if br.buf.Read(p) returns the error, it doesn't save it in
// br.err and it's not returned from br.Err().
func (br *bufReader) Read(p []byte) (n int, err error) {
	if br.err != nil {
		return 0, br.err
	}
	return br.buf.Read(p)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF.
// Because ReadAll is defined to read from src until EOF,
// it does not treat an EOF from Read as an error to be reported.
func (br *bufReader) ReadAll() []byte {
	if br.err != nil {
		return nil
	}
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	_, err := buf.ReadFrom(br)
	if err != nil && br.err == nil {
		br.err = err
		return nil
	}
	return buf.Bytes()
}

func (br *bufReader) ReadByte() byte {
	if br.err != nil {
		return 0
	}
	var b byte
	b, br.err = br.buf.ReadByte()
	return b
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (br *bufReader) Next(n int) []byte {
	if br.err != nil {
		return nil
	}
	var b []byte
	b, br.err = br.next(n)
	return b
}

func (br *bufReader) next(n int) ([]byte, error) {
	if n == 0 {
		return nil, nil
	}

	peeked, err := br.buf.Peek(n)
	if err != nil {
		return nil, err
	}

	if _, err := br.buf.Discard(n); err != nil {
		return nil, err
	}

	return peeked, nil
}

// readTillDelim reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and NOT including the delim.
// If ReadTillDelim encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself.
// ReadTillDelim returns err != nil if and only if ReadTillDelim didn't find
// delim.
func (br *bufReader) readTillDelim(delim byte) ([]byte, error) {
	read, err := br.buf.ReadBytes(delim)
	if err != nil || len(read) == 0 {
		return read, err
	}
	err = br.buf.UnreadByte()
	return read[:len(read)-1], err
}

// readTillDelims reads until the first occurrence of delims in the input,
// returning a slice containing the data up to and NOT including the delimiters.
// If ReadTillDelims encounters an error before finding a delimiters,
// it returns the data read before the error and the error itself.
// ReadTillDelims returns err != nil if and only if ReadTillDelims didn't find
// delims.
func (br *bufReader) readTillDelims(delims []byte) ([]byte, error) {
	if len(delims) == 0 {
		return nil, nil
	}
	if len(delims) == 1 {
		return br.readTillDelim(delims[0])
	}

	result := make([]byte, 0)

	for {
		read, err := br.readTillDelim(delims[0])
		if err != nil {
			return result, err
		}
		result = append(result, read...)

		peeked, err := br.buf.Peek(len(delims))
		if err != nil {
			return result, err
		}

		if bytes.Equal(peeked, delims) {
			break
		}

		b, err := br.buf.ReadByte()
		if err != nil {
			return result, err
		}
		result = append(result, b)
	}

	return result, nil
}

// ReadText reads until the first occurrence of delims in the input,
// returning a slice containing the data up to and NOT including the delimiters.
// But it discards then termination bytes according to provided encoding.
func (br *bufReader) ReadText(encoding Encoding) []byte {
	if br.err != nil {
		return nil
	}

	var text []byte
	delims := encoding.TerminationBytes
	text, br.err = br.readTillDelims(delims)

	// See https://github.com/bogem/id3v2/issues/51.
	if encoding.Equals(EncodingUTF16) &&
		// See https://github.com/bogem/id3v2/issues/53#issuecomment-604038434.
		!bytes.Equal(text, bom) {
		text = append(text, br.ReadByte())
	}

	br.Discard(len(delims))

	return text
}

func (br *bufReader) Reset(rd io.Reader) {
	br.buf.Reset(rd)
	br.err = nil
}
