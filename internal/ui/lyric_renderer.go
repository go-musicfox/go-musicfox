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
	currentTimeMs     int64 // Current playback time in milliseconds for YRC rendering
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

// SetCurrentTime sets the current playback time for YRC word highlighting.
func (r *LyricRenderer) SetCurrentTime(timeMs int64) {
	r.currentTimeMs = timeMs
}

// prepareYRCLines builds YRC word-by-word lyric lines for rendering.
func (r *LyricRenderer) prepareYRCLines(state lyric.State, centerIndex int) {
	index := state.YRCLineIndex

	// Fill current YRC line (with word-level details in a special format)
	currentLine := state.YRCLines[index]
	r.lyrics[centerIndex] = r.buildYRCLineString(currentLine, r.currentTimeMs, state.ShowTranslation)

	// Fill previous YRC lines
	for i := 1; i <= centerIndex; i++ {
		if index-i >= 0 {
			prevLine := state.YRCLines[index-i]
			r.lyrics[centerIndex-i] = r.buildYRCLineString(prevLine, -1, state.ShowTranslation) // No highlight for previous lines
		}
	}

	// Fill next YRC lines
	for i := 1; i < r.lyricLines-centerIndex; i++ {
		if index+i < len(state.YRCLines) {
			nextLine := state.YRCLines[index+i]
			r.lyrics[centerIndex+i] = r.buildYRCLineString(nextLine, -1, state.ShowTranslation) // No highlight for next lines
		}
	}
}

// buildYRCLineString constructs a displayable string from YRC line with word progress highlighting.
// If currentTimeMs >= 0, highlights words based on their timing with ANSI colors.
func (r *LyricRenderer) buildYRCLineString(line lyric.YRCLine, currentTimeMs int64, showTranslation bool) string {
	var result strings.Builder

	for _, word := range line.Words {
		if currentTimeMs >= 0 && currentTimeMs < word.StartTime {
			// Word not yet played: gray color
			result.WriteString(util.SetFgStyle(word.Word, termenv.ANSIBrightBlack))
		} else if currentTimeMs >= 0 && currentTimeMs >= word.EndTime {
			// Word fully played: cyan color
			result.WriteString(util.SetFgStyle(word.Word, termenv.ANSIBrightCyan))
		} else if currentTimeMs >= 0 && currentTimeMs >= word.StartTime {
			// Word currently playing: bright yellow/green for emphasis
			result.WriteString(util.SetFgStyle(word.Word, termenv.ANSIBrightYellow))
		} else {
			// No time tracking (non-current line): gray color
			result.WriteString(util.SetFgStyle(word.Word, termenv.ANSIBrightBlack))
		}
	}

	// Append translation if available and enabled (gray color)
	if showTranslation && line.TranslatedLyric != "" {
		result.WriteString(" ")
		result.WriteString(util.SetFgStyle("["+line.TranslatedLyric+"]", termenv.ANSIBrightBlack))
	}

	return result.String()
}

