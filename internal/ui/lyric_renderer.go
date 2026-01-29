package ui

import (
	"regexp"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

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
	lyrics            [5]string // A fixed-size array to hold lines for rendering
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
	// Get render mode from config (default: "smooth")
	renderMode := "smooth"
	if configs.AppConfig.Main.Lyric.RenderMode != "" {
		renderMode = configs.AppConfig.Main.Lyric.RenderMode
	}

	// For non-current lines (no time tracking), use simple gray rendering
	if currentTimeMs < 0 {
		var result strings.Builder
		for _, word := range line.Words {
			result.WriteString(util.SetFgStyle(word.Word, termenv.RGBColor(string(LyricInactiveColor))))
		}
		if showTranslation && line.TranslatedLyric != "" {
			result.WriteString(" ")
			result.WriteString(util.SetFgStyle("["+line.TranslatedLyric+"]", termenv.RGBColor(string(LyricInactiveColor))))
		}
		return result.String()
	}

	// Prepare word timing data for rendering
	var words []wordWithTiming
	var currentWordIndex int = -1
	var playedWords, totalWords int
	totalWords = len(line.Words)

	// Apply frame rate compensation for smoother animation
	frameCompensation := int64(configs.AppConfig.Main.FrameRate.DurationMs() / 2)
	adjustedTimeMs := currentTimeMs + frameCompensation

	for i, word := range line.Words {
		var state wordState
		var interpolation float64
		if adjustedTimeMs < word.StartTime {
			state = wordStateNotPlayed
			interpolation = 0.0
		} else if adjustedTimeMs >= word.EndTime {
			state = wordStatePlayed
			playedWords++
			interpolation = 1.0
		} else {
			state = wordStatePlaying
			currentWordIndex = i
			playedWords++
			// Calculate interpolation progress within the word (0.0 - 1.0)
			wordDuration := word.EndTime - word.StartTime
			if wordDuration > 0 {
				interpolation = float64(adjustedTimeMs-word.StartTime) / float64(wordDuration)
			}
		}
		words = append(words, wordWithTiming{text: word.Word, state: state, interpolation: interpolation})
	}

	// Calculate progress for smooth/wave modes
	progress := 0.0
	if totalWords > 0 {
		progress = float64(playedWords) / float64(totalWords)
	}

	// Render based on mode
	var result string
	// Use currentTimeMs as animation time for smoother effects (in milliseconds)
	animationTime := float64(currentTimeMs) * 0.001 // Convert to seconds
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
		// Default to smooth mode
		result = renderSmooth(words, progress)
	}

	// Append translation if available and enabled (gray color)
	if showTranslation && line.TranslatedLyric != "" {
		result += " " + util.SetFgStyle("["+line.TranslatedLyric+"]", termenv.RGBColor(string(LyricInactiveColor)))
	}

	return result
}

// stripAnsiCodes removes ANSI escape sequences from a string to get visible content
func stripAnsiCodes(s string) string {
	return ansiEscapeRegex.ReplaceAllString(s, "")
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
	endRow := r.netease.WindowHeight() - 4

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
		return util.SetFgStyle(content, termenv.RGBColor(string(LyricInactiveColor)))
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
		lyricStartCol += 2 // Add some padding after cover
	}
	availableWidth := windowWidth - lyricStartCol

	highlightLine := 2
	startLine := highlightLine - (r.lyricLines-1)/2
	endLine := highlightLine + (r.lyricLines-1)/2
	extraPadding := 8 + max(0, (availableWidth-40)/5)
	lyricsMaxLength := availableWidth - extraPadding
	if lyricsMaxLength < 20 {
		lyricsMaxLength = 20
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

		paddingLeft := lyricStartCol + (availableWidth-lineLength)/2
		lyricBuilder.WriteString(strings.Repeat(" ", paddingLeft))

		if !hasAnsi {
			if i == highlightLine {
				line = util.SetFgStyle(line, termenv.RGBColor(string(LyricActiveColor)))
			} else {
				line = util.SetFgStyle(line, termenv.RGBColor(string(LyricInactiveColor)))
			}
		}
		lyricBuilder.WriteString(line)
		lyricBuilder.WriteString(strings.Repeat(" ", windowWidth-paddingLeft-lineLength))
		lyricBuilder.WriteString("\n")
	}
}

// buildLyricsTraditional contains the rendering logic for the traditional layout.
func (r *LyricRenderer) buildLyricsTraditional(main *model.Main, lyricBuilder *strings.Builder) {
	coverEndCol := r.netease.GetCoverEndColumn()

	var startCol int
	if main.IsDualColumn() {
		startCol = main.MenuStartColumn() + 3
	} else {
		startCol = main.MenuStartColumn() - 4
	}

	// Add cover offset to start column if needed
	if coverEndCol > 0 && startCol < coverEndCol+2 {
		startCol = coverEndCol + 2 // Add some padding after cover
	}

	maxLen := r.netease.WindowWidth() - startCol - 4
	if maxLen < 20 {
		maxLen = 20
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
				lyricLine = util.SetFgStyle(lyricLine, termenv.RGBColor(string(LyricActiveColor)))
			} else {
				lyricLine = util.SetFgStyle(lyricLine, termenv.RGBColor(string(LyricInactiveColor)))
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
		for i := 1; i <= 3; i++ {
			renderLine(i, i == 2)
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
