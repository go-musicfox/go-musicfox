package flac

import (
	"bytes"
	"io"
)

// StreamInfoBlock represents the undecoded data of StreamInfo block
type StreamInfoBlock struct {
	// BlockSizeMin The minimum block size (in samples) used in the stream.
	BlockSizeMin int
	// BlockSizeMax The maximum block size (in samples) used in the stream. (Minimum blocksize == maximum blocksize) implies a fixed-blocksize stream.
	BlockSizeMax int
	// FrameSizeMin The minimum frame size (in bytes) used in the stream. May be 0 to imply the value is not known.
	FrameSizeMin int
	// FrameSizeMax The maximum frame size (in bytes) used in the stream. May be 0 to imply the value is not known.
	FrameSizeMax int
	// SampleRate Sample rate in Hz
	SampleRate int
	// ChannelCount Number of channels
	ChannelCount int
	// BitDepth  Bits per sample
	BitDepth int
	// SampleCount Total samples in stream.  'Samples' means inter-channel sample, i.e. one second of 44.1Khz audio will have 44100 samples regardless of the number of channels. A value of zero here means the number of total samples is unknown.
	SampleCount int64
	// AudioMD5 MD5 signature of the unencoded audio data
	AudioMD5 []byte
}

// GetStreamInfo parses the first metadata block of the File which should always be StreamInfo and returns a StreamInfoBlock containing the decoded StreamInfo data.
func (c *File) GetStreamInfo() (*StreamInfoBlock, error) {
	if c.Meta[0].Type != StreamInfo {
		return nil, ErrorNoStreamInfo
	}
	streamInfo := bytes.NewReader(c.Meta[0].Data)
	res := StreamInfoBlock{}

	if buf, err := readUint16(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.BlockSizeMin = int(buf)
	}

	if buf, err := readUint16(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.BlockSizeMax = int(buf)
	}

	buf := bytes.NewBuffer([]byte{0})
	buf24 := make([]byte, 3)
	if _, err := io.ReadFull(streamInfo, buf24); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}
	buf.Write(buf24)
	if buf, err := readUint32(buf); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.FrameSizeMin = int(buf)
	}
	buf.Reset()
	buf.WriteByte(0)
	if _, err := io.ReadFull(streamInfo, buf24); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}
	buf.Write(buf24)
	if buf, err := readUint32(buf); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.FrameSizeMax = int(buf)
	}

	buf.Reset()
	buf.WriteByte(0)
	smpl := make([]byte, 3)
	if _, err := io.ReadFull(streamInfo, smpl); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}
	buf.Write(smpl)
	if smplrate, err := readUint32(buf); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.SampleRate = int(smplrate >> 4)
	}
	if _, err := streamInfo.Seek(-1, io.SeekCurrent); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	if channel, err := readUint8(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.ChannelCount = int(channel<<4>>5) + 1
	}
	buf.Reset()
	if _, err := streamInfo.Seek(-1, io.SeekCurrent); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	if bitdepth, err := readUint16(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.BitDepth = int(bitdepth<<7>>11) + 1
	}
	if _, err := streamInfo.Seek(-1, io.SeekCurrent); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	var smplcount int64
	if count, err := readUint32(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		smplcount += int64(count<<4>>4) << 8
	}
	if count, err := readUint8(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		smplcount += int64(count)
	}
	res.SampleCount = smplcount

	res.AudioMD5 = make([]byte, 16)
	if _, err := io.ReadFull(streamInfo, res.AudioMD5); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	return &res, nil

}
