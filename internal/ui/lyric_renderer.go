package ui

import (
	"regexp"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// LyricRenderer is a dedicated UI component for rendering lyrics.
type LyricRenderer struct {
	netease      *Netease
	lyricService *lyric.Service

	isVisible         bool
	lyricLines        int // 3 or 5
	lyricStartRow     int
	lyrics            []string // Dynamic-size slice to hold lines for rendering
	lyricNowScrollBar *app.XScrollBar
	currentTimeMs     int64     // Current playback time in milliseconds for YRC rendering
	lastViewTime      time.Time // For debug logging
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

	// Fill current YRC line with the same offset used by lyric service indexing.
	currentLine := state.YRCLines[index]
	r.lyrics[centerIndex] = r.buildYRCLineString(currentLine, r.currentTimeMs+state.OffsetMs, state.ShowTranslation)

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
	renderMode := "smooth"
	if configs.AppConfig.Main.Lyric.RenderMode != "" {
		renderMode = configs.AppConfig.Main.Lyric.RenderMode
	}

	if currentTimeMs < 0 {
		var text strings.Builder
		for _, word := range line.Words {
			text.WriteString(word.Word)
		}
		if showTranslation && line.TranslatedLyric != "" {
			text.WriteString(" [")
			text.WriteString(line.TranslatedLyric)
			text.WriteString("]")
		}
		return util.SetFgStyle(text.String(), LyricInactiveColor)
	}

	adjustedTimeMs := currentTimeMs + int64(configs.AppConfig.Main.FrameRate.DurationMs()/2)
	words, currentWordIndex, playedWords := yrcWordTimings(line, adjustedTimeMs)
	progress := 0.0
	if len(words) > 0 {
		progress = float64(playedWords) / float64(len(words))
	}

	animationTime := float64(currentTimeMs) * 0.001
	var result string
	switch renderMode {
	case "smooth":
		result = renderSmooth(words, progress)
	case "wave":
		result = renderWave(words, progress, animationTime)
	case "glow":
		if currentWordIndex < 0 {
			currentWordIndex = playedWords - 1
		}
		result = renderGlow(words, currentWordIndex, animationTime)
	default:
		result = renderSmooth(words, progress)
	}

	if showTranslation && line.TranslatedLyric != "" {
		result += " " + util.SetFgStyle("["+line.TranslatedLyric+"]", LyricInactiveColor)
	}
	return result
}

func yrcWordTimings(line lyric.YRCLine, timeMs int64) ([]wordWithTiming, int, int) {
	progress := lyric.ProgressYRCLineAtTimeMs(line, timeMs)
	words := make([]wordWithTiming, len(line.Words))
	for i, word := range line.Words {
		words[i] = wordWithTiming{text: word.Word, state: wordStateNotPlayed}
		switch {
		case i < progress.CompletedWords:
			words[i].state = wordStatePlayed
			words[i].interpolation = 1
		case i == progress.CurrentWord:
			words[i].state = wordStatePlaying
			words[i].interpolation = progress.CurrentProgress
		}
	}

	playedWords := progress.CompletedWords
	if progress.CurrentWord >= 0 {
		playedWords++
	}
	return words, progress.CurrentWord, playedWords
}

// stripAnsiCodes removes ANSI escape sequences from a string to get visible content
func stripAnsiCodes(s string) string {
	return ansiEscapeRegex.ReplaceAllString(s, "")
}

// Update handles UI messages, primarily for resizing and configuration updates.
func (r *LyricRenderer) Update(msg tea.Msg, a *model.App) {
	main := r.netease.MustMain()
	specLines := r.netease.SpectrumLines(main)
	spaceHeight := r.netease.WindowHeight() - FixedTopBottomRows - main.MenuBottomRow() - specLines

	if !r.isVisible || spaceHeight < MinSpaceHeight {
		r.lyricLines = 0
		return
	}

	endRow := r.netease.WindowHeight() - EndRowMargin
	if spaceHeight >= FullLyricLines {
		r.lyricLines = FullLyricLines
	} else {
		r.lyricLines = CompactLyricLines
	}
	r.lyricStartRow = (main.MenuBottomRow() + specLines + endRow - r.lyricLines) / 2
}

// View renders the lyric component.
func (r *LyricRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	specLines := r.netease.SpectrumLines(main)
	endRow := r.netease.WindowHeight() - EndRowMargin

	if r.lyricLines == 0 {
		fillingLines := endRow - main.MenuBottomRow() - specLines
		if fillingLines > 0 {
			return strings.Repeat("\n", fillingLines), fillingLines
		}
		return "", 0
	}

	// Update YRC playback time for word-level progress
	if player := r.netease.Player(); player != nil {
		r.SetCurrentTime(player.PassedTime().Milliseconds())
	}

	r.prepareLyricLines()

	var lyricBuilder strings.Builder
	topPadding := r.lyricStartRow - main.MenuBottomRow() - specLines
	if topPadding > 0 {
		lyricBuilder.WriteString(strings.Repeat("\n", topPadding))
	}

	if main.CenterEverything() {
		r.buildLyricsCentered(main, &lyricBuilder)
	} else {
		r.buildLyricsTraditional(main, &lyricBuilder)
	}

	if endRow-r.lyricStartRow-r.lyricLines > 0 {
		lyricBuilder.WriteString(strings.Repeat("\n", endRow-r.lyricStartRow-r.lyricLines))
	}

	// Return the view with the actual number of newlines in the output.
	// Using the actual newline count ensures that Main.View()'s 'top' tracker
	// reflects exactly how many rows were consumed, preventing both underfill
	// and overfill. Previously, the returned lines value was based on
	// r.lyricStartRow - MenuBottomRow + r.lyricLines, which could differ
	// significantly from the actual output when lyrics are centered or when
	// no lyrics are displayed.
	viewResult := lyricBuilder.String()
	return viewResult, strings.Count(viewResult, "\n")
}

// prepareLyricLines fetches the latest state from the service and prepares the `r.lyrics` array for rendering.
func (r *LyricRenderer) prepareLyricLines() {
	state := r.lyricService.State()
	r.lyrics = make([]string, r.lyricLines) // Allocate based on current lyric line count

	if !state.IsRunning {
		return
	}

	centerIndex := (r.lyricLines - 1) / 2

	// If YRC is enabled, render word-by-word lyrics
	if state.YRCEnabled && len(state.YRCLines) > 0 && state.YRCLineIndex >= 0 {
		r.prepareYRCLines(state, centerIndex)
		return
	}

	// Otherwise, render traditional LRC with render mode support
	if state.CurrentIndex < 0 {
		return
	}

	r.renderLRCWithMode(state, centerIndex, r.currentTimeMs)
}

// renderLRCWithMode renders LRC lyrics with render mode support.
// LRC uses line-level color gradient based on playback progress.
func (r *LyricRenderer) renderLRCWithMode(state lyric.State, centerIndex int, currentTimeMs int64) {
	// Get render mode from config
	renderMode := "smooth"
	if configs.AppConfig.Main.Lyric.RenderMode != "" {
		renderMode = configs.AppConfig.Main.Lyric.RenderMode
	}

	// Animation time for wave/glow effects
	animationTime := float64(currentTimeMs) * 0.001 // Convert to seconds

	// Helper function to calculate line progress
	calculateLineProgress := func(startTimeMs, endTimeMs int64) float64 {
		if currentTimeMs <= startTimeMs {
			return 0.4
		}
		if currentTimeMs >= endTimeMs {
			return 1.0
		}
		lineDuration := endTimeMs - startTimeMs
		elapsedTime := currentTimeMs - startTimeMs
		return 0.4 + min(float64(elapsedTime)/float64(lineDuration), 0.6)
	}

	// Helper function to render line with mode
	renderLineWithMode := func(content string, progress float64) string {
		switch renderMode {
		case "smooth":
			return renderLRCLineSmooth(content, progress)
		case "wave":
			return renderLRCWave(content, progress, animationTime)
		case "glow":
			return renderLRCGlow(content, progress, animationTime)
		default:
			return renderLRCLineSmooth(content, progress)
		}
	}

	// Helper function to get fragment content with translation
	getFragmentContent := func(frag lyric.LRCFragment) string {
		content := frag.Content
		if state.ShowTranslation {
			if trans, ok := state.TranslatedFragments[frag.StartTimeMs]; ok && trans != "" {
				content += " [" + trans + "]"
			}
		}
		return content
	}

	// Helper function to render plain gray text (for non-current lines)
	renderPlainGray := func(content string) string {
		return util.SetFgStyle(content, LyricInactiveColor)
	}

	// Fill current line (with color gradient effect)
	if state.CurrentIndex >= 0 && state.CurrentIndex < len(state.Fragments) {
		currentFrag := state.Fragments[state.CurrentIndex]
		content := getFragmentContent(currentFrag)

		// Calculate current line progress
		var currentProgress float64
		if state.CurrentIndex+1 < len(state.Fragments) {
			nextFrag := state.Fragments[state.CurrentIndex+1]
			currentProgress = calculateLineProgress(currentFrag.StartTimeMs, nextFrag.StartTimeMs)
		} else {
			// Last line - assume 5 seconds duration
			currentProgress = calculateLineProgress(currentFrag.StartTimeMs, currentFrag.StartTimeMs+5000)
		}

		r.lyrics[centerIndex] = renderLineWithMode(content, currentProgress)
	}

	// Fill previous lines (gray - already played)
	for i := 1; i <= centerIndex; i++ {
		if state.CurrentIndex-i >= 0 {
			prevFrag := state.Fragments[state.CurrentIndex-i]
			content := getFragmentContent(prevFrag)
			r.lyrics[centerIndex-i] = renderPlainGray(content)
		}
	}

	// Fill next lines (gray - not yet played)
	for i := 1; i < r.lyricLines-centerIndex; i++ {
		if state.CurrentIndex+i < len(state.Fragments) {
			nextFrag := state.Fragments[state.CurrentIndex+i]
			content := getFragmentContent(nextFrag)
			r.lyrics[centerIndex+i] = renderPlainGray(content)
		}
	}
}

// buildLyricsCentered contains the rendering logic for the centered layout.
func (r *LyricRenderer) buildLyricsCentered(_ *model.Main, lyricBuilder *strings.Builder) {
	windowWidth := r.netease.WindowWidth()
	coverEndCol := r.netease.GetCoverEndColumn()

	// Adjust available width for lyrics when cover is displayed
	// coverEndCol is the column where cover ends (1-indexed), so we need to offset by coverEndCol
	lyricStartCol := coverEndCol
	if lyricStartCol > 0 {
		lyricStartCol += CoverRightPadding // Add some padding after cover
	}
	availableWidth := windowWidth - lyricStartCol

	// 以 lyrics 数组的中心为高亮行（3 行模式索引 1，5 行模式索引 2）。
	highlightLine := (r.lyricLines - 1) / 2
	startLine := highlightLine - (r.lyricLines-1)/2
	endLine := highlightLine + (r.lyricLines-1)/2
	extraPadding := MinLyricExtraPadding + max(0, (availableWidth-LyricBaseWidth)/LyricPaddingDivisor)
	lyricsMaxLength := availableWidth - extraPadding
	if lyricsMaxLength < MinLyricWidth {
		lyricsMaxLength = MinLyricWidth
	}

	for i := startLine; i <= endLine; i++ {
		line := r.lyrics[i]
		hasAnsi := strings.Contains(line, "\033[")

		if i == highlightLine && !hasAnsi {
			line = r.lyricNowScrollBar.Tick(lyricsMaxLength, line)
			line = strings.Trim(line, " ")
		}

		if !hasAnsi {
			line = runewidth.Truncate(line, lyricsMaxLength, "")
		}

		visibleLine := line
		if hasAnsi {
			visibleLine = stripAnsiCodes(line)
		}
		lineLength := runewidth.StringWidth(visibleLine)

		paddingLeft := max(0, lyricStartCol+(availableWidth-lineLength)/2)
		lyricBuilder.WriteString(strings.Repeat(" ", paddingLeft))

		if !hasAnsi {
			if i == highlightLine {
				line = util.SetFgStyle(line, LyricActiveColor)
			} else {
				line = util.SetFgStyle(line, LyricInactiveColor)
			}
		}
		lyricBuilder.WriteString(line)
		lyricBuilder.WriteString(strings.Repeat(" ", max(0, windowWidth-paddingLeft-lineLength)))
		lyricBuilder.WriteString("\n")
	}
}

// buildLyricsTraditional contains the rendering logic for the traditional layout.
func (r *LyricRenderer) buildLyricsTraditional(main *model.Main, lyricBuilder *strings.Builder) {
	coverEndCol := r.netease.GetCoverEndColumn()

	var startCol int
	if main.IsDualColumn() {
		startCol = main.MenuStartColumn() + DualColumnLyricPadding
	} else {
		startCol = main.MenuStartColumn() - LyricHorizontalMargin
	}

	// Add cover offset to start column if needed
	if coverEndCol > 0 && startCol < coverEndCol+CoverRightPadding {
		startCol = coverEndCol + CoverRightPadding // Add some padding after cover
	}

	maxLen := r.netease.WindowWidth() - startCol - LyricHorizontalMargin
	if maxLen < MinLyricWidth {
		maxLen = MinLyricWidth
	}

	renderLine := func(idx int, isHighlight bool) {
		if startCol > 0 {
			lyricBuilder.WriteString(strings.Repeat(" ", startCol))
		}

		line := r.lyrics[idx]
		hasAnsi := strings.Contains(line, "\033[")

		var lyricLine string
		if isHighlight && !hasAnsi {
			lyricLine = r.lyricNowScrollBar.Tick(maxLen, line)
		} else if hasAnsi {
			lyricLine = line
		} else {
			lyricLine = runewidth.Truncate(runewidth.FillRight(line, maxLen), maxLen, "")
		}

		lineHasAnsi := hasAnsi || strings.Contains(lyricLine, "\033[")
		if !lineHasAnsi {
			if isHighlight {
				lyricLine = util.SetFgStyle(lyricLine, LyricActiveColor)
			} else {
				lyricLine = util.SetFgStyle(lyricLine, LyricInactiveColor)
			}
		}

		lyricBuilder.WriteString(lyricLine)

		visibleLine := lyricLine
		if lineHasAnsi {
			visibleLine = stripAnsiCodes(lyricLine)
		}
		lineLen := runewidth.StringWidth(visibleLine)
		remainingWidth := r.netease.WindowWidth() - startCol - lineLen
		if remainingWidth > 0 {
			lyricBuilder.WriteString(strings.Repeat(" ", remainingWidth))
		}
		lyricBuilder.WriteString("\n")
	}

	switch r.lyricLines {
	case 3:
		for i := range 3 {
			renderLine(i, i == 1)
		}
	case 5:
		for i := range 5 {
			renderLine(i, i == 2)
		}
	}
}

// GetLyricPosition returns the current lyric display position.
// Returns (startRow, lineCount). If lyrics are not visible, returns (0, 0).
func (r *LyricRenderer) GetLyricPosition() (startRow int, lineCount int) {
	if !r.isVisible || r.lyricLines == 0 {
		return 0, 0
	}
	return r.lyricStartRow, r.lyricLines
}
