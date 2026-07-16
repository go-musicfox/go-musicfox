package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// foxfulRenderer adapts the v0.25.0-foxful renderer to the v2 renderer
// interface. It redraws padded text lines directly instead of diffing a cell
// buffer, preserving the renderer strategy used by go-musicfox before v2.
type foxfulRenderer struct {
	mu sync.Mutex

	out io.Writer
	buf bytes.Buffer

	view       View
	lastView   *View
	lastRender string

	linesRendered   int
	width, height   int
	altScreenActive bool
}

var _ renderer = (*foxfulRenderer)(nil)

func newFoxfulRenderer(w io.Writer, width, height int) *foxfulRenderer {
	return &foxfulRenderer{
		out:    w,
		width:  width,
		height: height,
	}
}

func (r *foxfulRenderer) start() {}

func (r *foxfulRenderer) close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var buf bytes.Buffer
	if r.lastView != nil {
		if !r.lastView.DisableBracketedPasteMode {
			buf.WriteString(ansi.ResetModeBracketedPaste)
		}
		if r.lastView.MouseMode != MouseModeNone {
			buf.WriteString(ansi.ResetModeMouseButtonEvent +
				ansi.ResetModeMouseAnyEvent +
				ansi.ResetModeMouseExtSgr)
		}
	}
	if r.altScreenActive {
		buf.WriteString(ansi.ResetModeAltScreenSaveCursor)
		r.altScreenActive = false
	} else {
		buf.WriteString(ansi.EraseEntireScreen + ansi.CursorHomePosition)
	}
	buf.WriteString(ansi.ShowCursor)
	_, err := r.out.Write(buf.Bytes())
	return err
}

func (r *foxfulRenderer) render(view View) {
	r.mu.Lock()
	r.view = view
	r.buf.Reset()
	if view.Content == "" {
		r.buf.WriteByte(' ')
	} else {
		r.buf.WriteString(view.Content)
	}
	r.mu.Unlock()

	_ = r.flush(false)
}

func (r *foxfulRenderer) flush(bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.buf.Len() == 0 {
		return nil
	}

	view := r.view
	content := r.buf.String()
	if content == r.lastRender {
		r.buf.Reset()
		return nil
	}
	r.syncTerminalModes(view)

	newLines := strings.Split(content, "\n")
	if r.height > 0 && len(newLines) > r.height {
		newLines = newLines[len(newLines)-r.height:]
	}

	oldLines := strings.Split(r.lastRender, "\n")
	numLines := len(newLines)
	skipLines := make(map[int]struct{})

	var buf bytes.Buffer
	if r.linesRendered > 0 {
		for i := r.linesRendered - 1; i > 0; i-- {
			if len(newLines) <= len(oldLines) && len(newLines) > i && len(oldLines) > i && newLines[i] == oldLines[i] {
				skipLines[i] = struct{}{}
			}
			buf.WriteString(ansi.CursorUp(1))
		}
		buf.WriteString(ansi.CursorBackward(r.width))
	}

	for i, line := range newLines {
		if _, skip := skipLines[i]; skip {
			if i < len(newLines)-1 {
				buf.WriteString(ansi.CursorDown(1))
			}
			continue
		}
		if r.width > 0 {
			line = ansi.TruncateWc(line, r.width, "")
		}
		if padding := r.width - ansi.StringWidthWc(line); padding > 0 {
			line += strings.Repeat(" ", padding)
		}
		buf.WriteString(line)
		if i < len(newLines)-1 {
			buf.WriteString("\r\n")
		}
	}

	r.linesRendered = numLines
	if r.altScreenActive {
		buf.WriteString(fmt.Sprintf("\x1b[%d;1H", max(r.linesRendered, 1)))
	} else {
		buf.WriteString(ansi.CursorBackward(r.width))
	}

	if _, err := r.out.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("bubbletea: error flushing foxful renderer: %w", err)
	}
	r.lastRender = content
	r.lastView = &view
	r.buf.Reset()
	return r.setCursor(view.Cursor)
}

func (r *foxfulRenderer) reset() {
	r.mu.Lock()
	r.buf.Reset()
	r.lastRender = ""
	r.linesRendered = 0
	r.mu.Unlock()
}

func (r *foxfulRenderer) insertAbove(str string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.altScreenActive || str == "" {
		return nil
	}
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	_, err := io.WriteString(r.out, str)
	return err
}

func (r *foxfulRenderer) setSyncdUpdates(bool) {}

func (r *foxfulRenderer) setWidthMethod(ansi.Method) {}

func (r *foxfulRenderer) resize(width, height int) {
	r.mu.Lock()
	r.width, r.height = width, height
	r.buf.Reset()
	r.lastRender = ""
	r.linesRendered = 0
	r.mu.Unlock()
}

func (r *foxfulRenderer) setColorProfile(colorprofile.Profile) {}

func (r *foxfulRenderer) clearScreen() {
	r.mu.Lock()
	// v2 processes ClearScreen independently from the following page render.
	// Force that render to cover the previous frame without exposing a blank one.
	r.buf.Reset()
	r.lastRender = ""
	r.mu.Unlock()
}

func (r *foxfulRenderer) writeString(str string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return io.WriteString(r.out, str)
}

func (r *foxfulRenderer) onMouse(msg MouseMsg) Cmd {
	r.mu.Lock()
	var handler func(MouseMsg) Cmd
	if r.lastView != nil {
		handler = r.lastView.OnMouse
	}
	r.mu.Unlock()
	if handler == nil {
		return nil
	}
	return handler(msg)
}

func (r *foxfulRenderer) syncTerminalModes(view View) {
	var buf bytes.Buffer

	if r.lastView == nil || r.lastView.AltScreen != view.AltScreen {
		if view.AltScreen {
			buf.WriteString(ansi.SetModeAltScreenSaveCursor)
			buf.WriteString(ansi.EraseEntireScreen + ansi.CursorHomePosition)
		} else if r.altScreenActive {
			buf.WriteString(ansi.ResetModeAltScreenSaveCursor)
		}
		r.altScreenActive = view.AltScreen
		r.lastRender = ""
		r.linesRendered = 0
	}

	if r.lastView == nil || r.lastView.DisableBracketedPasteMode != view.DisableBracketedPasteMode {
		if view.DisableBracketedPasteMode {
			buf.WriteString(ansi.ResetModeBracketedPaste)
		} else {
			buf.WriteString(ansi.SetModeBracketedPaste)
		}
	}

	if r.lastView == nil || r.lastView.MouseMode != view.MouseMode {
		if r.lastView != nil && r.lastView.MouseMode != MouseModeNone {
			buf.WriteString(ansi.ResetModeMouseButtonEvent +
				ansi.ResetModeMouseAnyEvent +
				ansi.ResetModeMouseExtSgr)
		}
		switch view.MouseMode {
		case MouseModeCellMotion:
			buf.WriteString(ansi.SetModeMouseButtonEvent + ansi.SetModeMouseExtSgr)
		case MouseModeAllMotion:
			buf.WriteString(ansi.SetModeMouseAnyEvent + ansi.SetModeMouseExtSgr)
		}
	}

	if buf.Len() > 0 {
		_, _ = r.out.Write(buf.Bytes())
	}
}

func (r *foxfulRenderer) setCursor(cursor *Cursor) error {
	if cursor == nil {
		_, err := io.WriteString(r.out, ansi.HideCursor)
		return err
	}
	_, err := io.WriteString(r.out, ansi.ShowCursor+fmt.Sprintf("\x1b[%d;%dH", cursor.Position.Y+1, cursor.Position.X+1))
	return err
}
