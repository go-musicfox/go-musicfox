package ui

import (
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// LyricRenderer is a dedicated UI component for rendering lyrics.
type LyricRenderer struct {
	netease      *Netease
	lyricService *lyric.Service

	isVisible         bool
	lyricLines        int // 3 or 5
	lyricStartRow     int
	lyrics            [5]string // A fixed-size array to hold lines for rendering
	lyricNowScrollBar *app.XScrollBar
}

// NewLyricRenderer creates a new lyric renderer component.
func NewLyricRenderer(netease *Netease, lyricService *lyric.Service, initialVisibility bool) *LyricRenderer {
	return &LyricRenderer{
		netease:           netease,
		lyricService:      lyricService,
		isVisible:         initialVisibility,
		lyricNowScrollBar: app.NewXScrollBar(),
	}
}

// SetVisibility allows dynamic control over the renderer's visibility.
func (r *LyricRenderer) SetVisibility(visible bool) {
	r.isVisible = visible
}

// Update handles UI messages, primarily for resizing and configuration updates.
func (r *LyricRenderer) Update(msg tea.Msg, a *model.App) {
	main := r.netease.MustMain()
	spaceHeight := r.netease.WindowHeight() - 5 - main.MenuBottomRow()

	if !r.isVisible || spaceHeight < 3 {
		r.lyricLines = 0
		return
	}

	if spaceHeight >= 5 {
		r.lyricLines = 5
		r.lyricStartRow = (r.netease.WindowHeight()-3+main.MenuBottomRow())/2 - 3
	} else {
		r.lyricLines = 3
		r.lyricStartRow = (r.netease.WindowHeight()-3+main.MenuBottomRow())/2 - 2
	}
}

// View renders the lyric component.
func (r *LyricRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	var (
		endRow = r.netease.WindowHeight() - 4
	)

	if r.lyricLines == 0 {
		if endRow-main.MenuBottomRow() > 0 {
			return strings.Repeat("\n", endRow-main.MenuBottomRow()), endRow - main.MenuBottomRow()
		}
		return "", 0
	}

	r.prepareLyricLines()

	var lyricBuilder strings.Builder
	if r.lyricStartRow > main.MenuBottomRow() {
		lyricBuilder.WriteString(strings.Repeat("\n", r.lyricStartRow-main.MenuBottomRow()))
	}

	if main.CenterEverything() {
		r.buildLyricsCentered(main, &lyricBuilder)
	} else {
		r.buildLyricsTraditional(main, &lyricBuilder)
	}

	if endRow-r.lyricStartRow-r.lyricLines > 0 {
		lyricBuilder.WriteString(strings.Repeat("\n", endRow-r.lyricStartRow-r.lyricLines))
	}

	return lyricBuilder.String(), r.lyricStartRow - main.MenuBottomRow() + r.lyricLines
}

// prepareLyricLines fetches the latest state from the service and prepares the `r.lyrics` array for rendering.
func (r *LyricRenderer) prepareLyricLines() {
	state := r.lyricService.State()
	r.lyrics = [5]string{} // Clear previous lines

	if !state.IsRunning || state.CurrentIndex < 0 {
		return
	}

	centerIndex := (r.lyricLines - 1) / 2

	// Fill current line
	currentFrag := state.Fragments[state.CurrentIndex]
	line := currentFrag.Content
	if trans, ok := state.TranslatedFragments[currentFrag.StartTimeMs]; ok && trans != "" {
		line += " [" + trans + "]"
	}
	r.lyrics[centerIndex] = line

	// Fill previous lines
	for i := 1; i <= centerIndex; i++ {
		if state.CurrentIndex-i >= 0 {
			prevFrag := state.Fragments[state.CurrentIndex-i]
			line := prevFrag.Content
			if trans, ok := state.TranslatedFragments[prevFrag.StartTimeMs]; ok && trans != "" {
				line += " [" + trans + "]"
			}
			r.lyrics[centerIndex-i] = line
		}
	}

	// Fill next lines
	for i := 1; i < r.lyricLines-centerIndex; i++ {
		if state.CurrentIndex+i < len(state.Fragments) {
			nextFrag := state.Fragments[state.CurrentIndex+i]
			line := nextFrag.Content
			if trans, ok := state.TranslatedFragments[nextFrag.StartTimeMs]; ok && trans != "" {
				line += " [" + trans + "]"
			}
			r.lyrics[centerIndex+i] = line
		}
	}
}

// buildLyricsCentered contains the rendering logic for the centered layout.
func (r *LyricRenderer) buildLyricsCentered(_ *model.Main, lyricBuilder *strings.Builder) {
	windowWidth := r.netease.WindowWidth()
	highlightLine := 2
	startLine := highlightLine - (r.lyricLines-1)/2
	endLine := highlightLine + (r.lyricLines-1)/2
	extraPadding := 8 + max(0, (windowWidth-40)/5)
	lyricsMaxLength := windowWidth - extraPadding
	for i := startLine; i <= endLine; i++ {
		line := r.lyrics[i]
		if i == highlightLine {
			line = r.lyricNowScrollBar.Tick(lyricsMaxLength, line)
			line = strings.Trim(line, " ")
		}
		line = runewidth.Truncate(line, lyricsMaxLength, "")
		lineLength := runewidth.StringWidth(line)
		paddingLeft := (windowWidth - lineLength) / 2
		lyricBuilder.WriteString(strings.Repeat(" ", paddingLeft))
		if i == highlightLine {
			line = util.SetFgStyle(line, termenv.ANSIBrightCyan)
		} else {
			line = util.SetFgStyle(line, termenv.ANSIBrightBlack)
		}
		lyricBuilder.WriteString(line)
		lyricBuilder.WriteString(strings.Repeat(" ", windowWidth-paddingLeft-lineLength))
		lyricBuilder.WriteString("\n")
	}
}

// buildLyricsTraditional contains the rendering logic for the traditional layout.
func (r *LyricRenderer) buildLyricsTraditional(main *model.Main, lyricBuilder *strings.Builder) {
	var startCol int
	if main.IsDualColumn() {
		startCol = main.MenuStartColumn() + 3
	} else {
		startCol = main.MenuStartColumn() - 4
	}

	maxLen := r.netease.WindowWidth() - startCol - 4
	switch r.lyricLines {
	case 3:
		for i := 1; i <= 3; i++ {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := r.lyricNowScrollBar.Tick(maxLen, r.lyrics[i])
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(r.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}
			lyricBuilder.WriteString("\n")
		}
	case 5:
		for i := range 5 {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := r.lyricNowScrollBar.Tick(maxLen, r.lyrics[i])
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(r.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}
			lyricBuilder.WriteString("\n")
		}
	}
}
