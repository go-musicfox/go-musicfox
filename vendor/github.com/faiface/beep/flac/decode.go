//go:build !linux
// +build !linux

package flac

import (
	"fmt"
	"io"

	"github.com/faiface/beep"
	"github.com/mewkiz/flac"
	"github.com/pkg/errors"
)

// Decode takes a Reader containing audio data in FLAC format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if r is not io.Seeker.
//
// Do not close the supplied Reader, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode(r io.Reader) (s beep.StreamSeekCloser, format beep.Format, err error) {
	d := decoder{r: r}
	defer func() { // hacky way to always close r if an error occurred
		if closer, ok := d.r.(io.Closer); ok {
			if err != nil {
				closer.Close()
			}
		}
	}()

	rs, seeker := r.(io.ReadSeeker)
	if seeker {
		d.stream, err = flac.NewSeek(rs)
		d.seekEnabled = true
	} else {
		d.stream, err = flac.New(r)
	}

	if err != nil {
		return nil, beep.Format{}, errors.Wrap(err, "flac")
	}
	format = beep.Format{
		SampleRate:  beep.SampleRate(d.stream.Info.SampleRate),
		NumChannels: int(d.stream.Info.NChannels),
		Precision:   int(d.stream.Info.BitsPerSample / 8),
	}
	return &d, format, nil
}

type decoder struct {
	r           io.Reader
	stream      *flac.Stream
	buf         [][2]float64
	pos         int
	err         error
	seekEnabled bool
}

func (d *decoder) Stream(samples [][2]float64) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	// Copy samples from buffer.
	j := 0
	for i := range samples {
		if j >= len(d.buf) {
			// refill buffer.
			if err := d.refill(); err != nil {
				d.err = err
				d.pos += n
				return n, n > 0
			}
			j = 0
		}
		samples[i] = d.buf[j]
		j++
		n++
	}
	d.buf = d.buf[j:]
	d.pos += n
	return n, true
}

// refill decodes audio samples to fill the decode buffer.
func (d *decoder) refill() error {
	// Empty buffer.
	d.buf = d.buf[:0]
	// Parse audio frame.
	frame, err := d.stream.ParseNext()
	if err != nil {
		return err
	}
	// Expand buffer size if needed.
	n := len(frame.Subframes[0].Samples)
	if cap(d.buf) < n {
		d.buf = make([][2]float64, n)
	} else {
		d.buf = d.buf[:n]
	}
	// Decode audio samples.
	bps := d.stream.Info.BitsPerSample
	nchannels := d.stream.Info.NChannels
	s := 1 << (bps - 1)
	q := 1 / float64(s)
	switch {
	case bps == 8 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int8(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = float64(int8(frame.Subframes[0].Samples[i])) * q
		}
	case bps == 16 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int16(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = float64(int16(frame.Subframes[0].Samples[i])) * q
		}
	case bps == 24 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int32(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = float64(int32(frame.Subframes[0].Samples[i])) * q
		}
	case bps == 8 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int8(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = float64(int8(frame.Subframes[1].Samples[i])) * q
		}
	case bps == 16 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int16(frame.Subframes[0].Samples[i])) * q
			d.buf[i][1] = float64(int16(frame.Subframes[1].Samples[i])) * q
		}
	case bps == 24 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(frame.Subframes[0].Samples[i]) * q
			d.buf[i][1] = float64(frame.Subframes[1].Samples[i]) * q
		}
	default:
		panic(fmt.Errorf("support for %d bits-per-sample and %d channels combination not yet implemented", bps, nchannels))
	}
	return nil
}

func (d *decoder) Err() error {
	return d.err
}

func (d *decoder) Len() int {
	return int(d.stream.Info.NSamples)
}

func (d *decoder) Position() int {
	return d.pos
}

// p represents flac sample num perhaps?
func (d *decoder) Seek(p int) error {
	if !d.seekEnabled {
		return errors.New("flac.decoder.Seek: not enabled")
	}

	pos, err := d.stream.Seek(uint64(p))
	d.pos = int(pos)
	return err
}

func (d *decoder) Close() error {
	if closer, ok := d.r.(io.Closer); ok {
		err := closer.Close()
		if err != nil {
			return errors.Wrap(err, "flac")
		}
	}
	return nil
}

func (d *decoder) ResetError() {
	d.err = nil
}
