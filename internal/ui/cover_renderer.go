package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
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

	animImageID     uint32      // ID for animated cover
	lastAngle       float64     // Last rendered rotation angle
	lastPlayerState types.State // Track player state to control animation

	renderingID int64              // Song ID currently being rendered in background
	renderChan  chan renderResult  // Channel for async render results
	cancelFunc  context.CancelFunc // Function to cancel the current rendering goroutine

	// Display dimensions
	cols int
	rows int
}

type renderResult struct {
	songID   int64
	sequence string
	startRow int
	startCol int
	animID   uint32
}

// NewCoverRenderer creates a new cover image renderer component.
func NewCoverRenderer(netease *Netease, state playerRendererState) *CoverRenderer {
	kittySupport := kitty.IsSupported()

	r := &CoverRenderer{
		netease:      netease,
		state:        state,
		imageCache:   kitty.NewImageCache(10),
		kittySupport: kittySupport,
		animImageID:  kitty.NewImageID(),
		renderChan:   make(chan renderResult, 1),
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

	// Get width ratio from config, default to 0.3 if not set or invalid
	widthRatio := configs.AppConfig.Main.Lyric.Cover.WidthRatio
	if widthRatio <= 0 || widthRatio > 1 {
		widthRatio = 0.3
	}

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
	// Position it at the very bottom, may overlap with song info
	windowHeight := r.netease.WindowHeight()
	menuBottomRow := main.MenuBottomRow()

	// Position cover to end at the very bottom (windowHeight - 1)
	// This puts it as low as possible, overlapping song info area if needed
	coverStartRow := windowHeight - 2 - r.rows

	// Ensure it doesn't go above the menu
	if coverStartRow <= menuBottomRow {
		coverStartRow = menuBottomRow + 1
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

	spin := configs.AppConfig.Main.Lyric.Cover.Spin

	if spin {
		// Native Animation Mode
		// 1. Check for Async Results (Non-blocking)
		select {
		case res := <-r.renderChan:
			// Verify this result is for the song we still want to show
			if res.songID == song.Id {
				// Apply to terminal
				_, _ = os.Stdout.WriteString(res.sequence)
				_ = os.Stdout.Sync()

				r.currentSongId = res.songID
				r.animImageID = res.animID
				r.lastStartRow = res.startRow
				r.lastStartCol = res.startCol
				r.imageRendered = true
				r.forceRerender = false
				r.renderingID = 0               // Clear rendering flag
				r.lastPlayerState = playerState // Initialize player state

				// Successfully updated, return immediately to avoid re-triggering logic below
				r.mu.Unlock()
				return "", 0
			} else {
				// Old result for different song, ignore but clear flag if it matched
				if r.renderingID == res.songID {
					r.renderingID = 0
				}
			}
		default:
			// No results pending
		}

		// Check if player state changed (pause/resume)
		stateChanged := playerState != r.lastPlayerState
		if stateChanged && r.imageRendered && r.animImageID != 0 {
			if playerState == types.Paused {
				// Pause animation
				_, _ = os.Stdout.WriteString(kitty.StopAnimation(r.animImageID))
				_ = os.Stdout.Sync()
			} else if playerState == types.Playing && r.lastPlayerState == types.Paused {
				// Resume animation
				_, _ = os.Stdout.WriteString(kitty.StartAnimation(r.animImageID))
				_ = os.Stdout.Sync()
			}
			r.lastPlayerState = playerState
		}

		// 2. Short-circuit if state is perfect
		if !forceRerender && !songChanged && !positionChanged && r.imageRendered && song.Id != 0 {
			r.mu.Unlock()
			return "", 0
		}

		// 3. If we are already generating this song, just wait
		if r.renderingID == song.Id {
			r.mu.Unlock()
			return "", 0
		}

		// 4. Start Async Generation
		// Only if we actually have something to render
		if (songChanged || forceRerender || !r.imageRendered || positionChanged) && song.Id != 0 {
			// Cancel previous work if any
			if r.cancelFunc != nil {
				r.cancelFunc()
			}
			// Create cancellable context
			ctx, cancel := context.WithCancel(context.Background())
			r.cancelFunc = cancel

			r.renderingID = song.Id

			// Use double-buffering: Prepare new ID, then swap and delete old
			// Capture current (old) ID to delete later
			oldAnimID := r.animImageID
			newAnimID := kitty.NewImageID()

			r.mu.Unlock() // Release lock before spawning goroutine

			// Capture variables for closure
			go func(ctx context.Context, bgSong structs.Song, bgUrl string, bgRow, bgCol int, bgCols, bgRows int, bgAnimID uint32, oldBgAnimID uint32) {
				// Fetch image with timeout (derived from cancellable context)
				fetchCtx, fetchCancel := context.WithTimeout(ctx, 15*time.Second)
				defer fetchCancel()

				img, err := r.imageCache.GetImage(fetchCtx, bgUrl, bgCols, bgRows)
				if err != nil || img == nil {
					// Log error but don't reset renderingID to avoid retry loops.
					// If we failed, we failed. Wait for song change.
					return
				}

				// Check for cancellation after network call
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Read FPS from config (default 30, max 60)
				fps := configs.AppConfig.Main.Lyric.Cover.SpinFPS
				if fps <= 0 || fps > 60 {
					fps = 30
				}
				// Calculate frame duration in milliseconds
				frameDuration := 1000 / fps

				// Read rotation duration from config (default 6, range 1-30)
				spinDuration := configs.AppConfig.Main.Lyric.Cover.SpinDuration
				if spinDuration <= 0 || spinDuration > 30 {
					spinDuration = 6
				}

				// Dynamic frame count calculation based on FPS and duration
				// frameCount = fps * rotationDuration
				frameCount := fps * spinDuration

				// Calculate step size (degrees per frame) to complete 360 degrees
				step := 360.0 / float64(frameCount)

				// Use ALL available CPU cores
				numWorkers := runtime.NumCPU()
				if numWorkers < 4 {
					numWorkers = 4 // Minimum 4 workers
				}
				// Set GOMAXPROCS to ensure all cores are used
				runtime.GOMAXPROCS(numWorkers)

				// Task and result structures
				type frameTask struct {
					index int
					angle float64
				}

				// Larger buffers to avoid blocking
				tasks := make(chan frameTask, numWorkers*2)

				// Pre-allocate result slice to avoid allocation overhead
				frameSeqs := make([]string, frameCount)
				var resultMu sync.Mutex

				// Launch worker goroutines
				var wg sync.WaitGroup
				for i := 0; i < numWorkers; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for task := range tasks {
							// Check cancellation less frequently
							if task.index%50 == 0 {
								select {
								case <-ctx.Done():
									return
								default:
								}
							}

							// Generate rotated image and encode
							rotated := kitty.RotateImage(img, task.angle)
							var seq string
							if task.index == 0 {
								// First frame is base image
								seq, _ = kitty.TransmitImage(rotated, bgCols, bgRows, bgAnimID)
							} else {
								// Subsequent frames
								seq, _ = kitty.TransmitFrame(rotated, bgAnimID, frameDuration)
							}

							// Write directly to slice with lock
							resultMu.Lock()
							frameSeqs[task.index] = seq
							resultMu.Unlock()
						}
					}()
				}

				// Dispatch tasks (inline to avoid goroutine overhead)
				for i := 0; i < frameCount; i++ {
					angle := float64(i) * step
					tasks <- frameTask{index: i, angle: angle}
				}
				close(tasks)

				// Wait for completion
				wg.Wait()

				// Check cancellation before assembly
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Assemble final sequence
				var sb strings.Builder

				// 1. DO NOT DeleteAllImages at start (Double Buffering)
				// Instead, we will overwrite/place new ID, then delete old ID

				// 2. Write all frames in order (Using bgAnimID which is the NEW ID)
				for _, seq := range frameSeqs {
					sb.WriteString(seq)
				}

				// 3. Setup Animation (NEW ID)
				sb.WriteString(kitty.SetFrameGap(bgAnimID, 1, frameDuration))
				sb.WriteString(kitty.StartAnimation(bgAnimID))

				// 4. Placement (NEW ID) - This will draw over the old one
				sb.WriteString("\x1b[s")
				sb.WriteString(fmt.Sprintf("\x1b[%d;%dH", bgRow, bgCol))
				sb.WriteString(kitty.PlaceImage(bgAnimID, bgCols, 0)) // 0 rows = auto height
				sb.WriteString("\x1b[u")

				// 5. Delete OLD ID to free resources (if different)
				if oldBgAnimID != 0 && oldBgAnimID != bgAnimID {
					sb.WriteString(kitty.DeleteImage(oldBgAnimID))
				}

				// Send result (only if not cancelled)
				select {
				case <-ctx.Done():
					return
				case r.renderChan <- renderResult{
					songID:   bgSong.Id,
					sequence: sb.String(),
					startRow: bgRow,
					startCol: bgCol,
					animID:   bgAnimID,
				}:
				}
			}(ctx, song, picUrl, coverStartRow, coverStartCol, r.cols, r.rows, newAnimID, oldAnimID)

			return "", 0
		}

		r.mu.Unlock()
		return "", 0
	}

	// Static Logic
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
