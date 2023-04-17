package ui

import (
	"strings"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/muesli/termenv"
)

func Progress(width, fullSize int, progressRamp []string) string {
	var fullCells strings.Builder
	for i := 0; i < fullSize && i < len(progressRamp); i++ {
		if i == 0 {
			fullCells.WriteString(termenv.String(string(configs.ConfigRegistry.ProgressFirstFullChar)).Foreground(termProfile.Color(progressRamp[i])).String())
		} else if i >= width-1 {
			fullCells.WriteString(termenv.String(string(configs.ConfigRegistry.ProgressLastFullChar)).Foreground(termProfile.Color(progressRamp[i])).String())
		} else {
			fullCells.WriteString(termenv.String(string(configs.ConfigRegistry.ProgressFullChar)).Foreground(termProfile.Color(progressRamp[i])).String())
		}
	}

	var (
		emptySize  = width - fullSize
		emptyCells strings.Builder
	)
	if emptySize > 0 {
		if fullSize == 0 {
			emptyCells.WriteRune(configs.ConfigRegistry.ProgressFirstEmptyChar)
			emptySize--
		}
		emptySize--
		if emptySize > 0 {
			emptyCells.WriteString(strings.Repeat(string(configs.ConfigRegistry.ProgressEmptyChar), emptySize))
		}
		if fullSize < width {
			emptyCells.WriteRune(configs.ConfigRegistry.ProgressLastEmptyChar)
		}
	}
	return fullCells.String() + SetFgStyle(emptyCells.String(), termenv.ANSIBrightBlack)
}
