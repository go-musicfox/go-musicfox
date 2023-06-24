package tea

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi/compressor"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/termenv"
)

func ReplaceWithFoxfulRenderer(p *Program) *Program {
	p.renderer = newFoxfulRenderer(p.output, p.startupOptions.has(withANSICompressor))
	return p
}

// foxfulRenderer Customized render for musicfox
type foxfulRenderer struct {
	standardRenderer
}

func newFoxfulRenderer(out *termenv.Output, useANSICompressor bool) renderer {
	r := &foxfulRenderer{
		standardRenderer: standardRenderer{
			out:                out,
			mtx:                &sync.Mutex{},
			done:               make(chan struct{}),
			framerate:          time.Second / time.Duration(defaultFPS),
			useANSICompressor:  useANSICompressor,
			queuedMessageLines: []string{},
		},
	}
	if r.useANSICompressor {
		r.out = termenv.NewOutput(&compressor.Writer{Forward: out})
	}
	return r
}

func (r *foxfulRenderer) start() {
}

func (r *foxfulRenderer) stop() {
}

func (r *foxfulRenderer) kill() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.out.ClearLine()
}

func (r *foxfulRenderer) listen() {
}

func (r *foxfulRenderer) flush() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.buf.Len() == 0 || r.buf.String() == r.lastRender {
		// Nothing to do
		return
	}

	// Output buffer
	buf := &bytes.Buffer{}
	out := termenv.NewOutput(buf)

	newLines := strings.Split(r.buf.String(), "\n")

	// If we know the output's height, we can use it to determine how many
	// lines we can render. We drop lines from the top of the render buffer if
	// necessary, as we can't navigate the cursor into the terminal's scrollback
	// buffer.
	if r.height > 0 && len(newLines) > r.height {
		newLines = newLines[len(newLines)-r.height:]
	}

	numLinesThisFlush := len(newLines)
	oldLines := strings.Split(r.lastRender, "\n")
	skipLines := make(map[int]struct{})
	flushQueuedMessages := len(r.queuedMessageLines) > 0 && !r.altScreenActive

	// Add any queued messages to this render
	if flushQueuedMessages {
		newLines = append(r.queuedMessageLines, newLines...)
		r.queuedMessageLines = []string{}
	}

	// Clear any lines we painted in the last render.
	if r.linesRendered > 0 {
		for i := r.linesRendered - 1; i > 0; i-- {
			// If the number of lines we want to render hasn't increased and
			// new line is the same as the old line we can skip rendering for
			// this line as a performance optimization.
			if (len(newLines) <= len(oldLines)) && (len(newLines) > i && len(oldLines) > i) && (newLines[i] == oldLines[i]) {
				skipLines[i] = struct{}{}
			} else if _, exists := r.ignoreLines[i]; !exists {
				//out.ClearLine()
			}

			out.CursorUp(1)
		}

		if _, exists := r.ignoreLines[0]; !exists {
			// We need to return to the start of the line here to properly
			// erase it. Going back the entire width of the terminal will
			// usually be farther than we need to go, but terminal emulators
			// will stop the cursor at the start of the line as a rule.
			//
			// We use this sequence in particular because it's part of the ANSI
			// standard (whereas others are proprietary to, say, VT100/VT52).
			// If cursor previous line (ESC[ + <n> + F) were better supported
			// we could use that above to eliminate this step.
			out.CursorBack(r.width)
			//out.ClearLine()
		}
	}

	// Merge the set of lines we're skipping as a rendering optimization with
	// the set of lines we've explicitly asked the renderer to ignore.
	if r.ignoreLines != nil {
		for k, v := range r.ignoreLines {
			skipLines[k] = v
		}
	}

	// Paint new lines
	for i := 0; i < len(newLines); i++ {
		if _, skip := skipLines[i]; skip {
			// Unless this is the last line, move the cursor down.
			if i < len(newLines)-1 {
				out.CursorDown(1)
			}
		} else {
			line := newLines[i]

			// Truncate lines wider than the width of the window to avoid
			// wrapping, which will mess up rendering. If we don't have the
			// width of the window this will be ignored.
			//
			// Note that on Windows we only get the width of the window on
			// program initialization, so after a resize this won't perform
			// correctly (signal SIGWINCH is not supported on Windows).
			if r.width > 0 {
				line = truncate.String(line, uint(r.width))
			}

			if spaceNum := r.width - runewidth.StringWidth(line); spaceNum > 0 {
				line += strings.Repeat(" ", spaceNum)
			}

			_, _ = out.WriteString(line)

			if i < len(newLines)-1 {
				_, _ = out.WriteString("\r\n")
			}
		}
	}
	r.linesRendered = numLinesThisFlush

	// Make sure the cursor is at the start of the last line to keep rendering
	// behavior consistent.
	if r.altScreenActive {
		// This case fixes a bug in macOS terminal. In other terminals the
		// other case seems to do the job regardless of whether or not we're
		// using the full terminal window.
		out.MoveCursor(r.linesRendered, 0)
	} else {
		out.CursorBack(r.width)
	}

	_, _ = r.out.Write(buf.Bytes())
	r.lastRender = r.buf.String()
	r.buf.Reset()
}

func (r *foxfulRenderer) write(s string) {
	r.mtx.Lock()
	r.buf.Reset()

	if s == "" {
		s = " "
	}

	_, _ = r.buf.WriteString(s)
	r.mtx.Unlock()

	r.flush()
}
