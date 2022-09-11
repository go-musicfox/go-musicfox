// Package vorbis implements audio data decoding in oggvorbis format.
package vorbis

import (
	"io"

	"github.com/faiface/beep"
	"github.com/jfreymuth/oggvorbis"
	"github.com/pkg/errors"
)

const (
	govorbisNumChannels = 2
	govorbisPrecision   = 2
)

// Decode takes a ReadCloser containing audio data in ogg/vorbis format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if rc is not io.Seeker.
//
// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "ogg/vorbis")
		}
	}()
	d, err := oggvorbis.NewReader(rc)
	if err != nil {
		return nil, beep.Format{}, err
	}
	format = beep.Format{
		SampleRate:  beep.SampleRate(d.SampleRate()),
		NumChannels: govorbisNumChannels,
		Precision:   govorbisPrecision,
	}
	return &decoder{rc, d, format, nil}, format, nil
}

type decoder struct {
	closer io.Closer
	d      *oggvorbis.Reader
	f      beep.Format
	err    error
}

func (d *decoder) Stream(samples [][2]float64) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	var tmp [2]float32
	for i := range samples {
		dn, err := d.d.Read(tmp[:])
		if dn == 2 {
			samples[i][0], samples[i][1] = float64(tmp[0]), float64(tmp[1])
			n++
			ok = true
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			d.err = errors.Wrap(err, "ogg/vorbis")
			break
		}
	}
	return n, ok
}

func (d *decoder) Err() error {
	return d.err
}

func (d *decoder) Len() int {
	return int(d.d.Length())
}

func (d *decoder) Position() int {
	return int(d.d.Position())
}

func (d *decoder) Seek(p int) error {
	err := d.d.SetPosition(int64(p))
	if err != nil {
		return errors.Wrap(err, "ogg/vorbis")
	}
	return nil
}

func (d *decoder) Close() error {
	err := d.closer.Close()
	if err != nil {
		return errors.Wrap(err, "ogg/vorbis")
	}
	return nil
}
