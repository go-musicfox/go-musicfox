// Package frame implements access to FLAC audio frames.
//
// A brief introduction of the FLAC audio format [1] follows. FLAC encoders
// divide the audio stream into blocks through a process called blocking [2]. A
// block contains the unencoded audio samples from all channels during a short
// period of time. Each audio block is divided into subblocks, one per channel.
//
// There is often a correlation between the left and right channel of stereo
// audio. Using inter-channel decorrelation [3] it is possible to store only one
// of the channels and the difference between the channels, or store the average
// of the channels and their difference. An encoder decorrelates audio samples
// as follows:
//    mid = (left + right)/2 // average of the channels
//    side = left - right    // difference between the channels
//
// The blocks are encoded using a variety of prediction methods [4][5] and
// stored in frames. Blocks and subblocks contains unencoded audio samples while
// frames and subframes contain encoded audio samples. A FLAC stream contains
// one or more audio frames.
//
//    [1]: https://www.xiph.org/flac/format.html#architecture
//    [2]: https://www.xiph.org/flac/format.html#blocking
//    [3]: https://www.xiph.org/flac/format.html#interchannel
//    [4]: https://www.xiph.org/flac/format.html#prediction
//    [5]: https://godoc.org/github.com/mewkiz/flac/frame#Pred
package frame

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"

	"github.com/mewkiz/flac/internal/bits"
	"github.com/mewkiz/flac/internal/hashutil"
	"github.com/mewkiz/flac/internal/hashutil/crc16"
	"github.com/mewkiz/flac/internal/hashutil/crc8"
	"github.com/mewkiz/flac/internal/utf8"
)

// A Frame contains the header and subframes of an audio frame. It holds the
// encoded samples from a block (a part) of the audio stream. Each subframe
// holding the samples from one of its channel.
//
// ref: https://www.xiph.org/flac/format.html#frame
type Frame struct {
	// Audio frame header.
	Header
	// One subframe per channel, containing encoded audio samples.
	Subframes []*Subframe
	// CRC-16 hash sum, calculated by read operations on hr.
	crc hashutil.Hash16
	// A bit reader, wrapping read operations to hr.
	br *bits.Reader
	// A CRC-16 hash reader, wrapping read operations to r.
	hr io.Reader
	// Underlying io.Reader.
	r io.Reader
}

// New creates a new Frame for accessing the audio samples of r. It reads and
// parses an audio frame header. It returns io.EOF to signal a graceful end of
// FLAC stream.
//
// Call Frame.Parse to parse the audio samples of its subframes.
func New(r io.Reader) (frame *Frame, err error) {
	// Create a new CRC-16 hash reader which adds the data from all read
	// operations to a running hash.
	crc := crc16.NewIBM()
	hr := io.TeeReader(r, crc)

	// Parse frame header.
	frame = &Frame{crc: crc, hr: hr, r: r}
	err = frame.parseHeader()
	return frame, err
}

// Parse reads and parses the header, and the audio samples from each subframe
// of a frame. If the samples are inter-channel decorrelated between the
// subframes, it correlates them. It returns io.EOF to signal a graceful end of
// FLAC stream.
//
// ref: https://www.xiph.org/flac/format.html#interchannel
func Parse(r io.Reader) (frame *Frame, err error) {
	// Parse frame header.
	frame, err = New(r)
	if err != nil {
		return frame, err
	}

	// Parse subframes.
	err = frame.Parse()
	return frame, err
}

// Parse reads and parses the audio samples from each subframe of the frame. If
// the samples are inter-channel decorrelated between the subframes, it
// correlates them.
//
// ref: https://www.xiph.org/flac/format.html#interchannel
func (frame *Frame) Parse() error {
	// Parse subframes.
	frame.Subframes = make([]*Subframe, frame.Channels.Count())
	var err error
	for channel := range frame.Subframes {
		// The side channel requires an extra bit per sample when using
		// inter-channel decorrelation.
		bps := uint(frame.BitsPerSample)
		switch frame.Channels {
		case ChannelsSideRight:
			// channel 0 is the side channel.
			if channel == 0 {
				bps++
			}
		case ChannelsLeftSide, ChannelsMidSide:
			// channel 1 is the side channel.
			if channel == 1 {
				bps++
			}
		}

		// Parse subframe.
		frame.Subframes[channel], err = frame.parseSubframe(frame.br, bps)
		if err != nil {
			return err
		}
	}

	// Inter-channel correlation of subframe samples.
	frame.correlate()

	// 2 bytes: CRC-16 checksum.
	var want uint16
	if err = binary.Read(frame.r, binary.BigEndian, &want); err != nil {
		return unexpected(err)
	}
	got := frame.crc.Sum16()
	if got != want {
		return fmt.Errorf("frame.Frame.Parse: CRC-16 checksum mismatch; expected 0x%04X, got 0x%04X", want, got)
	}

	return nil
}

