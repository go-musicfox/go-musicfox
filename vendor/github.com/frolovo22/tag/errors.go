package tag

import "errors"

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrIncorrectLength   = errors.New("tag incorrect length")
	ErrUnsupportedTag    = errors.New("unsupported tag")
	ErrTagNotFound       = errors.New("tag not found")
	ErrEmptyFile         = errors.New("empty file")
	ErrFileMarker        = errors.New("error file marker")
	ErrReadFile          = errors.New("error read file")
	ErrSeekFile          = errors.New("error seek file")
	ErrWriteFile         = errors.New("error write file")
	ErrIncorrectTag      = errors.New("incorrect tag")
	ErrNotFoundGenre     = errors.New("not found genre")
	ErrNotPictureBlock   = errors.New("not a picture block")
	ErrWriting           = errors.New("writing error")
	ErrDecodeEvenLength  = errors.New("must have even length byte slice")
	ErrEncodingFormat    = errors.New("unknown encoding format")
)
