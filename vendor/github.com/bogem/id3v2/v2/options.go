// Copyright 2017 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

// Options influence on processing the tag.
type Options struct {
	// Parse defines, if tag will be parsed.
	Parse bool

	// ParseFrames defines, that frames do you only want to parse. For example,
	// `ParseFrames: []string{"Artist", "Title"}` will only parse artist
	// and title frames. You can specify IDs ("TPE1", "TIT2") as well as
	// descriptions ("Artist", "Title"). If ParseFrame is blank or nil,
	// id3v2 will parse all frames in tag. It works only if Parse is true.
	//
	// It's very useful for performance, so for example
	// if you want to get only some text frames,
	// id3v2 will not parse huge picture or unknown frames.
	ParseFrames []string
}
