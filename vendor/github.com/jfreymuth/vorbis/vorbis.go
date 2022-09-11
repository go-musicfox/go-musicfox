package vorbis

import "errors"

// A Decoder stores the information necessary to decode a vorbis steam.
type Decoder struct {
	headerRead bool
	setupRead  bool

	sampleRate int
	channels   int
	Bitrate    Bitrate
	blocksize  [2]int
	CommentHeader

	codebooks []codebook
	floors    []floor
	residues  []residue
	mappings  []mapping
	modes     []mode

	overlap      []float32
	hasOverlap   bool
	overlapShort bool

	windows       [2][]float32
	lookup        [2]imdctLookup
	residueBuffer [][]float32
	floorBuffer   []floorData
	rawBuffer     [][]float32
}

// The Bitrate of a vorbis stream.
// Some or all of the fields can be zero.
type Bitrate struct {
	Nominal int
	Minimum int
	Maximum int
}

// The CommentHeader of a vorbis stream.
type CommentHeader struct {
	Vendor   string
	Comments []string
}

// SampleRate returns the sample rate of the vorbis stream.
// This will be zero if the headers have not been read yet.
func (d *Decoder) SampleRate() int { return d.sampleRate }

// Channels returns the number of channels of the vorbis stream.
// This will be zero if the headers have not been read yet.
func (d *Decoder) Channels() int { return d.channels }

// BufferSize returns the highest amount of data that can be decoded from a single packet.
// The result is already multiplied with the number of channels.
// This will be zero if the headers have not been read yet.
func (d *Decoder) BufferSize() int {
	return d.blocksize[1] / 2 * d.channels
}

// IsHeader returns wether the packet is a vorbis header.
func IsHeader(packet []byte) bool {
	return len(packet) > 6 && packet[0]&1 == 1 &&
		packet[1] == 'v' &&
		packet[2] == 'o' &&
		packet[3] == 'r' &&
		packet[4] == 'b' &&
		packet[5] == 'i' &&
		packet[6] == 's'
}

// ReadHeader reads a vorbis header.
// Three headers (identification, comment, and setup) must be read before any samples can be decoded.
func (d *Decoder) ReadHeader(header []byte) error {
	if !IsHeader(header) {
		return errors.New("vorbis: invalid header")
	}
	headerType := header[0]
	header = header[7:]
	switch headerType {
	case headerTypeIdentification:
		err := d.readIdentificationHeader(header)
		if err != nil {
			return err
		}
		d.headerRead = true
	case headerTypeComment:
		return d.readCommentHeader(header)
	case headerTypeSetup:
		err := d.readSetupHeader(header)
		if err != nil {
			return err
		}
		d.overlap = make([]float32, d.blocksize[1]*d.channels)
		d.setupRead = true
	default:
		return errors.New("vorbis: unknown header type")
	}
	return nil
}

// HeadersRead returns wether the headers necessary for decoding have been read.
func (d *Decoder) HeadersRead() bool {
	return d.headerRead && d.setupRead
}

// Decode decodes a packet and returns the result as an interleaved float slice.
// The number of samples decoded varies and can be zero, but will be at most BufferSize()
func (d *Decoder) Decode(in []byte) ([]float32, error) {
	if !d.HeadersRead() {
		return nil, errors.New("vorbis: missing headers")
	}
	return d.decodePacket(newBitReader(in), nil)
}

// DecodeInto decodes a packet and stores the result in the given buffer.
// The size of the buffer must be at least BufferSize().
// The method will always return a slice of the buffer or nil.
func (d *Decoder) DecodeInto(in []byte, buffer []float32) ([]float32, error) {
	if !d.HeadersRead() {
		return nil, errors.New("vorbis: missing headers")
	}
	if len(buffer) < d.BufferSize() {
		return nil, errors.New("vorbis: buffer too short")
	}
	return d.decodePacket(newBitReader(in), buffer)
}

// Clear must be called between decoding two non-consecutive packets.
func (d *Decoder) Clear() {
	d.hasOverlap = false
}
