package model

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/util"
)

type ProgressOptions struct {
	EmptyCharWhenFirst rune
	EmptyChar          rune
	EmptyCharWhenLast  rune
	FirstEmptyChar     rune
	FullCharWhenFirst  rune
	FullChar           rune
	FullCharWhenLast   rune
	LastFullChar       rune
}

func Progress(options *ProgressOptions, width, fullSize int, progressRamp []color.Color) string {
	var fullCells strings.Builder
	for i := 0; i < fullSize && i < len(progressRamp); i++ {
		style := lipgloss.NewStyle().Foreground(progressRamp[i])
		if i == 0 {
			fullCells.WriteString(style.Render(string(options.FullCharWhenFirst)))
		} else if i >= width-1 {
			fullCells.WriteString(style.Render(string(options.FullCharWhenLast)))
		} else if i == fullSize-1 {
			fullCells.WriteString(style.Render(string(options.LastFullChar)))
		} else {
			fullCells.WriteString(style.Render(string(options.FullChar)))
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
	return fullCells.String() + util.SetFgStyle(emptyCells.String(), lipgloss.BrightBlack)
}
