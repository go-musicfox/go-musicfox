package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// CoverRenderer is a dedicated UI component for rendering album cover images
// using the Kitty graphics protocol.
type CoverRenderer struct {
	netease      *Netease
	state        playerRendererState
	imageCache   *kitty.ImageCache
	kittySupport bool

	mu            sync.Mutex
	currentSongId int64  // Track currently displayed song to avoid redundant renders
	cachedSeq     string // Cached kitty sequence
	lastStartRow  int    // Last rendered start row position
	lastStartCol  int    // Last rendered start column position
	imageRendered bool   // Whether the image has been rendered to terminal
	forceRerender bool   // Force re-render on next View call (set after resize)
	skipFrames    int    // Number of View calls to skip before rendering (for resize timing)

	// Display dimensions
	cols int
	rows int
}

// NewCoverRenderer creates a new cover image renderer component.
func NewCoverRenderer(netease *Netease, state playerRendererState) *CoverRenderer {
	kittySupport := kitty.IsSupported()

	r := &CoverRenderer{
		netease:      netease,
		state:        state,
		imageCache:   kitty.NewImageCache(10),
		kittySupport: kittySupport,
	}
	return r
}

// IsEnabled returns whether cover rendering is enabled and supported.
func (r *CoverRenderer) IsEnabled() bool {
	return r.kittySupport && configs.AppConfig.Main.Lyric.Cover.Show
}

// Update handles UI messages, primarily for resizing.
func (r *CoverRenderer) Update(msg tea.Msg, a *model.App) {
	if !r.IsEnabled() {
		return
	}

	switch msg.(type) {
	case tea.WindowSizeMsg:
		// Reset state to force re-render after resize
		// Note: Don't calculate dimensions here - netease.WindowWidth/Height
		// might not be updated yet. We'll calculate in View instead.
		r.mu.Lock()
		r.cachedSeq = ""
		r.imageRendered = false
		r.lastStartRow = 0
		r.lastStartCol = 0
		r.currentSongId = 0
		r.forceRerender = true // Force re-render on next View call
		r.cols = 0             // Reset to trigger recalculation in View
		r.rows = 0
		r.skipFrames = 2 // Skip 2 frames to let bubbletea finish redrawing
		r.mu.Unlock()
	}
}

// calculateDimensions calculates the cover image display dimensions.
func (r *CoverRenderer) calculateDimensions() {
	main := r.netease.MustMain()
	spaceHeight := r.netease.WindowHeight() - 5 - main.MenuBottomRow()

	if spaceHeight < 3 {
		r.rows = 0
		r.cols = 0
		return
	}

	// Calculate columns based on fixed width ratio (30%)
	const widthRatio = 0.3

	windowWidth := r.netease.WindowWidth()
	r.cols = int(float64(windowWidth) * widthRatio)
	if r.cols < 10 {
		r.cols = 10 // Minimum width
	}

	// Calculate rows to maintain square visual aspect ratio
	// Terminal cells are typically 2:1 (twice as tall as wide, e.g., 8x16 pixels)
	// So rows = cols / 2 makes the image appear visually square
	r.rows = r.cols / 2
	if r.rows < 3 {
		r.rows = 3
	}
	// Don't exceed available space
	if r.rows > spaceHeight {
		r.rows = spaceHeight
		// Adjust cols to maintain square aspect ratio (cols = rows * 2)
		r.cols = r.rows * 2
	}
}

