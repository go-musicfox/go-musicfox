package meta

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"

	"github.com/mewkiz/flac/internal/bits"
)

// StreamInfo contains the basic properties of a FLAC audio stream, such as its
// sample rate and channel count. It is the only mandatory metadata block and
// must be present as the first metadata block of a FLAC stream.
//
// ref: https://www.xiph.org/flac/format.html#metadata_block_streaminfo
type StreamInfo struct {
	// Minimum block size (in samples) used in the stream; between 16 and 65535
	// samples.
	BlockSizeMin uint16
	// Maximum block size (in samples) used in the stream; between 16 and 65535
	// samples.
	BlockSizeMax uint16
	// Minimum frame size in bytes; a 0 value implies unknown.
	FrameSizeMin uint32
	// Maximum frame size in bytes; a 0 value implies unknown.
	FrameSizeMax uint32
	// Sample rate in Hz; between 1 and 655350 Hz.
	SampleRate uint32
	// Number of channels; between 1 and 8 channels.
	NChannels uint8
	// Sample size in bits-per-sample; between 4 and 32 bits.
	BitsPerSample uint8
	// Total number of inter-channel samples in the stream. One second of 44.1
	// KHz audio will have 44100 samples regardless of the number of channels. A
	// 0 value implies unknown.
	NSamples uint64
	// MD5 checksum of the unencoded audio data.
	MD5sum [md5.Size]uint8
}

// parseStreamInfo reads and parses the body of a StreamInfo metadata block.
func (block *Block) parseStreamInfo() error {
	// 16 bits: BlockSizeMin.
	br := bits.NewReader(block.lr)
	x, err := br.Read(16)
	if err != nil {
		return unexpected(err)
	}
	if x < 16 {
		return fmt.Errorf("meta.Block.parseStreamInfo: invalid minimum block size (%d); expected >= 16", x)
	}
	si := new(StreamInfo)
	block.Body = si
	si.BlockSizeMin = uint16(x)

	// 16 bits: BlockSizeMax.
	x, err = br.Read(16)
	if err != nil {
		return unexpected(err)
	}
	if x < 16 {
		return fmt.Errorf("meta.Block.parseStreamInfo: invalid maximum block size (%d); expected >= 16", x)
	}
	si.BlockSizeMax = uint16(x)

	// 24 bits: FrameSizeMin.
	x, err = br.Read(24)
	if err != nil {
		return unexpected(err)
	}
	si.FrameSizeMin = uint32(x)

	// 24 bits: FrameSizeMax.
	x, err = br.Read(24)
	if err != nil {
		return unexpected(err)
	}
	si.FrameSizeMax = uint32(x)

	// 20 bits: SampleRate.
	x, err = br.Read(20)
	if err != nil {
		return unexpected(err)
	}
	if x == 0 {
		return errors.New("meta.Block.parseStreamInfo: invalid sample rate (0)")
	}
	si.SampleRate = uint32(x)

	// 3 bits: NChannels.
	x, err = br.Read(3)
	if err != nil {
		return unexpected(err)
	}
	// x contains: (number of channels) - 1
	si.NChannels = uint8(x + 1)

	// 5 bits: BitsPerSample.
	x, err = br.Read(5)
	if err != nil {
		return unexpected(err)
	}
	// x contains: (bits-per-sample) - 1
	si.BitsPerSample = uint8(x + 1)

	// 36 bits: NSamples.
	si.NSamples, err = br.Read(36)
	if err != nil {
		return unexpected(err)
	}

	// 16 bytes: MD5sum.
	_, err = io.ReadFull(block.lr, si.MD5sum[:])
	return unexpected(err)
}