// Hash adds the decoded audio samples of the frame to a running MD5 hash. It
// can be used in conjunction with StreamInfo.MD5sum to verify the integrity of
// the decoded audio samples.
//
// Note: The audio samples of the frame must be decoded before calling Hash.
func (frame *Frame) Hash(md5sum hash.Hash) {
	// Write decoded samples to a running MD5 hash.
	bps := frame.BitsPerSample
	var buf [3]byte
	for i := 0; i < int(frame.BlockSize); i++ {
		for _, subframe := range frame.Subframes {
			sample := subframe.Samples[i]
			switch bps {
			case 8:
				buf[0] = uint8(sample)
				md5sum.Write(buf[:1])
			case 16:
				buf[0] = uint8(sample)
				buf[1] = uint8(sample >> 8)
				md5sum.Write(buf[:2])
			case 24:
				buf[0] = uint8(sample)
				buf[1] = uint8(sample >> 8)
				buf[2] = uint8(sample >> 16)
				md5sum.Write(buf[:])
			default:
				log.Printf("frame.Frame.Hash: support for %d-bit sample size not yet implemented", bps)
			}
		}
	}
}

// A Header contains the basic properties of an audio frame, such as its sample
// rate and channel count. To facilitate random access decoding each frame
// header starts with a sync-code. This allows the decoder to synchronize and
// locate the start of a frame header.
//
// ref: https://www.xiph.org/flac/format.html#frame_header
type Header struct {
	// Specifies if the block size is fixed or variable.
	HasFixedBlockSize bool
	// Block size in inter-channel samples, i.e. the number of audio samples in
	// each subframe.
	BlockSize uint16
	// Sample rate in Hz; a 0 value implies unknown, get sample rate from
	// StreamInfo.
	SampleRate uint32
	// Specifies the number of channels (subframes) that exist in the frame,
	// their order and possible inter-channel decorrelation.
	Channels Channels
	// Sample size in bits-per-sample; a 0 value implies unknown, get sample size
	// from StreamInfo.
	BitsPerSample uint8
	// Specifies the frame number if the block size is fixed, and the first
	// sample number in the frame otherwise. When using fixed block size, the
	// first sample number in the frame can be derived by multiplying the frame
	// number with the block size (in samples).
	Num uint64
}

// Errors returned by Frame.parseHeader.
var (
	ErrInvalidSync = errors.New("frame.Frame.parseHeader: invalid sync-code")
)

// parseHeader reads and parses the header of an audio frame.
func (frame *Frame) parseHeader() error {
	// Create a new CRC-8 hash reader which adds the data from all read
	// operations to a running hash.
	h := crc8.NewATM()
	hr := io.TeeReader(frame.hr, h)

	// Create bit reader.
	br := bits.NewReader(hr)
	frame.br = br

	// 14 bits: sync-code (11111111111110)
	x, err := br.Read(14)
	if err != nil {
		// This is the only place an audio frame may return io.EOF, which signals
		// a graceful end of a FLAC stream.
		return err
	}
	if x != 0x3FFE {
		return ErrInvalidSync
	}

	// 1 bit: reserved.
	x, err = br.Read(1)
	if err != nil {
		return unexpected(err)
	}
	if x != 0 {
		return errors.New("frame.Frame.parseHeader: non-zero reserved value")
	}

	// 1 bit: HasFixedBlockSize.
	x, err = br.Read(1)
	if err != nil {
		return unexpected(err)
	}
	if x == 0 {
		frame.HasFixedBlockSize = true
	}

	// 4 bits: BlockSize. The block size parsing is simplified by deferring it to
	// the end of the header.
	blockSize, err := br.Read(4)
	if err != nil {
		return unexpected(err)
	}

	// 4 bits: SampleRate. The sample rate parsing is simplified by deferring it
	// to the end of the header.
	sampleRate, err := br.Read(4)
	if err != nil {
		return unexpected(err)
	}

	// Parse channels.
	if err := frame.parseChannels(br); err != nil {
		return err
	}

	// Parse bits per sample.
	if err := frame.parseBitsPerSample(br); err != nil {
		return err
	}

	// 1 bit: reserved.
	x, err = br.Read(1)
	if err != nil {
		return unexpected(err)
	}
	if x != 0 {
		return errors.New("frame.Frame.parseHeader: non-zero reserved value")
	}

	// if (fixed block size)
	//    1-6 bytes: UTF-8 encoded frame number.
	// else
	//    1-7 bytes: UTF-8 encoded sample number.
	frame.Num, err = utf8.Decode(hr)
	if err != nil {
		return unexpected(err)
	}

	// Parse block size.
	if err := frame.parseBlockSize(br, blockSize); err != nil {
		return err
	}

	// Parse sample rate.
	if err := frame.parseSampleRate(br, sampleRate); err != nil {
		return err
	}

	// 1 byte: CRC-8 checksum.
	var want uint8
	if err = binary.Read(frame.hr, binary.BigEndian, &want); err != nil {
		return unexpected(err)
	}
	got := h.Sum8()
	if want != got {
		return fmt.Errorf("frame.Frame.parseHeader: CRC-8 checksum mismatch; expected 0x%02X, got 0x%02X", want, got)
	}

	return nil
}

