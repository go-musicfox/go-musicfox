// TODO(u): Evaluate storing the samples (and residuals) during frame audio
// decoding in a buffer allocated for the stream. This buffer would be allocated
// using BlockSize and NChannels from the StreamInfo block, and it could be
// reused in between calls to Next and ParseNext. This should reduce GC
// pressure.

// TODO: Remove note about encoder API.

// Package flac provides access to FLAC (Free Lossless Audio Codec) streams.
//
// A brief introduction of the FLAC stream format [1] follows. Each FLAC stream
// starts with a 32-bit signature ("fLaC"), followed by one or more metadata
// blocks, and then one or more audio frames. The first metadata block
// (StreamInfo) describes the basic properties of the audio stream and it is the
// only mandatory metadata block. Subsequent metadata blocks may appear in an
// arbitrary order.
//
// Please refer to the documentation of the meta [2] and the frame [3] packages
// for a brief introduction of their respective formats.
//
//    [1]: https://www.xiph.org/flac/format.html#stream
//    [2]: https://godoc.org/github.com/mewkiz/flac/meta
//    [3]: https://godoc.org/github.com/mewkiz/flac/frame
//
// Note: the Encoder API is experimental until the 1.1.x release. As such, it's
// API is expected to change.
package flac

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

// A Stream contains the metadata blocks and provides access to the audio frames
// of a FLAC stream.
//
// ref: https://www.xiph.org/flac/format.html#stream
type Stream struct {
	// The StreamInfo metadata block describes the basic properties of the FLAC
	// audio stream.
	Info *meta.StreamInfo
	// Zero or more metadata blocks.
	Blocks []*meta.Block

	// seekTable contains one or more pre-calculated audio frame seek points of
	// the stream; nil if uninitialized.
	seekTable *meta.SeekTable
	// seekTableSize determines how many seek points the seekTable should have if
	// the flac file does not include one in the metadata.
	seekTableSize int
	// dataStart is the offset of the first frame header since SeekPoint.Offset
	// is relative to this position.
	dataStart int64

	// Underlying io.Reader.
	r io.Reader
	// Underlying io.Closer of file if opened with Open and ParseFile, and nil
	// otherwise.
	c io.Closer
}

// New creates a new Stream for accessing the audio samples of r. It reads and
// parses the FLAC signature and the StreamInfo metadata block, but skips all
// other metadata blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
func New(r io.Reader) (stream *Stream, err error) {
	// Verify FLAC signature and parse the StreamInfo metadata block.
	br := bufio.NewReader(r)
	stream = &Stream{r: br}
	block, err := stream.parseStreamInfo()
	if err != nil {
		return nil, err
	}

	// Skip the remaining metadata blocks.
	for !block.IsLast {
		block, err = meta.New(br)
		if err != nil && err != meta.ErrReservedType {
			return stream, err
		}
		if err = block.Skip(); err != nil {
			return stream, err
		}
	}

	return stream, nil
}

// NewSeek returns a Stream that has seeking enabled. The incoming io.ReadSeeker
// will not be buffered, which might result in performance issues. Using an
// in-memory buffer like *bytes.Reader should work well.
func NewSeek(rs io.ReadSeeker) (stream *Stream, err error) {
	stream = &Stream{r: rs, seekTableSize: defaultSeekTableSize}

	// Verify FLAC signature and parse the StreamInfo metadata block.
	block, err := stream.parseStreamInfo()
	if err != nil {
		return stream, err
	}

	for !block.IsLast {
		block, err = meta.Parse(stream.r)
		if err != nil {
			if err != meta.ErrReservedType {
				return stream, err
			}
			if err = block.Skip(); err != nil {
				return stream, err
			}
		}

		if block.Header.Type == meta.TypeSeekTable {
			stream.seekTable = block.Body.(*meta.SeekTable)
		}
	}

	// Record file offset of the first frame header.
	stream.dataStart, err = rs.Seek(0, io.SeekCurrent)
	return stream, err
}