// stripAnsiCodes removes ANSI escape sequences from a string to get visible content
func stripAnsiCodes(s string) string {
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // skip '['
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
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

	// Update YRC playback time for word-level progress
	if player := r.netease.Player(); player != nil {
		r.SetCurrentTime(player.PassedTime().Milliseconds())
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

	if !state.IsRunning {
		return
	}

	centerIndex := (r.lyricLines - 1) / 2

	// If YRC is enabled, render word-by-word lyrics
	if state.YRCEnabled && len(state.YRCLines) > 0 && state.YRCLineIndex >= 0 {
		r.prepareYRCLines(state, centerIndex)
		return
	}

	// Otherwise, render traditional LRC
	if state.CurrentIndex < 0 {
		return
	}

	// Fill current line
	currentFrag := state.Fragments[state.CurrentIndex]
	line := currentFrag.Content
	if state.ShowTranslation {
		if trans, ok := state.TranslatedFragments[currentFrag.StartTimeMs]; ok && trans != "" {
			line += " [" + trans + "]"
		}
	}
	r.lyrics[centerIndex] = line

	// Fill previous lines
	for i := 1; i <= centerIndex; i++ {
		if state.CurrentIndex-i >= 0 {
			prevFrag := state.Fragments[state.CurrentIndex-i]
			line := prevFrag.Content
			if state.ShowTranslation {
				if trans, ok := state.TranslatedFragments[prevFrag.StartTimeMs]; ok && trans != "" {
					line += " [" + trans + "]"
				}
			}
			r.lyrics[centerIndex-i] = line
		}
	}

	// Fill next lines
	for i := 1; i < r.lyricLines-centerIndex; i++ {
		if state.CurrentIndex+i < len(state.Fragments) {
			nextFrag := state.Fragments[state.CurrentIndex+i]
			line := nextFrag.Content
			if state.ShowTranslation {
				if trans, ok := state.TranslatedFragments[nextFrag.StartTimeMs]; ok && trans != "" {
					line += " [" + trans + "]"
				}
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
		// 中心行如果是普通 LRC，可以滚动；如果含 ANSI 颜色码（YRC 高亮），禁止滚动避免破坏转义序列
		if i == highlightLine && !strings.Contains(line, "\033[") {
			line = r.lyricNowScrollBar.Tick(lyricsMaxLength, line)
			line = strings.Trim(line, " ")
		}
		// 不对带 ANSI 颜色码的行做截断（YRC 模式），否则会截掉 ESC 导致 "[96m" 之类残留
		if !strings.Contains(line, "\033[") {
			line = runewidth.Truncate(line, lyricsMaxLength, "")
		}
		// 计算可见宽度（去除 ANSI 码）
		visibleLine := line
		if strings.Contains(line, "\033[") {
			visibleLine = stripAnsiCodes(line)
		}
		lineLength := runewidth.StringWidth(visibleLine)
		paddingLeft := (windowWidth - lineLength) / 2
		lyricBuilder.WriteString(strings.Repeat(" ", paddingLeft))
		// Only apply uniform color if line doesn't already have ANSI codes (YRC mode)
		if !strings.Contains(line, "\033[") {
			if i == highlightLine {
				line = util.SetFgStyle(line, termenv.ANSIBrightCyan)
			} else {
				line = util.SetFgStyle(line, termenv.ANSIBrightBlack)
			}
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
			var lyricLine string
			if i == 2 && !strings.Contains(r.lyrics[i], "\033[") {
				// 只有普通 LRC 中行允许滚动；YRC 有 ANSI 颜色码时禁止滚动
				lyricLine = r.lyricNowScrollBar.Tick(maxLen, r.lyrics[i])
			} else {
				// Don't truncate lines with ANSI codes (YRC mode) as it breaks escape sequences
				if strings.Contains(r.lyrics[i], "\033[") {
					lyricLine = r.lyrics[i]
				} else {
					lyricLine = runewidth.Truncate(runewidth.FillRight(r.lyrics[i], maxLen), maxLen, "")
				}
			}
			// Only apply uniform color if line doesn't already have ANSI codes (YRC mode)
			if !strings.Contains(lyricLine, "\033[") {
				if i == 2 {
					lyricLine = util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan)
				} else {
					lyricLine = util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack)
				}
			}
			lyricBuilder.WriteString(lyricLine)
			// 计算可见宽度并填充空格到行末，避免残留
			visibleLine := lyricLine
			if strings.Contains(lyricLine, "\033[") {
				visibleLine = stripAnsiCodes(lyricLine)
			}
			lineLen := runewidth.StringWidth(visibleLine)
			remainingWidth := r.netease.WindowWidth() - startCol - lineLen
			if remainingWidth > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", remainingWidth))
			}
			lyricBuilder.WriteString("\n")
		}
	case 5:
		for i := range 5 {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			var lyricLine string
			if i == 2 && !strings.Contains(r.lyrics[i], "\033[") {
				lyricLine = r.lyricNowScrollBar.Tick(maxLen, r.lyrics[i])
			} else {
				// Don't truncate lines with ANSI codes (YRC mode) as it breaks escape sequences
				if strings.Contains(r.lyrics[i], "\033[") {
					lyricLine = r.lyrics[i]
				} else {
					lyricLine = runewidth.Truncate(runewidth.FillRight(r.lyrics[i], maxLen), maxLen, "")
				}
			}
			// Only apply uniform color if line doesn't already have ANSI codes (YRC mode)
			if !strings.Contains(lyricLine, "\033[") {
				if i == 2 {
					lyricLine = util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan)
				} else {
					lyricLine = util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack)
				}
			}
			lyricBuilder.WriteString(lyricLine)
			// 计算可见宽度并填充空格到行末，避免残留
			visibleLine := lyricLine
			if strings.Contains(lyricLine, "\033[") {
				visibleLine = stripAnsiCodes(lyricLine)
			}
			lineLen := runewidth.StringWidth(visibleLine)
			remainingWidth := r.netease.WindowWidth() - startCol - lineLen
			if remainingWidth > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", remainingWidth))
			}
			lyricBuilder.WriteString("\n")
		}
	}
}
