// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"errors"
	"io"
)

var ErrInvalidLanguageLength = errors.New("language code must consist of three letters according to ISO 639-2")

// Framer provides a generic interface for frames.
// You can create your own frames. They must implement only this interface.
type Framer interface {
	// Size returns the size of frame body.
	Size() int

	// UniqueIdentifier returns the string that makes this frame unique from others.
	// For example, some frames with same id can be added in tag, but they should be differ in other properties.
	// E.g. It would be "Description" for TXXX and APIC.
	//
	// Frames that can be added only once with same id (e.g. all text frames) should return just "".
	UniqueIdentifier() string

	// WriteTo writes body slice into w.
	WriteTo(w io.Writer) (n int64, err error)
}