var (
	// flacSignature marks the beginning of a FLAC stream.
	flacSignature = []byte("fLaC")

	// id3Signature marks the beginning of an ID3 stream, used to skip over ID3
	// data.
	id3Signature = []byte("ID3")

	// ErrNoSeeker reports that flac.NewSeek was called with an io.Reader not
	// implementing io.Seeker, and thus does not allow for seeking.
	ErrNoSeeker = errors.New("stream.Seek: reader does not implement io.Seeker")

	// ErrNoSeektable reports that no seektable has been generated. Therefore,
	// it is not possible to seek in the stream.
	ErrNoSeektable = errors.New("stream.searchFromStart: no seektable exists")
)

const (
	defaultSeekTableSize = 100
)

// parseStreamInfo verifies the signature which marks the beginning of a FLAC
// stream, and parses the StreamInfo metadata block. It returns a boolean value
// which specifies if the StreamInfo block was the last metadata block of the
// FLAC stream.
func (stream *Stream) parseStreamInfo() (block *meta.Block, err error) {
	// Verify FLAC signature.
	r := stream.r
	var buf [4]byte
	if _, err = io.ReadFull(r, buf[:]); err != nil {
		return block, err
	}

	// Skip prepended ID3v2 data.
	if bytes.Equal(buf[:3], id3Signature) {
		if err := stream.skipID3v2(); err != nil {
			return block, err
		}

		// Second attempt at verifying signature.
		if _, err = io.ReadFull(r, buf[:]); err != nil {
			return block, err
		}
	}

	if !bytes.Equal(buf[:], flacSignature) {
		return block, fmt.Errorf("flac.parseStreamInfo: invalid FLAC signature; expected %q, got %q", flacSignature, buf)
	}

	// Parse StreamInfo metadata block.
	block, err = meta.Parse(r)
	if err != nil {
		return block, err
	}
	si, ok := block.Body.(*meta.StreamInfo)
	if !ok {
		return block, fmt.Errorf("flac.parseStreamInfo: incorrect type of first metadata block; expected *meta.StreamInfo, got %T", block.Body)
	}
	stream.Info = si
	return block, nil
}

// skipID3v2 skips ID3v2 data prepended to flac files.
func (stream *Stream) skipID3v2() error {
	r := bufio.NewReader(stream.r)

	// Discard unnecessary data from the ID3v2 header.
	if _, err := r.Discard(2); err != nil {
		return err
	}

	// Read the size from the ID3v2 header.
	var sizeBuf [4]byte
	if _, err := r.Read(sizeBuf[:]); err != nil {
		return err
	}
	// The size is encoded as a synchsafe integer.
	size := int(sizeBuf[0])<<21 | int(sizeBuf[1])<<14 | int(sizeBuf[2])<<7 | int(sizeBuf[3])

	_, err := r.Discard(size)
	return err
}

// Parse creates a new Stream for accessing the metadata blocks and audio
// samples of r. It reads and parses the FLAC signature and all metadata blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
func Parse(r io.Reader) (stream *Stream, err error) {
	// Verify FLAC signature and parse the StreamInfo metadata block.
	br := bufio.NewReader(r)
	stream = &Stream{r: br}
	block, err := stream.parseStreamInfo()
	if err != nil {
		return nil, err
	}

	// Parse the remaining metadata blocks.
	for !block.IsLast {
		block, err = meta.Parse(br)
		if err != nil {
			if err != meta.ErrReservedType {
				return stream, err
			}
			// Skip the body of unknown (reserved) metadata blocks, as stated by
			// the specification.
			//
			// ref: https://www.xiph.org/flac/format.html#format_overview
			if err = block.Skip(); err != nil {
				return stream, err
			}
		}
		stream.Blocks = append(stream.Blocks, block)
	}

	return stream, nil
}

// Open creates a new Stream for accessing the audio samples of path. It reads
// and parses the FLAC signature and the StreamInfo metadata block, but skips
// all other metadata blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
//
// Note: The Close method of the stream must be called when finished using it.
func Open(path string) (stream *Stream, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	stream, err = New(f)
	if err != nil {
		return nil, err
	}
	stream.c = f
	return stream, err
}

