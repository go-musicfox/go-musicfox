package cmdline

import (
	"strings"
)

// LineBuilder build command line string.
// codes refer from strings.Builder
type LineBuilder struct {
	buf []byte
}

// NewBuilder create
func NewBuilder(binFile string, args ...string) *LineBuilder {
	b := &LineBuilder{}
	b.AddArg(binFile)
	b.AddArray(args)
	return b
}

// LineBuild build command line string by given args.
func LineBuild(binFile string, args []string) string {
	return NewBuilder(binFile, args...).String()
}

// AddArg to builder
func (b *LineBuilder) AddArg(arg string) {
	_, _ = b.WriteString(arg)
}

// AddArgs to builder
func (b *LineBuilder) AddArgs(args ...string) {
	b.AddArray(args)
}

// AddArray to builder
func (b *LineBuilder) AddArray(args []string) {
	for _, arg := range args {
		_, _ = b.WriteString(arg)
	}
}

// WriteString arg string to the builder, will auto quote special string.
// refer strconv.Quote()
func (b *LineBuilder) WriteString(a string) (int, error) {
	var quote byte
	if strings.ContainsRune(a, '"') {
		quote = '\''
	} else if a == "" || strings.ContainsRune(a, '\'') || strings.ContainsRune(a, ' ') {
		quote = '"'
	}

	// add sep on first write.
	if b.buf != nil {
		b.buf = append(b.buf, ' ')
	}

	// no quote char
	if quote == 0 {
		b.buf = append(b.buf, a...)
		return len(a) + 1, nil
	}

	b.buf = append(b.buf, quote) // add start quote
	b.buf = append(b.buf, a...)
	b.buf = append(b.buf, quote) // add end quote
	return len(a) + 3, nil
}

// String to command line string
func (b *LineBuilder) String() string {
	return string(b.buf)
}

// Len of the builder
func (b *LineBuilder) Len() int {
	return len(b.buf)
}

// Reset builder
func (b *LineBuilder) Reset() {
	b.buf = nil
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
// func (b *LineBuilder) grow(n int) {
// 	buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
// 	copy(buf, b.buf)
// 	b.buf = buf
// }
