package oggvorbis

import (
	"io"

	"github.com/jfreymuth/vorbis"
)

// Format contains information about the audio format of an ogg/vorbis file.
type Format struct {
	SampleRate int
	Channels   int
	Bitrate    vorbis.Bitrate
}

// ReadAll decodes audio from in until an error or EOF.
func ReadAll(in io.Reader) ([]float32, *Format, error) {
	r, err := NewReader(in)
	if err != nil {
		return nil, nil, err
	}
	format := &Format{
		SampleRate: r.SampleRate(),
		Channels:   r.Channels(),
		Bitrate:    r.Bitrate(),
	}
	if r.Length() > 0 {
		result := make([]float32, (r.Length()-r.Position())*int64(r.Channels()))
		read := 0

		for {
			n, err := r.Read(result[read:])
			read += n
			if err != nil || read == len(result) {
				if err == io.EOF {
					err = nil
				}
				return result[:read], format, err
			}
			if n == 0 {
				return result[:read], format, io.ErrNoProgress
			}
		}
	}
	buf := make([]float32, r.dec.BufferSize())
	var result []float32
	for {
		n, err := r.Read(buf)
		result = append(result, buf[:n]...)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return result, format, err
		}
		if n == 0 {
			return result, format, io.ErrNoProgress
		}
	}
}

// GetFormat reads the first ogg page from in to get the audio format.
func GetFormat(in io.Reader) (*Format, error) {
	var dec vorbis.Decoder
	r := oggReader{source: in}
	p, err := r.NextPacket()
	if err != nil {
		return nil, err
	}
	err = dec.ReadHeader(p)
	if err != nil {
		return nil, err
	}
	return &Format{
		SampleRate: dec.SampleRate(),
		Channels:   dec.Channels(),
		Bitrate:    dec.Bitrate,
	}, nil
}

// GetCommentHeader returns a struct containing info from the comment header.
func GetCommentHeader(in io.Reader) (vorbis.CommentHeader, error) {
	var dec vorbis.Decoder
	r := oggReader{source: in}
	for i := 0; i < 2; i++ {
		p, err := r.NextPacket()
		if err != nil {
			return vorbis.CommentHeader{}, err
		}
		err = dec.ReadHeader(p)
		if err != nil {
			return vorbis.CommentHeader{}, err
		}
	}
	return dec.CommentHeader, nil
}

// GetLength returns the length of the file in samples and the audio format.
func GetLength(in io.ReadSeeker) (int64, *Format, error) {
	r := oggReader{source: in, seeker: in}
	var dec vorbis.Decoder
	p, err := r.NextPacket()
	if err != nil {
		return 0, nil, err
	}
	err = dec.ReadHeader(p)
	if err != nil {
		return 0, nil, err
	}
	format := &Format{
		SampleRate: dec.SampleRate(),
		Channels:   dec.Channels(),
		Bitrate:    dec.Bitrate,
	}
	length, err := r.LastPosition()
	return length, format, err
}