// parseBitsPerSample parses the bits per sample of the header.
func (frame *Frame) parseBitsPerSample(br *bits.Reader) error {
	// 3 bits: BitsPerSample.
	x, err := br.Read(3)
	if err != nil {
		return unexpected(err)
	}

	// The 3 bits are used to specify the sample size as follows:
	//    000: unknown sample size; get from StreamInfo.
	//    001: 8 bits-per-sample.
	//    010: 12 bits-per-sample.
	//    011: reserved.
	//    100: 16 bits-per-sample.
	//    101: 20 bits-per-sample.
	//    110: 24 bits-per-sample.
	//    111: reserved.
	switch x {
	case 0x0:
		// 000: unknown bits-per-sample; get from StreamInfo.
	case 0x1:
		// 001: 8 bits-per-sample.
		frame.BitsPerSample = 8
	case 0x2:
		// 010: 12 bits-per-sample.
		frame.BitsPerSample = 12
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with %d bits-per-sample. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.BitsPerSample)
	case 0x4:
		// 100: 16 bits-per-sample.
		frame.BitsPerSample = 16
	case 0x5:
		// 101: 20 bits-per-sample.
		frame.BitsPerSample = 20
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with %d bits-per-sample. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.BitsPerSample)
	case 0x6:
		// 110: 24 bits-per-sample.
		frame.BitsPerSample = 24
	default:
		// 011: reserved.
		// 111: reserved.
		return fmt.Errorf("frame.Frame.parseHeader: reserved sample size bit pattern (%03b)", x)
	}
	return nil
}

// parseChannels parses the channels of the header.
func (frame *Frame) parseChannels(br *bits.Reader) error {
	// 4 bits: Channels.
	//
	// The 4 bits are used to specify the channels as follows:
	//    0000: (1 channel) mono.
	//    0001: (2 channels) left, right.
	//    0010: (3 channels) left, right, center.
	//    0011: (4 channels) left, right, left surround, right surround.
	//    0100: (5 channels) left, right, center, left surround, right surround.
	//    0101: (6 channels) left, right, center, LFE, left surround, right surround.
	//    0110: (7 channels) left, right, center, LFE, center surround, side left, side right.
	//    0111: (8 channels) left, right, center, LFE, left surround, right surround, side left, side right.
	//    1000: (2 channels) left, side; using inter-channel decorrelation.
	//    1001: (2 channels) side, right; using inter-channel decorrelation.
	//    1010: (2 channels) mid, side; using inter-channel decorrelation.
	//    1011: reserved.
	//    1100: reserved.
	//    1101: reserved.
	//    1111: reserved.
	x, err := br.Read(4)
	if err != nil {
		return unexpected(err)
	}
	if x >= 0xB {
		return fmt.Errorf("frame.Frame.parseHeader: reserved channels bit pattern (%04b)", x)
	}
	frame.Channels = Channels(x)
	return nil
}

