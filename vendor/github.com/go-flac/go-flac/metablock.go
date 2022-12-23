package flac

import (
	"bytes"
)

// BlockType representation of types of FLAC Metadata Block
type BlockType int

// BlockData data in a FLAC Metadata Block. Custom Metadata decoders and modifiers should accept/modify whole MetaDataBlock instead.
type BlockData []byte

const (
	// StreamInfo METADATA_BLOCK_STREAMINFO
	// This block has information about the whole stream, like sample rate, number of channels, total number of samples, etc. It must be present as the first metadata block in the stream. Other metadata blocks may follow, and ones that the decoder doesn't understand, it will skip.
	StreamInfo BlockType = iota
	// Padding METADATA_BLOCK_PADDING
	// This block allows for an arbitrary amount of padding. The contents of a PADDING block have no meaning. This block is useful when it is known that metadata will be edited after encoding; the user can instruct the encoder to reserve a PADDING block of sufficient size so that when metadata is added, it will simply overwrite the padding (which is relatively quick) instead of having to insert it into the right place in the existing file (which would normally require rewriting the entire file).
	Padding
	// Application METADATA_BLOCK_APPLICATION
	// This block is for use by third-party applications. The only mandatory field is a 32-bit identifier. This ID is granted upon request to an application by the FLAC maintainers. The remainder is of the block is defined by the registered application. Visit the registration page if you would like to register an ID for your application with FLAC.
	Application
	// SeekTable METADATA_BLOCK_SEEKTABLE
	// This is an optional block for storing seek points. It is possible to seek to any given sample in a FLAC stream without a seek table, but the delay can be unpredictable since the bitrate may vary widely within a stream. By adding seek points to a stream, this delay can be significantly reduced. Each seek point takes 18 bytes, so 1% resolution within a stream adds less than 2k. There can be only one SEEKTABLE in a stream, but the table can have any number of seek points. There is also a special 'placeholder' seekpoint which will be ignored by decoders but which can be used to reserve space for future seek point insertion.
	SeekTable
	// VorbisComment METADATA_BLOCK_VORBIS_COMMENT
	// This block is for storing a list of human-readable name/value pairs. Values are encoded using UTF-8. It is an implementation of the Vorbis comment specification (without the framing bit). This is the only officially supported tagging mechanism in FLAC. There may be only one VORBIS_COMMENT block in a stream. In some external documentation, Vorbis comments are called FLAC tags to lessen confusion.
	VorbisComment
	// CueSheet METADATA_BLOCK_CUESHEET
	// This block is for storing various information that can be used in a cue sheet. It supports track and index points, compatible with Red Book CD digital audio discs, as well as other CD-DA metadata such as media catalog number and track ISRCs. The CUESHEET block is especially useful for backing up CD-DA discs, but it can be used as a general purpose cueing mechanism for playback.
	CueSheet
	// Picture METADATA_BLOCK_PICTURE
	// This block is for storing pictures associated with the file, most commonly cover art from CDs. There may be more than one PICTURE block in a file. The picture format is similar to the APIC frame in ID3v2. The PICTURE block has a type, MIME type, and UTF-8 description like ID3v2, and supports external linking via URL (though this is discouraged). The differences are that there is no uniqueness constraint on the description field, and the MIME type is mandatory. The FLAC PICTURE block also includes the resolution, color depth, and palette size so that the client can search for a suitable picture without having to scan them all.
	Picture
	// Reserved Reserved Metadata Block Types
	Reserved
	// Invalid Invalid Metadata Block Type
	Invalid BlockType = 127
)

// MetaDataBlock is the struct representation of a FLAC Metadata Block
type MetaDataBlock struct {
	Type BlockType
	Data BlockData
}

// Marshal encodes this MetaDataBlock without touching block data
// isfinal defines whether this is the last metadata block of the FLAC file
func (c *MetaDataBlock) Marshal(isfinal bool) []byte {
	res := bytes.NewBuffer([]byte{})
	if isfinal {
		res.WriteByte(byte(c.Type + 1<<7))
	} else {
		res.WriteByte(byte(c.Type))
	}
	size := encodeUint32(uint32(len(c.Data)))
	res.Write(size[len(size)-3:])
	res.Write(c.Data)
	return res.Bytes()
}
