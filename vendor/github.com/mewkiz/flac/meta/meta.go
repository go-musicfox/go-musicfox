// Package meta implements access to FLAC metadata blocks.
//
// A brief introduction of the FLAC metadata format [1] follows. FLAC metadata
// is stored in blocks; each block contains a header followed by a body. The
// block header describes the type of the block body, its length in bytes, and
// specifies if the block was the last metadata block in a FLAC stream. The
// contents of the block body depends on the type specified in the block header.
//
// At the time of this writing, the FLAC metadata format defines seven different
// metadata block types, namely:
//    * StreamInfo [2]
//    * Padding [3]
//    * Application [4]
//    * SeekTable [5]
//    * VorbisComment [6]
//    * CueSheet [7]
//    * Picture [8]
// Please refer to their respective documentation for further information.
//
//    [1]: https://www.xiph.org/flac/format.html#format_overview
//    [2]: https://godoc.org/github.com/mewkiz/flac/meta#StreamInfo
//    [3]: https://www.xiph.org/flac/format.html#metadata_block_padding
//    [4]: https://godoc.org/github.com/mewkiz/flac/meta#Application
//    [5]: https://godoc.org/github.com/mewkiz/flac/meta#SeekTable
//    [6]: https://godoc.org/github.com/mewkiz/flac/meta#VorbisComment
//    [7]: https://godoc.org/github.com/mewkiz/flac/meta#CueSheet
//    [8]: https://godoc.org/github.com/mewkiz/flac/meta#Picture
package meta

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/mewkiz/flac/internal/bits"
)

// A Block contains the header and body of a metadata block.
//
// ref: https://www.xiph.org/flac/format.html#metadata_block
type Block struct {
	// Metadata block header.
	Header
	// Metadata block body of type *StreamInfo, *Application, ... etc. Body is
	// initially nil, and gets populated by a call to Block.Parse.
	Body interface{}
	// Underlying io.Reader; limited by the length of the block body.
	lr io.Reader
}

// New creates a new Block for accessing the metadata of r. It reads and parses
// a metadata block header.
//
// Call Block.Parse to parse the metadata block body, and call Block.Skip to
// ignore it.
func New(r io.Reader) (block *Block, err error) {
	block = new(Block)
	if err = block.parseHeader(r); err != nil {
		return block, err
	}
	block.lr = io.LimitReader(r, block.Length)
	return block, nil
}

// Parse reads and parses the header and body of a metadata block. Use New for
// additional granularity.
func Parse(r io.Reader) (block *Block, err error) {
	block, err = New(r)
	if err != nil {
		return block, err
	}
	if err = block.Parse(); err != nil {
		return block, err
	}
	return block, nil
}

// Errors returned by Parse.
var (
	ErrReservedType = errors.New("meta.Block.Parse: reserved block type")
	ErrInvalidType  = errors.New("meta.Block.Parse: invalid block type")
)

// Parse reads and parses the metadata block body.
func (block *Block) Parse() error {
	switch block.Type {
	case TypeStreamInfo:
		return block.parseStreamInfo()
	case TypePadding:
		return block.verifyPadding()
	case TypeApplication:
		return block.parseApplication()
	case TypeSeekTable:
		return block.parseSeekTable()
	case TypeVorbisComment:
		return block.parseVorbisComment()
	case TypeCueSheet:
		return block.parseCueSheet()
	case TypePicture:
		return block.parsePicture()
	}
	if block.Type >= 7 && block.Type <= 126 {
		return ErrReservedType
	}
	return ErrInvalidType
}

// Skip ignores the contents of the metadata block body.
func (block *Block) Skip() error {
	if sr, ok := block.lr.(io.Seeker); ok {
		_, err := sr.Seek(0, io.SeekEnd)
		return err
	}
	_, err := io.Copy(ioutil.Discard, block.lr)
	return err
}

// A Header contains information about the type and length of a metadata block.
//
// ref: https://www.xiph.org/flac/format.html#metadata_block_header
type Header struct {
	// Metadata block body type.
	Type Type
	// Length of body data in bytes.
	Length int64
	// IsLast specifies if the block is the last metadata block.
	IsLast bool
}

// parseHeader reads and parses the header of a metadata block.
func (block *Block) parseHeader(r io.Reader) error {
	// 1 bit: IsLast.
	br := bits.NewReader(r)
	x, err := br.Read(1)
	if err != nil {
		// This is the only place a metadata block may return io.EOF, which
		// signals a graceful end of a FLAC stream (from a metadata point of
		// view).
		//
		// Note that valid FLAC streams always contain at least one audio frame
		// after the last metadata block. Therefore an io.EOF error at this
		// location is always invalid. This logic is to be handled by the flac
		// package however.
		return err
	}
	if x != 0 {
		block.IsLast = true
	}

	// 7 bits: Type.
	x, err = br.Read(7)
	if err != nil {
		return unexpected(err)
	}
	block.Type = Type(x)

	// 24 bits: Length.
	x, err = br.Read(24)
	if err != nil {
		return unexpected(err)
	}
	block.Length = int64(x)

	return nil
}

// Type represents the type of a metadata block body.
type Type uint8

// Metadata block body types.
const (
	TypeStreamInfo    Type = 0
	TypePadding       Type = 1
	TypeApplication   Type = 2
	TypeSeekTable     Type = 3
	TypeVorbisComment Type = 4
	TypeCueSheet      Type = 5
	TypePicture       Type = 6
)

func (t Type) String() string {
	switch t {
	case TypeStreamInfo:
		return "stream info"
	case TypePadding:
		return "padding"
	case TypeApplication:
		return "application"
	case TypeSeekTable:
		return "seek table"
	case TypeVorbisComment:
		return "vorbis comment"
	case TypeCueSheet:
		return "cue sheet"
	case TypePicture:
		return "picture"
	default:
		return "<unknown block type>"
	}
}

// unexpected returns io.ErrUnexpectedEOF if err is io.EOF, and returns err
// otherwise.
func unexpected(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}
