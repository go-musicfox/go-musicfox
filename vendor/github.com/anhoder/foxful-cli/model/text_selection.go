package model

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// textSelection holds the mouse-driven text-selection state shared by widgets
// that support "click-drag to select, release to copy" (Popup, Notification).
// Coordinates are in full-content space: a line index into contentLines and a
// display column. Anchor is where the drag began; cursor tracks the current
// mouse position.
//
// Widgets embed textSelection anonymously so its fields and methods are
// promoted (e.g. p.selecting, p.selectionText()). Rendering that depends on a
// widget's own layout (scroll offset, per-line width) stays on the widget via
// applySelectionHighlight; everything coordinate-agnostic lives here.
type textSelection struct {
	selecting     bool
	hasSelection  bool
	selAnchorLine int
	selAnchorCol  int
	selCursorLine int
	selCursorCol  int

	// contentLines is the full, unwrapped-into-display-rows content the
	// selection indexes into. Populated by the embedding widget during render.
	contentLines []string
}

// clearSelection resets all selection state to empty.
func (s *textSelection) clearSelection() {
	s.selecting = false
	s.hasSelection = false
	s.selAnchorLine, s.selAnchorCol = 0, 0
	s.selCursorLine, s.selCursorCol = 0, 0
}

// normalizedSelection returns the selection bounds ordered so that
// (startLine, startCol) precedes (endLine, endCol) in reading order.
func (s *textSelection) normalizedSelection() (int, int, int, int) {
	if s.selAnchorLine < s.selCursorLine ||
		(s.selAnchorLine == s.selCursorLine && s.selAnchorCol <= s.selCursorCol) {
		return s.selAnchorLine, s.selAnchorCol, s.selCursorLine, s.selCursorCol
	}
	return s.selCursorLine, s.selCursorCol, s.selAnchorLine, s.selAnchorCol
}

// selectionRangeForLine returns the [left, right) display-column range selected
// on full-content line i, and whether any range is selected there.
func (s *textSelection) selectionRangeForLine(i, width int) (int, int, bool) {
	sL, sC, eL, eC := s.normalizedSelection()
	if i < sL || i > eL {
		return 0, 0, false
	}
	left, right := 0, width
	if i == sL {
		left = sC
	}
	if i == eL {
		right = eC
	}
	left = clampInt(left, 0, width)
	right = clampInt(right, 0, width)
	if right <= left {
		return 0, 0, false
	}
	return left, right, true
}

// selectionText extracts the plain-text content of the current selection,
// joining lines with newlines and trimming trailing padding on line-spanning rows.
func (s *textSelection) selectionText() string {
	if !s.hasSelection {
		return ""
	}
	sL, _, eL, _ := s.normalizedSelection()
	sL = clampInt(sL, 0, max(len(s.contentLines)-1, 0))
	eL = clampInt(eL, 0, max(len(s.contentLines)-1, 0))
	parts := make([]string, 0, eL-sL+1)
	for i := sL; i <= eL; i++ {
		line := s.contentLines[i]
		width := lipgloss.Width(line)
		left, right, ok := s.selectionRangeForLine(i, width)
		if !ok {
			parts = append(parts, "")
			continue
		}
		segment := ansi.Strip(ansi.Cut(line, left, right))
		if right >= width {
			segment = strings.TrimRight(segment, " ")
		}
		parts = append(parts, segment)
	}
	return strings.Join(parts, "\n")
}

// finalizeSelection copies the selected text to the system clipboard (OSC 52).
// A collapsed or whitespace-only selection is discarded.
func (s *textSelection) finalizeSelection() tea.Cmd {
	text := s.selectionText()
	if strings.TrimSpace(text) == "" {
		s.clearSelection()
		return nil
	}
	return tea.SetClipboard(text)
}
