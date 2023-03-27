// OSC52 is a terminal escape sequence that allows copying text to the clipboard.
//
// The sequence consists of the following:
//
//	OSC 52 ; Pc ; Pd BEL
//
// Pc is the clipboard choice:
//
//	c: clipboard
//	p: primary
//	q: secondary (not supported)
//	s: select (not supported)
//	0-7: cut-buffers (not supported)
//
// Pd is the data to copy to the clipboard. This string should be encoded in
// base64 (RFC-4648).
//
// If Pd is "?", the terminal replies to the host with the current contents of
// the clipboard.
//
// If Pd is neither a base64 string nor "?", the terminal clears the clipboard.
//
// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
// where Ps = 52 => Manipulate Selection Data.
package osc52

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

// Clipboard is the clipboard buffer to use.
type Clipboard uint

const (
	// SystemClipboard is the system clipboard buffer.
	SystemClipboard Clipboard = iota
	// PrimaryClipboard is the primary clipboard buffer (X11).
	PrimaryClipboard
)

// String implements the fmt.Stringer interface for [Clipboard].
func (c Clipboard) String() string {
	return []string{
		"c", "p",
	}[c]
}

// output is the default output for Copy which uses os.Stdout and os.Environ.
var output = NewOutput(os.Stdout, os.Environ())

// envs is a map of environment variables.
type envs map[string]string

// Get returns the value of the environment variable named by the key.
func (e envs) Get(key string) string {
	v, ok := e[key]
	if !ok {
		return ""
	}
	return v
}

// Output is where the OSC52 string should be written.
type Output struct {
	out  io.Writer
	envs envs
}

// NewOutput returns a new Output.
func NewOutput(out io.Writer, envs []string) *Output {
	e := make(map[string]string, 0)
	for _, env := range envs {
		s := strings.Split(env, "=")
		k := s[0]
		v := strings.Join(s[1:], "=")
		e[k] = v
	}
	o := &Output{
		out:  out,
		envs: e,
	}
	return o
}

// DefaultOutput returns the default output for Copy.
func DefaultOutput() *Output {
	return output
}

// Copy copies the OSC52 string to the output. This uses the system clipboard buffer.
func Copy(str string) {
	output.Copy(str)
}

// Copy copies the OSC52 string to the output. This uses the system clipboard buffer.
func (o *Output) Copy(str string) {
	o.CopyClipboard(str, SystemClipboard)
}

// CopyPrimary copies the OSC52 string to the output. This uses the primary clipboard buffer.
func CopyPrimary(str string) {
	output.CopyPrimary(str)
}

// CopyPrimary copies the OSC52 string to the output. This uses the primary clipboard buffer.
func (o *Output) CopyPrimary(str string) {
	o.CopyClipboard(str, PrimaryClipboard)
}

// CopyClipboard copies the OSC52 string to the output. This uses the passed clipboard buffer.
func CopyClipboard(str string, c Clipboard) {
	output.CopyClipboard(str, c)
}

// CopyClipboard copies the OSC52 string to the output. This uses the passed clipboard buffer.
func (o *Output) CopyClipboard(str string, c Clipboard) {
	o.osc52Write(str, c)
}

func (o *Output) osc52Write(str string, c Clipboard) {
	var seq string
	term := strings.ToLower(o.envs.Get("TERM"))
	switch {
	case o.envs.Get("TMUX") != "", strings.HasPrefix(term, "tmux"):
		seq = Sequence(str, "tmux", c)
	case strings.HasPrefix(term, "screen"):
		seq = Sequence(str, "screen", c)
	case strings.Contains(term, "kitty"):
		// First, we flush the keyboard before copying, this is required for
		// Kitty < 0.22.0.
		o.out.Write([]byte(Clear(term, c)))
		seq = Sequence(str, "kitty", c)
	default:
		seq = Sequence(str, term, c)
	}
	o.out.Write([]byte(seq))
}

func seqStart(term string, c Clipboard) string {
	var seq strings.Builder
	switch {
	case strings.Contains(term, "tmux"):
		// Write the start of a tmux escape sequence.
		seq.WriteString("\x1bPtmux;\x1b")
	case strings.Contains(term, "screen"):
		// Write the start of a DCS sequence.
		seq.WriteString("\x1bP")
	}
	// OSC52 sequence start.
	seq.WriteString(fmt.Sprintf("\x1b]52;%s;", c))
	return seq.String()
}

func seqEnd(term string) string {
	var seq strings.Builder
	// OSC52 sequence end.
	seq.WriteString("\x07")
	switch {
	case strings.Contains(term, "tmux"):
		// Terminate the tmux escape sequence.
		seq.WriteString("\x1b\\")
	case strings.Contains(term, "screen"):
		// Write the end of a DCS sequence.
		seq.WriteString("\x1b\x5c")
	}
	return seq.String()
}

// sequence returns the OSC52 sequence for the passed content.
// Beware that the string here is not base64 encoded.
func sequence(contents string, term string, c Clipboard) string {
	var seq strings.Builder
	term = strings.ToLower(term)
	seq.WriteString(seqStart(term, c))
	switch {
	case strings.Contains(term, "screen"):
		// Screen doesn't support OSC52 but will pass the contents of a DCS sequence to
		// the outer terminal unchanged.
		//
		// Here, we split the encoded string into 76 bytes chunks and then join the
		// chunks with <end-dsc><start-dsc> sequences. Finally, wrap the whole thing in
		// <start-dsc><start-osc52><joined-chunks><end-osc52><end-dsc>.
		s := strings.SplitN(contents, "", 76)
		seq.WriteString(strings.Join(s, "\x1b\\\x1bP"))
	default:
		seq.WriteString(contents)
	}
	seq.WriteString(seqEnd(term))
	return seq.String()
}

// Sequence returns the OSC52 sequence for the given string, terminal, and clipboard choice.
func Sequence(str string, term string, c Clipboard) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(str))
	return sequence(b64, term, c)
}

// Contents returns the contents of the clipboard.
func Contents(term string, c Clipboard) string {
	var seq strings.Builder
	seq.WriteString(seqStart(term, c))
	seq.WriteString("?")
	seq.WriteString(seqEnd(term))
	return seq.String()
}

// Clear returns the OSC52 sequence to clear the clipboard.
func Clear(term string, c Clipboard) string {
	var seq strings.Builder
	seq.WriteString(seqStart(term, c))
	// Clear the clipboard
	seq.WriteString("!")
	seq.WriteString(seqEnd(term))
	return seq.String()
}
