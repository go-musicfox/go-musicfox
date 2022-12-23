// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package id3v2 is the ID3 parsing and writing library for Go.
package id3v2

import (
	"io"
	"os"
)

// Available picture types for picture frame.
const (
	PTOther = iota
	PTFileIcon
	PTOtherFileIcon
	PTFrontCover
	PTBackCover
	PTLeafletPage
	PTMedia
	PTLeadArtistSoloist
	PTArtistPerformer
	PTConductor
	PTBandOrchestra
	PTComposer
	PTLyricistTextWriter
	PTRecordingLocation
	PTDuringRecording
	PTDuringPerformance
	PTMovieScreenCapture
	PTBrightColouredFish
	PTIllustration
	PTBandArtistLogotype
	PTPublisherStudioLogotype
)

// Open opens file with name and passes it to OpenFile.
// If there is no tag in file, it will create new one with version ID3v2.4.
func Open(name string, opts Options) (*Tag, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return ParseReader(file, opts)
}

// ParseReader parses rd and finds tag in it considering opts.
// If there is no tag in rd, it will create new one with version ID3v2.4.
func ParseReader(rd io.Reader, opts Options) (*Tag, error) {
	tag := NewEmptyTag()
	err := tag.parse(rd, opts)
	return tag, err
}

// NewEmptyTag returns an empty ID3v2.4 tag without any frames and reader.
func NewEmptyTag() *Tag {
	tag := new(Tag)
	tag.init(nil, 0, 4)
	return tag
}