// parseBlockSize parses the block size of the header.
func (frame *Frame) parseBlockSize(br *bits.Reader, blockSize uint64) error {
	// The 4 bits of n are used to specify the block size as follows:
	//    0000: reserved.
	//    0001: 192 samples.
	//    0010-0101: 576 * 2^(n-2) samples.
	//    0110: get 8 bit (block size)-1 from the end of the header.
	//    0111: get 16 bit (block size)-1 from the end of the header.
	//    1000-1111: 256 * 2^(n-8) samples.
	n := blockSize
	switch {
	case n == 0x0:
		// 0000: reserved.
		return errors.New("frame.Frame.parseHeader: reserved block size bit pattern (0000)")
	case n == 0x1:
		// 0001: 192 samples.
		frame.BlockSize = 192
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with block size %d. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.BlockSize)
	case n >= 0x2 && n <= 0x5:
		// 0010-0101: 576 * 2^(n-2) samples.
		frame.BlockSize = 576 * (1 << (n - 2))
	case n == 0x6:
		// 0110: get 8 bit (block size)-1 from the end of the header.
		x, err := br.Read(8)
		if err != nil {
			return unexpected(err)
		}
		frame.BlockSize = uint16(x + 1)
	case n == 0x7:
		// 0111: get 16 bit (block size)-1 from the end of the header.
		x, err := br.Read(16)
		if err != nil {
			return unexpected(err)
		}
		frame.BlockSize = uint16(x + 1)
	default:
		//    1000-1111: 256 * 2^(n-8) samples.
		frame.BlockSize = 256 * (1 << (n - 8))
	}
	return nil
}

// parseSampleRate parses the sample rate of the header.
func (frame *Frame) parseSampleRate(br *bits.Reader, sampleRate uint64) error {
	// The 4 bits are used to specify the sample rate as follows:
	//    0000: unknown sample rate; get from StreamInfo.
	//    0001: 88.2 kHz.
	//    0010: 176.4 kHz.
	//    0011: 192 kHz.
	//    0100: 8 kHz.
	//    0101: 16 kHz.
	//    0110: 22.05 kHz.
	//    0111: 24 kHz.
	//    1000: 32 kHz.
	//    1001: 44.1 kHz.
	//    1010: 48 kHz.
	//    1011: 96 kHz.
	//    1100: get 8 bit sample rate (in kHz) from the end of the header.
	//    1101: get 16 bit sample rate (in Hz) from the end of the header.
	//    1110: get 16 bit sample rate (in daHz) from the end of the header.
	//    1111: invalid.
	switch sampleRate {
	case 0x0:
		// 0000: unknown sample rate; get from StreamInfo.
	case 0x1:
		// 0001: 88.2 kHz.
		frame.SampleRate = 88200
	case 0x2:
		// 0010: 176.4 kHz.
		frame.SampleRate = 176400
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with sample rate %d. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.SampleRate)
	case 0x3:
		// 0011: 192 kHz.
		frame.SampleRate = 192000
	case 0x4:
		// 0100: 8 kHz.
		frame.SampleRate = 8000
	case 0x5:
		// 0101: 16 kHz.
		frame.SampleRate = 16000
	case 0x6:
		// 0110: 22.05 kHz.
		frame.SampleRate = 22050
	case 0x7:
		// 0111: 24 kHz.
		frame.SampleRate = 24000
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with sample rate %d. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.SampleRate)
	case 0x8:
		// 1000: 32 kHz.
		frame.SampleRate = 32000
	case 0x9:
		// 1001: 44.1 kHz.
		frame.SampleRate = 44100
	case 0xA:
		// 1010: 48 kHz.
		frame.SampleRate = 48000
	case 0xB:
		// 1011: 96 kHz.
		frame.SampleRate = 96000
	case 0xC:
		// 1100: get 8 bit sample rate (in kHz) from the end of the header.
		x, err := br.Read(8)
		if err != nil {
			return unexpected(err)
		}
		frame.SampleRate = uint32(x * 1000)
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with sample rate %d. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.SampleRate)
	case 0xD:
		// 1101: get 16 bit sample rate (in Hz) from the end of the header.
		x, err := br.Read(16)
		if err != nil {
			return unexpected(err)
		}
		frame.SampleRate = uint32(x)
	case 0xE:
		// 1110: get 16 bit sample rate (in daHz) from the end of the header.
		x, err := br.Read(16)
		if err != nil {
			return unexpected(err)
		}
		frame.SampleRate = uint32(x * 10)
		// TODO(u): Remove log message when the test cases have been extended.
		log.Printf("frame.Frame.parseHeader: The flac library test cases do not yet include any audio files with sample rate %d. If possible please consider contributing this audio sample to improve the reliability of the test cases.", frame.SampleRate)
	default:
		// 1111: invalid.
		return errors.New("frame.Frame.parseHeader: invalid sample rate bit pattern (1111)")
	}
	return nil
}