// View renders the cover image component.
// This component writes directly to stdout for kitty graphics,
// bypassing bubbletea's rendering pipeline which may not handle APC sequences correctly.
func (r *CoverRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	if !r.IsEnabled() {
		return "", 0
	}

	// Only render when music is playing or paused (not at startup)
	playerState := r.state.State()
	if playerState != types.Playing && playerState != types.Paused {
		return "", 0
	}

	// Skip frames after resize to let bubbletea finish redrawing
	r.mu.Lock()
	if r.skipFrames > 0 {
		r.skipFrames--
		r.mu.Unlock()
		return "", 0
	}
	r.mu.Unlock()

	// Ensure dimensions are calculated (may not have received WindowSizeMsg yet)
	if r.cols == 0 || r.rows == 0 {
		r.calculateDimensions()
	}

	if r.rows == 0 {
		return "", 0
	}

	// Calculate the starting row for the cover
	// Align with the center of the lyrics area
	windowHeight := r.netease.WindowHeight()
	menuBottomRow := main.MenuBottomRow()
	spaceHeight := windowHeight - 5 - menuBottomRow

	var coverStartRow int
	if spaceHeight >= 5 {
		// For 5-line lyrics, center the cover with the lyrics
		// Lyrics start at (windowHeight-3+menuBottomRow)/2 - 3
		// Add offset to align cover with lyrics center
		coverStartRow = (windowHeight-3+menuBottomRow)/2 - 1
	} else {
		coverStartRow = (windowHeight-3+menuBottomRow)/2 - 1
	}

	// Ensure the start row is valid
	if coverStartRow <= menuBottomRow {
		coverStartRow = menuBottomRow + 1
	}

	// Ensure the cover fits within the visible area
	// Leave room for the progress bar at the bottom (last 3 rows)
	maxStartRow := windowHeight - 3 - r.rows
	if coverStartRow > maxStartRow {
		coverStartRow = maxStartRow
	}
	if coverStartRow < 1 {
		coverStartRow = 1
	}

	// If cover can't fit at all, skip rendering
	if r.rows > windowHeight-5 {
		return "", 0
	}

	// Calculate start column to align with menu arrow (same as song info start)
	coverStartCol := main.MenuStartColumn()
	if coverStartCol < 1 {
		coverStartCol = 1
	}

	song := r.state.CurSong()
	picUrl := getCoverUrl(song)

	if picUrl == "" {
		return "", 0
	}

	// Check if we need to re-render
	r.mu.Lock()
	forceRerender := r.forceRerender
	songChanged := song.Id != r.currentSongId
	positionChanged := r.lastStartRow != coverStartRow || r.lastStartCol != coverStartCol

	// If force rerender is set (e.g., after resize), skip all caching logic
	if !forceRerender {
		// If nothing changed and image is already rendered, skip
		if !songChanged && !positionChanged && r.imageRendered && r.cachedSeq != "" && song.Id != 0 {
			r.mu.Unlock()
			return "", 0
		}

		// If only position changed but same song, re-render at new position
		if !songChanged && r.cachedSeq != "" && song.Id != 0 {
			seq := r.cachedSeq
			r.lastStartRow = coverStartRow
			r.lastStartCol = coverStartCol
			r.mu.Unlock()
			r.writeToTerminal(seq, coverStartRow, coverStartCol, true)
			r.mu.Lock()
			r.imageRendered = true
			r.mu.Unlock()
			return "", 0
		}
	}
	r.mu.Unlock()

	// Fetch and generate kitty sequence
	kittySeq, err := r.imageCache.GetOrFetch(context.Background(), picUrl, r.cols, r.rows)
	if err != nil {
		slog.Debug("CoverRenderer: failed to fetch image", slog.Any("error", err))
		return "", 0
	}
	if kittySeq == "" {
		return "", 0
	}

	// Cache the result and render
	r.mu.Lock()
	r.currentSongId = song.Id
	r.cachedSeq = kittySeq
	r.lastStartRow = coverStartRow
	r.lastStartCol = coverStartCol
	r.mu.Unlock()

	// Write directly to stdout, delete old images when song changes
	r.writeToTerminal(kittySeq, coverStartRow, coverStartCol, true)

	r.mu.Lock()
	r.imageRendered = true
	r.forceRerender = false // Reset forceRerender after successful render
	r.mu.Unlock()

	return "", 0
}

