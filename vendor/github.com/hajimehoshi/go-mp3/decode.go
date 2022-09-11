// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mp3

import (
	"errors"
	"io"

	"github.com/hajimehoshi/go-mp3/internal/consts"
	"github.com/hajimehoshi/go-mp3/internal/frame"
	"github.com/hajimehoshi/go-mp3/internal/frameheader"
)

// A Decoder is a MP3-decoded stream.
//
// Decoder decodes its underlying source on the fly.
type Decoder struct {
	source        *source
	sampleRate    int
	length        int64
	frameStarts   []int64
	buf           []byte
	frame         *frame.Frame
	pos           int64
	bytesPerFrame int64
}

func (d *Decoder) readFrame() error {
	var err error
	d.frame, _, err = frame.Read(d.source, d.source.pos, d.frame)
	if err != nil {
		if err == io.EOF {
			return io.EOF
		}
		if _, ok := err.(*consts.UnexpectedEOF); ok {
			// TODO: Log here?
			return io.EOF
		}
		return err
	}
	d.buf = append(d.buf, d.frame.Decode()...)
	return nil
}

// Read is io.Reader's Read.
func (d *Decoder) Read(buf []byte) (int, error) {
	for len(d.buf) == 0 {
		if err := d.readFrame(); err != nil {
			return 0, err
		}
	}
	n := copy(buf, d.buf)
	d.buf = d.buf[n:]
	d.pos += int64(n)
	return n, nil
}

// Seek is io.Seeker's Seek.
//
// Seek returns an error when the underlying source is not io.Seeker.
//
// Note that seek uses a byte offset but samples are aligned to 4 bytes (2
// channels, 2 bytes each). Be careful to seek to an offset that is divisible by
// 4 if you want to read at full sample boundaries.
func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekCurrent {
		// Handle the special case of asking for the current position specially.
		return d.pos, nil
	}

	npos := int64(0)
	switch whence {
	case io.SeekStart:
		npos = offset
	case io.SeekCurrent:
		npos = d.pos + offset
	case io.SeekEnd:
		npos = d.Length() + offset
	default:
		return 0, errors.New("mp3: invalid whence")
	}
	d.pos = npos
	d.buf = nil
	d.frame = nil
	f := d.pos / d.bytesPerFrame
	// If the frame is not first, read the previous ahead of reading that
	// because the previous frame can affect the targeted frame.
	if f > 0 {
		f--
		if _, err := d.source.Seek(d.frameStarts[f], 0); err != nil {
			return 0, err
		}
		if err := d.readFrame(); err != nil {
			return 0, err
		}
		if err := d.readFrame(); err != nil {
			return 0, err
		}
		d.buf = d.buf[d.bytesPerFrame+(d.pos%d.bytesPerFrame):]
	} else {
		if _, err := d.source.Seek(d.frameStarts[f], 0); err != nil {
			return 0, err
		}
		if err := d.readFrame(); err != nil {
			return 0, err
		}
		d.buf = d.buf[d.pos:]
	}
	return npos, nil
}

// SampleRate returns the sample rate like 44100.
//
// Note that the sample rate is retrieved from the first frame.
func (d *Decoder) SampleRate() int {
	return d.sampleRate
}

func (d *Decoder) ensureFrameStartsAndLength() error {
	if d.length != invalidLength {
		return nil
	}

	if _, ok := d.source.reader.(io.Seeker); !ok {
		return nil
	}

	// Keep the current position.
	pos, err := d.source.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := d.source.rewind(); err != nil {
		return err
	}

	if err := d.source.skipTags(); err != nil {
		return err
	}
	l := int64(0)
	for {
		h, pos, err := frameheader.Read(d.source, d.source.pos)
		if err != nil {
			if err == io.EOF {
				break
			}
			if _, ok := err.(*consts.UnexpectedEOF); ok {
				// TODO: Log here?
				break
			}
			return err
		}
		d.frameStarts = append(d.frameStarts, pos)
		d.bytesPerFrame = int64(h.BytesPerFrame())
		l += d.bytesPerFrame

		framesize, err := h.FrameSize()
		if err != nil {
			return err
		}
		buf := make([]byte, framesize-4)
		if _, err := d.source.ReadFull(buf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	d.length = l

	if _, err := d.source.Seek(pos, io.SeekStart); err != nil {
		return err
	}
	return nil
}

const invalidLength = -1

// Length returns the total size in bytes.
//
// Length returns -1 when the total size is not available
// e.g. when the given source is not io.Seeker.
func (d *Decoder) Length() int64 {
	return d.length
}

// NewDecoder decodes the given io.Reader and returns a decoded stream.
//
// The stream is always formatted as 16bit (little endian) 2 channels
// even if the source is single channel MP3.
// Thus, a sample always consists of 4 bytes.
func NewDecoder(r io.Reader) (*Decoder, error) {
	s := &source{
		reader: r,
	}
	d := &Decoder{
		source: s,
		length: invalidLength,
	}

	if err := s.skipTags(); err != nil {
		return nil, err
	}
	// TODO: Is readFrame here really needed?
	if err := d.readFrame(); err != nil {
		return nil, err
	}
	freq, err := d.frame.SamplingFrequency()
	if err != nil {
		return nil, err
	}
	d.sampleRate = freq

	if err := d.ensureFrameStartsAndLength(); err != nil {
		return nil, err
	}

	return d, nil
}
