package ui

import "testing"

// TestCenteredCoverLyricLayout verifies the cover + lyric block is centered as
// a single group in centerEverything mode, and that disabling the cover
// preserves full-window lyric centering.
func TestCenteredCoverLyricLayout(t *testing.T) {
	cases := []struct {
		windowWidth int
		coverWidth  int
	}{
		{windowWidth: 130, coverWidth: 14},
		{windowWidth: 200, coverWidth: 30},
		{windowWidth: 80, coverWidth: 10},
	}
	for _, c := range cases {
		coverStart, lyricStart, lyricWidth := centeredCoverLyricLayout(c.windowWidth, c.coverWidth)
		gap := CoverRightPadding
		groupWidth := c.coverWidth + gap + lyricWidth
		leftMargin := coverStart - 1
		rightMargin := c.windowWidth - (coverStart - 1 + groupWidth)

		// The group must fit within the window.
		if coverStart-1+groupWidth > c.windowWidth {
			t.Errorf("group overflows window: win=%d start=%d group=%d", c.windowWidth, coverStart, groupWidth)
		}
		// Left/right margins balanced within 1 col => group centered.
		if diff := leftMargin - rightMargin; diff < -1 || diff > 1 {
			t.Errorf("group not centered: win=%d leftMargin=%d rightMargin=%d", c.windowWidth, leftMargin, rightMargin)
		}
		// Lyric block starts immediately after the cover + gap.
		if lyricStart != coverStart+c.coverWidth+gap {
			t.Errorf("lyric start mismatch: got=%d want=%d", lyricStart, coverStart+c.coverWidth+gap)
		}
		// Cover column is 1-indexed and on-screen.
		if coverStart < 1 {
			t.Errorf("cover start not 1-indexed: %d", coverStart)
		}
	}

	// No cover: lyrics span the full window (previous behavior preserved).
	coverStart, lyricStart, lyricWidth := centeredCoverLyricLayout(130, 0)
	if coverStart != 1 || lyricStart != 0 || lyricWidth != 130 {
		t.Errorf("no-cover layout changed: coverStart=%d lyricStart=%d lyricWidth=%d", coverStart, lyricStart, lyricWidth)
	}
}
