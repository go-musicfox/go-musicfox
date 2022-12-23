// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"sync"
)

// sequence is used to manipulate with frames, which can be in tag
// more than one (e.g. APIC, COMM, USLT and etc.)
type sequence struct {
	frames []Framer
}

func (s *sequence) AddFrame(f Framer) {
	i := indexOfFrame(f, s.frames)

	if i == -1 {
		s.frames = append(s.frames, f)
	} else {
		s.frames[i] = f
	}
}

func indexOfFrame(f Framer, fs []Framer) int {
	for i, ff := range fs {
		if f.UniqueIdentifier() == ff.UniqueIdentifier() {
			return i
		}
	}
	return -1
}

func (s *sequence) Count() int {
	return len(s.frames)
}

func (s *sequence) Frames() []Framer {
	return s.frames
}

var seqPool = sync.Pool{New: func() interface{} {
	return &sequence{frames: []Framer{}}
}}

func getSequence() *sequence {
	s := seqPool.Get().(*sequence)
	if s.Count() > 0 {
		s.frames = []Framer{}
	}
	return s
}

func putSequence(s *sequence) {
	seqPool.Put(s)
}
