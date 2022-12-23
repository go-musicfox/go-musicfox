// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"bytes"
	"errors"
	"io"
)

const tagHeaderSize = 10

var (
	id3Identifier = []byte("ID3")
	errNoTag      = errors.New("there is no tag in file")
)

var ErrSmallHeaderSize = errors.New("size of tag header is less than expected")

type tagHeader struct {
	FramesSize int64
	Version    byte
}

// parseHeader parses tag header in rd.
// If there is no tag in rd, it returns errNoTag.
// If rd is smaller than expected, it returns ErrSmallHeaderSize.
func parseHeader(rd io.Reader) (tagHeader, error) {
	var header tagHeader

	data := make([]byte, tagHeaderSize)
	n, err := rd.Read(data)
	if err != nil {
		return header, err
	}
	if n < tagHeaderSize {
		return header, ErrSmallHeaderSize
	}

	if !isID3Tag(data[0:3]) {
		return header, errNoTag
	}

	header.Version = data[3]

	// Tag header size is always synchsafe.
	size, err := parseSize(data[6:], true)
	if err != nil {
		return header, err
	}

	header.FramesSize = size
	return header, nil
}

func isID3Tag(data []byte) bool {
	return len(data) == len(id3Identifier) && bytes.Equal(data, id3Identifier)
}
