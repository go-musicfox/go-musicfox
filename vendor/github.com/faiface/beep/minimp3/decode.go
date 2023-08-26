// Package minimp3 implements audio data decoding in MP3 format.
package minimp3

import (
	"io"

	"github.com/faiface/beep"
	"github.com/pkg/errors"
	"github.com/tosone/minimp3"
)

func Decode(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "mp3")
		}
	}()
	d, err := minimp3.NewDecoder(rc)
	if err != nil {
		return nil, beep.Format{}, err
	}
	res := <-d.Started()
	if !res {
		return nil, beep.Format{}, errors.New("ctx done")
	}
	format = beep.Format{
		SampleRate:  beep.SampleRate(d.SampleRate),
		NumChannels: d.Channels,
		Precision:   2,
	}
	return &decoder{rc, d, format, 0, nil}, format, nil
}

type decoder struct {
	closer io.Closer
	d      *minimp3.Decoder
	f      beep.Format
	pos    int
	err    error
}

func (d *decoder) Stream(samples [][2]float64) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	var tmp = make([]byte, d.f.NumChannels*d.f.Precision)
	for i := range samples {
		dn, err := d.d.Read(tmp[:])
		if dn == len(tmp) {
			samples[i], _ = d.f.DecodeSigned(tmp[:])
			d.pos += dn
			n++
			ok = true
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			d.err = errors.Wrap(err, "mp3")
			break
		}
	}
	return n, ok
}

func (d *decoder) Err() error {
	return d.err
}

func (d *decoder) Len() int {
	return 1
}

func (d *decoder) Position() int {
	return d.pos / (d.f.Precision * d.f.NumChannels)
}

func (d *decoder) Seek(_ int) error {
	return errors.New("unimplemented")
}

func (d *decoder) Close() error {
	d.d.Close()
	err := d.closer.Close()
	if err != nil {
		return errors.Wrap(err, "mp3")
	}
	return nil
}

func (d *decoder) ResetError() {
	d.err = nil
}