// ParseFile creates a new Stream for accessing the metadata blocks and audio
// samples of path. It reads and parses the FLAC signature and all metadata
// blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
//
// Note: The Close method of the stream must be called when finished using it.
func ParseFile(path string) (stream *Stream, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	stream, err = Parse(f)
	if err != nil {
		return nil, err
	}
	stream.c = f
	return stream, err
}

// Close closes the stream if opened through a call to Open or ParseFile, and
// performs no operation otherwise.
func (stream *Stream) Close() error {
	if stream.c != nil {
		return stream.c.Close()
	}
	return nil
}

// Next parses the frame header of the next audio frame. It returns io.EOF to
// signal a graceful end of FLAC stream.
//
// Call Frame.Parse to parse the audio samples of its subframes.
func (stream *Stream) Next() (f *frame.Frame, err error) {
	return frame.New(stream.r)
}

// ParseNext parses the entire next frame including audio samples. It returns
// io.EOF to signal a graceful end of FLAC stream.
func (stream *Stream) ParseNext() (f *frame.Frame, err error) {
	return frame.Parse(stream.r)
}

// Seek seeks to the frame containing the given absolute sample number. The
// return value specifies the first sample number of the frame containing
// sampleNum.
func (stream *Stream) Seek(sampleNum uint64) (uint64, error) {
	if stream.seekTable == nil && stream.seekTableSize > 0 {
		if err := stream.makeSeekTable(); err != nil {
			return 0, err
		}
	}

	rs := stream.r.(io.ReadSeeker)

	isBiggerThanStream := stream.Info.NSamples != 0 && sampleNum > stream.Info.NSamples
	if isBiggerThanStream || sampleNum < 0 {
		return 0, fmt.Errorf("unable to seek to sample number %d", sampleNum)
	}
	point, err := stream.searchFromStart(sampleNum)
	if err != nil {
		return 0, err
	}

	if _, err := rs.Seek(stream.dataStart+int64(point.Offset), io.SeekStart); err != nil {
		return 0, err
	}
	for {
		// Record seek offset to start of frame.
		offset, err := rs.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, err
		}
		frame, err := stream.ParseNext()
		if err != nil {
			return 0, err
		}
		if frame.SampleNumber()+uint64(frame.BlockSize) >= sampleNum {
			// Restore seek offset to the start of the frame containing the
			// specified sample number.
			_, err := rs.Seek(offset, io.SeekStart)
			return frame.SampleNumber(), err
		}
	}
}

// TODO(_): Utilize binary search in searchFromStart.

// searchFromStart searches for the given sample number from the start of the
// seek table and returns the last seek point containing the sample number. If
// no seek point contains the sample number, the last seek point preceding the
// sample number is returned. If the sample number is lower than the first seek
// point, the first seek point is returned.
func (stream *Stream) searchFromStart(sampleNum uint64) (meta.SeekPoint, error) {
	if len(stream.seekTable.Points) == 0 {
		return meta.SeekPoint{}, ErrNoSeektable
	}
	prev := stream.seekTable.Points[0]
	for _, p := range stream.seekTable.Points {
		if p.SampleNum+uint64(p.NSamples) >= sampleNum {
			return prev, nil
		}
		prev = p
	}
	return prev, nil
}

// makeSeekTable creates a seek table with seek points to each frame of the FLAC
// stream.
func (stream *Stream) makeSeekTable() (err error) {
	rs, ok := stream.r.(io.ReadSeeker)
	if !ok {
		return ErrNoSeeker
	}

	pos, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = rs.Seek(stream.dataStart, io.SeekStart)
	if err != nil {
		return err
	}

	var i int
	var sampleNum uint64
	var points []meta.SeekPoint
	for {
		// Record seek offset to start of frame.
		off, err := rs.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		f, err := stream.ParseNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		points = append(points, meta.SeekPoint{
			SampleNum: sampleNum,
			Offset:    uint64(off - stream.dataStart),
			NSamples:  f.BlockSize,
		})

		sampleNum += uint64(f.BlockSize)
		i++
	}

	stream.seekTable = &meta.SeekTable{Points: points}

	_, err = rs.Seek(pos, io.SeekStart)
	return err
}
