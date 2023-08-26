package model

import (
	"strings"

	"github.com/anhoder/foxful-cli/util"
	"github.com/muesli/termenv"
)

type ProgressOptions struct {
	FirstEmptyChar rune
	EmptyChar      rune
	LastEmptyChar  rune
	FirstFullChar  rune
	FullChar       rune
	LastFullChar   rune
}

func Progress(options *ProgressOptions, width, fullSize int, progressRamp []string) string {
	var fullCells strings.Builder
	for i := 0; i < fullSize && i < len(progressRamp); i++ {
		if i == 0 {
			fullCells.WriteString(termenv.String(string(options.FirstFullChar)).Foreground(util.TermProfile.Color(progressRamp[i])).String())
		} else if i >= width-1 {
			fullCells.WriteString(termenv.String(string(options.LastFullChar)).Foreground(util.TermProfile.Color(progressRamp[i])).String())
		} else {
			fullCells.WriteString(termenv.String(string(options.FullChar)).Foreground(util.TermProfile.Color(progressRamp[i])).String())
		}
	}

	var (
		emptySize  = width - fullSize
		emptyCells strings.Builder
	)
	if emptySize > 0 {
		if fullSize == 0 {
			emptyCells.WriteRune(options.FirstEmptyChar)
			emptySize--
		}
		emptySize--
		if emptySize > 0 {
			emptyCells.WriteString(strings.Repeat(string(options.EmptyChar), emptySize))
		}
		if fullSize < width {
			emptyCells.WriteRune(options.LastEmptyChar)
		}
	}
	return fullCells.String() + util.SetFgStyle(emptyCells.String(), termenv.ANSIBrightBlack)
}
