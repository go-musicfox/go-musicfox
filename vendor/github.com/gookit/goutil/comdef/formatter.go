package comdef

import (
	"bytes"
	"io"

	"github.com/gookit/goutil/stdio"
)

// DataFormatter interface
type DataFormatter interface {
	Format() string
	FormatTo(w io.Writer)
}

// BaseFormatter struct
type BaseFormatter struct {
	ow ByteStringWriter
	// Out formatted to the writer
	Out io.Writer
	// Src data(array, map, struct) for format
	Src any
	// MaxDepth limit depth for array, map data TODO
	MaxDepth int
	// Prefix string for each element
	Prefix string
	// Indent string for format each element
	Indent string
	// ClosePrefix string for last "]", "}"
	ClosePrefix string
}

// Reset after format
func (f *BaseFormatter) Reset() {
	f.Out = nil
	f.Src = nil
}

// SetOutput writer
func (f *BaseFormatter) SetOutput(out io.Writer) {
	f.Out = out
}

// BsWriter build and get
func (f *BaseFormatter) BsWriter() ByteStringWriter {
	if f.ow == nil {
		if f.Out == nil {
			f.ow = new(bytes.Buffer)
		} else if ow, ok := f.Out.(ByteStringWriter); ok {
			f.ow = ow
		} else {
			f.ow = stdio.NewWriteWrapper(f.Out)
		}
	}

	return f.ow
}
