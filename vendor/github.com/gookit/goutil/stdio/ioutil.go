package stdio

import (
	"fmt"
	"io"
	"strings"
)

// QuietFprint to writer, will ignore error
func QuietFprint(w io.Writer, ss ...string) {
	_, _ = fmt.Fprint(w, strings.Join(ss, ""))
}

// QuietFprintf to writer, will ignore error
func QuietFprintf(w io.Writer, tpl string, vs ...any) {
	_, _ = fmt.Fprintf(w, tpl, vs...)
}

// QuietFprintln to writer, will ignore error
func QuietFprintln(w io.Writer, ss ...string) {
	_, _ = fmt.Fprintln(w, strings.Join(ss, ""))
}

// QuietWriteString to writer, will ignore error
func QuietWriteString(w io.Writer, ss ...string) {
	_, _ = io.WriteString(w, strings.Join(ss, ""))
}

// DiscardReader anything from the reader
func DiscardReader(src io.Reader) {
	_, _ = io.Copy(io.Discard, src)
}

// MustReadReader read contents from io.Reader, will panic on error
func MustReadReader(r io.Reader) []byte {
	bs, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return bs
}