// writeToTerminal writes the kitty graphics sequence directly to stdout,
// bypassing bubbletea's rendering pipeline.
// deleteOld controls whether to delete existing images first (only needed when changing images).
func (r *CoverRenderer) writeToTerminal(kittySeq string, startRow, startCol int, deleteOld bool) {
	// Build the output sequence
	var output string

	// Delete previous images only when requested (e.g., when song changes)
	if deleteOld {
		output += kitty.DeleteAllImages()
	}

	// Save cursor position
	output += "\x1b[s"

	// Move to the cover position (row, col)
	// Using CSI sequence for cursor positioning: ESC [ row ; col H
	output += fmt.Sprintf("\x1b[%d;%dH", startRow, startCol)

	// Output the kitty image sequence
	output += kittySeq

	// Restore cursor position
	output += "\x1b[u"

	// Write directly to stdout
	_, _ = os.Stdout.WriteString(output)

	// Sync to ensure the write is flushed immediately
	_ = os.Stdout.Sync()
}

// ClearCache clears the image cache.
func (r *CoverRenderer) ClearCache() {
	r.imageCache.Clear()
	r.mu.Lock()
	r.cachedSeq = ""
	r.currentSongId = 0
	r.imageRendered = false
	r.mu.Unlock()
}

// GetCoverWidth returns the current cover width in columns.
// Returns 0 if cover is not enabled.
func (r *CoverRenderer) GetCoverWidth() int {
	if !r.IsEnabled() {
		return 0
	}
	if r.cols == 0 {
		r.calculateDimensions()
	}
	return r.cols
}

// GetCoverEndColumn returns the column where the cover ends (start column + width).
// Returns 0 if cover is not enabled.
func (r *CoverRenderer) GetCoverEndColumn() int {
	if !r.IsEnabled() {
		return 0
	}
	main := r.netease.MustMain()
	startCol := main.MenuStartColumn()
	if startCol < 1 {
		startCol = 1
	}
	if r.cols == 0 {
		r.calculateDimensions()
	}
	return startCol + r.cols
}

// getCoverUrl extracts the cover URL from a song, with resize parameter.
func getCoverUrl(song structs.Song) string {
	picUrl := song.PicUrl
	if picUrl == "" {
		return ""
	}
	// Add resize parameter for better performance (request smaller image)
	return app.AddResizeParamForPicUrl(picUrl, 512)
}

// Close cleans up the cover renderer, clearing any displayed images.
// This should be called when the application exits.
func (r *CoverRenderer) Close() {
	if !r.kittySupport {
		return
	}

	r.mu.Lock()
	wasRendered := r.imageRendered
	r.mu.Unlock()

	// Only attempt cleanup if an image was actually rendered
	if !wasRendered {
		r.ClearCache()
		return
	}

	// Delete all Kitty graphics images
	_, _ = os.Stdout.WriteString(kitty.DeleteAllImages())

	// In non-alt-screen mode, we need to be more aggressive with cleanup.
	// Move cursor to where the image was and clear that area.
	r.mu.Lock()
	if r.lastStartRow > 0 && r.lastStartCol > 0 && r.rows > 0 {
		// Build cleanup sequence
		var cleanup strings.Builder

		// Save cursor position
		cleanup.WriteString("\x1b[s")

		// Move to where the image started
		cleanup.WriteString(fmt.Sprintf("\x1b[%d;%dH", r.lastStartRow, r.lastStartCol))

		// Clear the area where the image was (clear each line)
		for i := 0; i < r.rows; i++ {
			cleanup.WriteString("\x1b[2K") // Clear entire line
			if i < r.rows-1 {
				cleanup.WriteString("\x1b[B") // Move down one line
			}
		}

		// Restore cursor position
		cleanup.WriteString("\x1b[u")

		_, _ = os.Stdout.WriteString(cleanup.String())
	}
	r.mu.Unlock()

	// Flush to ensure everything is written before exit
	_ = os.Stdout.Sync()

	// Small delay to ensure terminal processes the commands
	// This is especially important in non-alt-screen mode
	time.Sleep(10 * time.Millisecond)

	r.ClearCache()
}
