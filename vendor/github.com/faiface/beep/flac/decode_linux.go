// github action不好构建暂时只支持linux

package flac

import (
	"fmt"
	"io"

	libflac "github.com/cocoonlife/goflac"
	"github.com/faiface/beep"
	"github.com/pkg/errors"
)

// Decode takes a Reader containing audio data in FLAC format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if r is not io.Seeker.
//
// Do not close the supplied Reader, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode(r io.ReadSeekCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	d := decoder{r: r}
	defer func() { // hacky way to always close r if an error occurred
		if closer, ok := d.r.(io.Closer); ok {
			if err != nil {
				closer.Close()
			}
		}
	}()

	d.d, err = libflac.NewDecoderReader(r)

	if err != nil {
		return nil, beep.Format{}, errors.Wrap(err, "flac")
	}
	format = beep.Format{
		SampleRate:  beep.SampleRate(d.d.Rate),
		NumChannels: d.d.Channels,
		Precision:   d.d.Depth / 8,
	}
	return &d, format, nil
}

type decoder struct {
	r   io.Reader
	d   *libflac.Decoder
	buf [][2]float64
	pos int
	err error
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
	frame, err := d.d.ReadFrame()
	if err != nil {
		return err
	}
	// Expand buffer size if needed.
	var n int
	if d.d.Channels == 1 {
		n = len(frame.Buffer)
	} else {
		n = len(frame.Buffer) / d.d.Channels
	}
	if cap(d.buf) < n {
		d.buf = make([][2]float64, n)
	} else {
		d.buf = d.buf[:n]
	}
	// Decode audio samples.
	bps := d.d.Depth
	nchannels := d.d.Channels
	s := 1 << (bps - 1)
	q := 1 / float64(s)
	switch {
	case bps == 8 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int8(frame.Buffer[i])) * q
			d.buf[i][1] = float64(int8(frame.Buffer[i])) * q
		}
	case bps == 16 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int16(frame.Buffer[i])) * q
			d.buf[i][1] = float64(int16(frame.Buffer[i])) * q
		}
	case bps == 24 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int32(frame.Buffer[i])) * q
			d.buf[i][1] = float64(int32(frame.Buffer[i])) * q
		}
	case bps == 8 && nchannels >= 2:
		var j int
		for i := 0; i < n; i++ {
			j = i << 1
			d.buf[i][0] = float64(int8(frame.Buffer[j])) * q
			d.buf[i][1] = float64(int8(frame.Buffer[j+1])) * q
		}
	case bps == 16 && nchannels >= 2:
		var j int
		for i := 0; i < n; i++ {
			j = i << 1
			d.buf[i][0] = float64(int16(frame.Buffer[j])) * q
			d.buf[i][1] = float64(int16(frame.Buffer[j+1])) * q
		}
	case bps == 24 && nchannels >= 2:
		var j int
		for i := 0; i < n; i++ {
			j = i << 1
			d.buf[i][0] = float64(frame.Buffer[j]) * q
			d.buf[i][1] = float64(frame.Buffer[j+1]) * q
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
	_, total, _ := d.d.Tell()
	return int(total)
}

func (d *decoder) Position() int {
	return d.pos
}

func (d *decoder) Seek(p int) error {
	pos, err := d.d.Seek(uint64(p))
	d.pos = int(pos)
	return err
}

func (d *decoder) Close() error {
	d.d.Close()
	if closer, ok := d.r.(io.Closer); ok {
		_ = closer.Close()
	}
	return nil
}

func (d *decoder) ResetError() {
	d.err = nil
}
