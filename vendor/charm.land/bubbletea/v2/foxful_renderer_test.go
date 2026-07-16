package tea

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestFoxfulRendererRedrawsPaddedLinesWithoutErasing(t *testing.T) {
	var output bytes.Buffer
	renderer := newFoxfulRenderer(&output, 8, 2)

	renderer.render(NewView("cat"))
	got := output.String()
	if !strings.Contains(got, "cat     ") {
		t.Fatalf("foxful renderer did not pad the line width: %q", got)
	}
	if strings.Contains(got, ansi.EraseEntireLine) {
		t.Fatalf("foxful renderer erased the line before redrawing: %q", got)
	}
}

func TestFoxfulRendererOverwritesShorterContentWithPadding(t *testing.T) {
	var output bytes.Buffer
	renderer := newFoxfulRenderer(&output, 8, 1)

	renderer.render(NewView("longtext"))
	output.Reset()
	renderer.render(NewView("a"))

	if got := output.String(); !strings.Contains(got, "a       ") {
		t.Fatalf("foxful renderer left stale cells after shortening content: %q", got)
	}
}

func TestFoxfulRendererSkipsEmptyTickerFlush(t *testing.T) {
	var output bytes.Buffer
	renderer := newFoxfulRenderer(&output, 8, 1)

	renderer.render(NewView("cat"))
	output.Reset()
	if err := renderer.flush(false); err != nil {
		t.Fatalf("flushing empty frame: %v", err)
	}
	if got := output.String(); got != "" {
		t.Fatalf("empty ticker flush wrote terminal output: %q", got)
	}
}

func TestFoxfulRendererDefersClearUntilNextFrame(t *testing.T) {
	var output bytes.Buffer
	renderer := newFoxfulRenderer(&output, 8, 1)

	renderer.render(NewView("old"))
	output.Reset()
	renderer.clearScreen()
	if got := output.String(); got != "" {
		t.Fatalf("clear screen exposed a blank frame: %q", got)
	}

	renderer.render(NewView("new"))
	got := output.String()
	if !strings.Contains(got, "new     ") {
		t.Fatalf("next frame did not redraw after clear request: %q", got)
	}
	if strings.Contains(got, ansi.EraseEntireScreen) {
		t.Fatalf("next frame performed a full-screen erase: %q", got)
	}
}

func TestFoxfulRendererPadsStyledLinesByDisplayWidth(t *testing.T) {
	var output bytes.Buffer
	renderer := newFoxfulRenderer(&output, 8, 1)

	renderer.render(NewView("\x1b[31mcat\x1b[0m"))
	if got := output.String(); !strings.Contains(got, "\x1b[0m     ") {
		t.Fatalf("foxful renderer did not pad a styled line by display width: %q", got)
	}
}

func TestFoxfulRendererClearsInlineScreenOnClose(t *testing.T) {
	var output bytes.Buffer
	renderer := newFoxfulRenderer(&output, 8, 1)

	renderer.render(NewView("cat"))
	output.Reset()
	if err := renderer.close(); err != nil {
		t.Fatalf("closing renderer: %v", err)
	}

	got := output.String()
	if !strings.Contains(got, ansi.EraseEntireScreen+ansi.CursorHomePosition) {
		t.Fatalf("renderer did not clear the inline screen on close: %q", got)
	}
	if !strings.Contains(got, ansi.ShowCursor) {
		t.Fatalf("renderer did not restore the cursor on close: %q", got)
	}
}
