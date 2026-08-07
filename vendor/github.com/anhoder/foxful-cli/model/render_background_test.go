package model

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

// TestRenderAppBackgroundMatchesLipgloss verifies the manual background
// filler produces the same visible output as the original full-frame lipgloss
// implementation for representative frame rows.
func TestRenderAppBackgroundMatchesLipgloss(t *testing.T) {
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	width := 40

	// Boundary cases around the old `len(line) < width*2` byte heuristic:
	// rows whose byte length overflows width*2 while the visible width does
	// not, and rows wider than the frame. Both previously slipped through
	// unpainted or untruncated.
	longAnsiShortVisible := strings.Repeat("\x1b[38;2;255;255;255mX\x1b[m", 15)  // ~360 bytes, 15 visible cols
	exactBoundary := "\x1b[38;2;1;2;3m" + strings.Repeat("x", 62) + "\x1b[m"    // exactly width*2 bytes, over-wide
	ansiWideShort := "\x1b[38;2;1;2;3m" + strings.Repeat("x", 50) + "\x1b[m"    // under width*2 bytes, over-wide
	ansiCjkShortVisible := "\x1b[38;2;1;2;3m" + strings.Repeat("中", 26) + "\x1b[m" // 96 bytes, 26 visible cols

	cases := []string{
		"",
		"short text",
		"styled \x1b[38;2;255;0;0mred\x1b[m text",
		"wide CJK 中文内容填满一行",
		"line1\nline2",
		"multi\nstyled \x1b[1mbold\x1b[m\nrows",
		"\x1b[38;2;1;2;3m\x1b[48;2;4;5;6mcell\x1b[m rest",
		"tail \x1b[m reset only",
		longAnsiShortVisible,
		exactBoundary,
		ansiWideShort,
		ansiCjkShortVisible,
		strings.Repeat("y", 60), // over-wide plain text
	}
	for _, c := range cases {
		got := renderAppBackground(c, width, bgStyle, backgroundRect{})
		want := renderAppBackgroundLipgloss(c, width, bgStyle, backgroundRect{})
		if stripAnsiForTest(got) != stripAnsiForTest(want) {
			t.Errorf("visible content mismatch for %q:\n  got:  %q\n  want: %q", c, stripAnsiForTest(got), stripAnsiForTest(want))
		}
		// Every visible line must reach the frame width after filling.
		for i, line := range strings.Split(stripAnsiForTest(got), "\n") {
			if runewidth.StringWidth(line) != width {
				t.Errorf("line %d of %q has width %d, want %d", i, c, runewidth.StringWidth(line), width)
			}
		}
	}
}

func stripAnsiForTest(s string) string {
	var b strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// TestRenderAppBackgroundExclusion verifies the cover-image exclusion span is
// left unpainted while surrounding segments keep the background.
func TestRenderAppBackgroundExclusion(t *testing.T) {
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	width := 40
	exclusion := backgroundRect{x: 10, y: 0, w: 5, h: 2}
	content := "line one\nline two\nline three"

	got := renderAppBackground(content, width, bgStyle, exclusion)
	want := renderAppBackgroundLipgloss(content, width, bgStyle, exclusion)

	// Visible content must match exactly.
	if stripAnsiForTest(got) != stripAnsiForTest(want) {
		t.Fatalf("visible mismatch:\n  got:  %q\n  want: %q", stripAnsiForTest(got), stripAnsiForTest(want))
	}
	// The exclusion span (visible cols 10-14) must show the same visible
	// content as the lipgloss version (raw SGR wrapping may differ).
	gotRows := strings.Split(got, "\n")
	wantRows := strings.Split(want, "\n")
	for y := 0; y < 2; y++ {
		gotMid := stripAnsiForTest(ansi.Cut(gotRows[y], 10, 15))
		wantMid := stripAnsiForTest(ansi.Cut(wantRows[y], 10, 15))
		if gotMid != wantMid {
			t.Errorf("row %d exclusion span differs: got %q, want %q", y, gotMid, wantMid)
		}
	}
}

// TestRenderAppBackgroundBlackUsesManualPath ensures the most common theme
// background (black) is painted by the manual SGR path, not the lipgloss
// fallback (whose full-frame processing was the per-frame CPU hotspot).
func TestRenderAppBackgroundBlackUsesManualPath(t *testing.T) {
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	out := renderAppBackground("x", 10, bgStyle, backgroundRect{})
	if !strings.Contains(out, "\x1b[48;2;0;0;0m") {
		t.Fatalf("black background not painted by manual SGR path: %q", out)
	}
}

// TestRenderAppBackgroundTransparentSkipsProcessing ensures an unconfigured
// (empty hex) theme background returns the content untouched instead of
// running the full-frame lipgloss pipeline.
func TestRenderAppBackgroundTransparentSkipsProcessing(t *testing.T) {
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color(""))
	content := "line one\nline two"
	if out := renderAppBackground(content, 40, bgStyle, backgroundRect{}); out != content {
		t.Fatalf("transparent background should return content untouched, got %q", out)
	}
}
