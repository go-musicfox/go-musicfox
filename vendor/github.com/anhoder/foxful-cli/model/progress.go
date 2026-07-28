package model

import (
	"image/color"
	"strings"

	"github.com/anhoder/foxful-cli/style"
)

// ProgressOptions configures the characters used to render a progress bar.
// Distinct runes for the first, last, and edge cells allow rounded or tapered
// bar styles. Filled cells use the Full* runes; empty cells use the Empty* and
// *EmptyChar runes.
type ProgressOptions struct {
	EmptyCharWhenFirst rune // empty cell rune when the bar is completely empty
	EmptyChar          rune // interior empty cell rune
	EmptyCharWhenLast  rune // empty cell rune at the trailing edge
	FirstEmptyChar     rune // first empty cell rune following the filled region
	FullCharWhenFirst  rune // filled cell rune at the leading edge
	FullChar           rune // interior filled cell rune
	FullCharWhenLast   rune // filled cell rune at the trailing edge
	LastFullChar       rune // filled cell rune for the final filled cell
}

// Progress renders a single-line progress bar of the given width. fullSize is
// the number of filled cells (0..width); progressRamp supplies a per-cell color
// gradient for the filled region (indexed by cell position). Empty cells are
// styled with the active StyleSet's ProgressEmpty style. Returns the rendered,
// ANSI-styled string.
func Progress(options *ProgressOptions, width, fullSize int, progressRamp []color.Color) string {
	// Resolve the app background once. Both filled and empty cells are painted
	// over it (component→app→transparent chain) so progress-bar cells never stay
	// transparent and reveal content drawn beneath the TUI.
	appBg := style.CurrentStyleSet().AppBackground.GetBackground()
	fullCell := func(char string, c color.Color) string {
		if appBg != nil {
			return style.FGBG(char, c, appBg)
		}
		return style.FG(char, c)
	}

	var fullCells strings.Builder
	for i := 0; i < fullSize && i < len(progressRamp); i++ {
		if i == 0 {
			fullCells.WriteString(fullCell(string(options.FullCharWhenFirst), progressRamp[i]))
		} else if i >= width-1 {
			fullCells.WriteString(fullCell(string(options.FullCharWhenLast), progressRamp[i]))
		} else if i == fullSize-1 {
			fullCells.WriteString(fullCell(string(options.LastFullChar), progressRamp[i]))
		} else {
			fullCells.WriteString(fullCell(string(options.FullChar), progressRamp[i]))
		}
	}

	var (
		emptySize  = width - fullSize
		emptyCells strings.Builder
	)
	if emptySize > 0 {
		if fullSize == 0 {
			emptyCells.WriteRune(options.EmptyCharWhenFirst)
			emptySize--
		}
		emptySize--
		if emptySize > 0 {
			emptyCells.WriteString(string(options.FirstEmptyChar))
			if emptySize > 1 {
				emptyCells.WriteString(strings.Repeat(string(options.EmptyChar), emptySize-1))
			}
		}
		if fullSize < width {
			emptyCells.WriteRune(options.EmptyCharWhenLast)
		}
	}
	// ProgressEmpty carries the empty-cell foreground; inherit the app background
	// so the empty region is opaque when a theme sets an app background.
	emptyStyle := style.CurrentStyleSet().ProgressEmpty
	if appBg != nil {
		emptyStyle = emptyStyle.Background(appBg)
	}
	return fullCells.String() + emptyStyle.Render(emptyCells.String())
}