// Channels specifies the number of channels (subframes) that exist in a frame,
// their order and possible inter-channel decorrelation.
type Channels uint8

// Channel assignments. The following abbreviations are used:
//    C:   center (directly in front)
//    R:   right (standard stereo)
//    Sr:  side right (directly to the right)
//    Rs:  right surround (back right)
//    Cs:  center surround (rear center)
//    Ls:  left surround (back left)
//    Sl:  side left (directly to the left)
//    L:   left (standard stereo)
//    Lfe: low-frequency effect (placed according to room acoustics)
//
// The first 6 channel constants follow the SMPTE/ITU-R channel order:
//    L R C Lfe Ls Rs
const (
	ChannelsMono           Channels = iota // 1 channel: mono.
	ChannelsLR                             // 2 channels: left, right.
	ChannelsLRC                            // 3 channels: left, right, center.
	ChannelsLRLsRs                         // 4 channels: left, right, left surround, right surround.
	ChannelsLRCLsRs                        // 5 channels: left, right, center, left surround, right surround.
	ChannelsLRCLfeLsRs                     // 6 channels: left, right, center, LFE, left surround, right surround.
	ChannelsLRCLfeCsSlSr                   // 7 channels: left, right, center, LFE, center surround, side left, side right.
	ChannelsLRCLfeLsRsSlSr                 // 8 channels: left, right, center, LFE, left surround, right surround, side left, side right.
	ChannelsLeftSide                       // 2 channels: left, side; using inter-channel decorrelation.
	ChannelsSideRight                      // 2 channels: side, right; using inter-channel decorrelation.
	ChannelsMidSide                        // 2 channels: mid, side; using inter-channel decorrelation.
)

// nChannels specifies the number of channels used by each channel assignment.
var nChannels = [...]int{
	ChannelsMono:           1,
	ChannelsLR:             2,
	ChannelsLRC:            3,
	ChannelsLRLsRs:         4,
	ChannelsLRCLsRs:        5,
	ChannelsLRCLfeLsRs:     6,
	ChannelsLRCLfeCsSlSr:   7,
	ChannelsLRCLfeLsRsSlSr: 8,
	ChannelsLeftSide:       2,
	ChannelsSideRight:      2,
	ChannelsMidSide:        2,
}

// Count returns the number of channels (subframes) used by the provided channel
// assignment.
func (channels Channels) Count() int {
	return nChannels[channels]
}

// correlate reverts any inter-channel decorrelation between the samples of the
// subframes.
//
// An encoder decorrelates audio samples as follows:
//    mid = (left + right)/2
//    side = left - right
func (frame *Frame) correlate() {
	switch frame.Channels {
	case ChannelsLeftSide:
		// 2 channels: left, side; using inter-channel decorrelation.
		left := frame.Subframes[0].Samples
		side := frame.Subframes[1].Samples
		for i := range side {
			// right = left - side
			side[i] = left[i] - side[i]
		}
	case ChannelsSideRight:
		// 2 channels: side, right; using inter-channel decorrelation.
		side := frame.Subframes[0].Samples
		right := frame.Subframes[1].Samples
		// left = right + side
		for i := range side {
			side[i] += right[i]
		}
	case ChannelsMidSide:
		// 2 channels: mid, side; using inter-channel decorrelation.
		mid := frame.Subframes[0].Samples
		side := frame.Subframes[1].Samples
		for i := range side {
			// left = (2*mid + side)/2
			// right = (2*mid - side)/2
			m := mid[i]
			s := side[i]
			m *= 2
			// Notice that the integer division in mid = (left + right)/2 discards
			// the least significant bit. It can be reconstructed however, since a
			// sum A+B and a difference A-B has the same least significant bit.
			//
			// ref: Data Compression: The Complete Reference (ch. 7, Decorrelation)
			m |= s & 1
			mid[i] = (m + s) / 2
			side[i] = (m - s) / 2
		}
	}
}

// SampleNumber returns the first sample number contained within the frame.
func (frame *Frame) SampleNumber() uint64 {
	if frame.HasFixedBlockSize {
		return frame.Num * uint64(frame.BlockSize)
	}
	return frame.Num
}

// unexpected returns io.ErrUnexpectedEOF if err is io.EOF, and returns err
// otherwise.
func unexpected(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}
