package oggvorbis

import (
	"errors"
	"io"

	"github.com/jfreymuth/vorbis"
)

// A Reader can read audio from an ogg/vorbis file.
type Reader struct {
	r   oggReader
	dec vorbis.Decoder

	position       int64
	buffer         []float32
	originalBuffer []float32
	toSkip         int

	length int64
}

// NewReader creates a new Reader.
// Some of the returned reader's methods will only work if in also implements
// io.Seeker
func NewReader(in io.Reader) (*Reader, error) {
	r := new(Reader)
	r.r.source = in
	r.r.seeker, _ = in.(io.Seeker)
	if err := r.init(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reader) init() error {
	if r.r.seeker != nil {
		length, err := r.r.LastPosition()
		if err == nil {
			r.length = length
		}
		r.r.buffer = nil
		r.r.seeker.Seek(0, io.SeekStart)
	}

	// read headers
	for i := 0; i < 3; i++ {
		packet, err := r.r.NextPacket()
		if err != nil {
			return noEOF(err)
		}
		err = r.dec.ReadHeader(packet)
		if err != nil {
			return err
		}
	}
	r.originalBuffer = make([]float32, r.dec.BufferSize())

	// read the first two packets, because the data might not start at position zero, see:
	// https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-132000A.2
	err := r.fillBuffer()
	if err != nil {
		return err
	}
	if r.r.packetIndex == r.r.currentPage.packetCount {
		offset := r.r.currentPage.AbsoluteGranulePosition - r.position
		if offset < 0 {
			r.buffer = r.buffer[-offset:]
		}
		r.position = r.r.currentPage.AbsoluteGranulePosition
	}

	return nil
}

// SampleRate returns the sample rate of the vorbis stream.
func (r *Reader) SampleRate() int { return r.dec.SampleRate() }

// Channels returns the number of channels of the vorbis stream.
func (r *Reader) Channels() int { return r.dec.Channels() }

// Bitrate returns a struct containing information about the bitrate.
func (r *Reader) Bitrate() vorbis.Bitrate { return r.dec.Bitrate }

// CommentHeader returns a struct containing info from the comment header.
func (r *Reader) CommentHeader() vorbis.CommentHeader { return r.dec.CommentHeader }

// Position returns the current position in samples.
func (r *Reader) Position() int64 {
	return r.position - int64(len(r.buffer)/r.Channels()) + int64(r.toSkip/r.Channels())
}

// Length returns the length of the audio data in samples.
// A return value of zero means the length is unknown, probably because the
// underlying reader is not seekable.
func (r *Reader) Length() int64 { return r.length }

// SetPosition seeks to a position in samples.
// It will return an error if the underlying reader is not seekable.
func (r *Reader) SetPosition(pos int64) error {
	if r.r.seeker == nil {
		return errors.New("oggvorbis: reader is not seekable")
	}
	r.buffer = nil
	if pos >= r.length {
		r.r.lastPacket = true
		r.position = r.length
		return nil
	}
	start, err := r.r.SeekPageBefore(pos)
	if err != nil {
		return err
	}
	packet, err := r.r.NextPacket()
	r.dec.Clear()
	r.dec.DecodeInto(packet, r.originalBuffer)
	r.position = start
	r.toSkip = int(pos-start) * r.Channels()
	return nil
}

// Read reads and decodes audio data and stores the result in p.
//
// It returns the number of values decoded (number of samples * channels) and
// any error encountered, similarly to an io.Reader's Read method.
//
// The number of values produced will always be a multiple of Channels().
func (r *Reader) Read(p []float32) (int, error) {
	if r.r.lastPacket && len(r.buffer) == 0 {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	if len(p)%r.Channels() != 0 {
		p = p[:len(p)/r.Channels()*r.Channels()]
	}
	out := p
	total := 0
	err := error(nil)
	if r.toSkip > 0 {
		err = r.skip()
		if err != nil {
			return 0, err
		}
	}
	if len(r.buffer) > 0 {
		n := copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		total += n
		p = p[n:]
	}
	for len(p) >= len(r.originalBuffer) {
		var n int
		n, err = r.read(p)
		total += n
		if err != nil {
			goto end
		}
		p = p[n:]
	}
	if total == 0 {
		err = r.fillBuffer()
		if err != nil {
			goto end
		}
		n := copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		total += n
		p = p[n:]
	}
end:
	for i := range out[:total] {
		if out[i] > 1 {
			out[i] = 1
		} else if out[i] < -1 {
			out[i] = -1
		}
	}
	return total, err
}

func (r *Reader) fillBuffer() error {
	n, err := r.read(r.originalBuffer)
	r.buffer = r.originalBuffer[:n]
	if err != nil {
		return err
	}
	if n == 0 {
		return r.fillBuffer()
	}
	return nil
}

func (r *Reader) read(p []float32) (int, error) {
	packet, err := r.r.NextPacket()
	if err != nil {
		return 0, err
	}
	out, err := r.dec.DecodeInto(packet, p)
	n := len(out)
	r.position += int64(n / r.Channels())
	if err != nil {
		return n, err
	}
	if r.r.lastPacket {
		discard := int(r.position-r.r.currentPage.AbsoluteGranulePosition) * r.Channels()
		if discard > 0 {
			n -= discard
		}
		r.position = r.r.currentPage.AbsoluteGranulePosition
		r.length = r.position
	}
	return n, nil
}

func (r *Reader) skip() error {
	for {
		err := r.fillBuffer()
		if len(r.buffer) > r.toSkip {
			r.buffer = r.buffer[r.toSkip:]
			r.toSkip = 0
			return err
		}
		r.toSkip -= len(r.buffer)
		if err != nil {
			return err
		}
	}
}
