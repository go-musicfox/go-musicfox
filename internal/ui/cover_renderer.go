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

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// placeBackoffInitial is the first retry delay after a failed placement
// (tmux pane offset query).
const placeBackoffInitial = time.Second

// placeBackoffMax caps the exponential placement retry delay.
const placeBackoffMax = 30 * time.Second

const (
	coverWriteTraceMinBytes = 1024
	coverWriteSlowThreshold = 200 * time.Millisecond

	tmuxImageSingleMaxBytes = 768 * 1024
	tmuxImageRateBytes      = 1024 * 1024
	tmuxImageBurstBytes     = 1024 * 1024
	tmuxSlowCooldownInitial = 5 * time.Second
	tmuxSlowCooldownMax     = 30 * time.Second
)

// coverStdoutWrite is the actual stdout write used by writeStdout. Tests stub it.
var coverStdoutWrite = func(s string) (int, error) {
	return os.Stdout.WriteString(s)
}

var (
	tmuxCoverDisabledLogOnce sync.Once
	tmuxImageOversizeLogOnce sync.Once
	coverTmuxPaneOffset      = kitty.TmuxPaneOffset
)

type coverWriteResult struct {
	written  int
	duration time.Duration
	err      error
}

func (r coverWriteResult) complete(want int) bool {
	return r.err == nil && r.written == want
}

type tmuxImageLimitDecision struct {
	allowed    bool
	reason     string
	retryAfter time.Duration
}

type tmuxImageLimiterSnapshot struct {
	admittedBytes     int64
	tokens            int64
	cooldownRemaining time.Duration
	limitedCount      int64
}

type tmuxImageLimiter struct {
	mu sync.Mutex

	tokens        float64
	lastRefill    time.Time
	cooldownUntil time.Time
	cooldownLevel time.Duration
	admittedBytes int64
	limitedCount  int64
}

func newTmuxImageLimiter(now time.Time) *tmuxImageLimiter {
	return &tmuxImageLimiter{
		tokens:     tmuxImageBurstBytes,
		lastRefill: now,
	}
}

func (l *tmuxImageLimiter) refill(now time.Time) {
	if now.After(l.lastRefill) {
		l.tokens = min(
			float64(tmuxImageBurstBytes),
			l.tokens+now.Sub(l.lastRefill).Seconds()*float64(tmuxImageRateBytes),
		)
		l.lastRefill = now
	}
}

func (l *tmuxImageLimiter) allow(now time.Time, bytes int) tmuxImageLimitDecision {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill(now)
	if bytes > tmuxImageSingleMaxBytes {
		l.limitedCount++
		return tmuxImageLimitDecision{reason: "single_packet_limit"}
	}
	if now.Before(l.cooldownUntil) {
		l.limitedCount++
		return tmuxImageLimitDecision{
			reason:     "cooldown",
			retryAfter: l.cooldownUntil.Sub(now),
		}
	}
	if float64(bytes) > l.tokens {
		l.limitedCount++
		retryAfter := time.Duration((float64(bytes) - l.tokens) / float64(tmuxImageRateBytes) * float64(time.Second))
		return tmuxImageLimitDecision{
			reason:     "rate_limit",
			retryAfter: max(retryAfter, time.Nanosecond),
		}
	}

	l.tokens -= float64(bytes)
	l.admittedBytes += int64(bytes)
	return tmuxImageLimitDecision{allowed: true}
}

func (l *tmuxImageLimiter) report(now time.Time, result coverWriteResult, want int) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !result.complete(want) || coverWriteIsSlow(result.duration) {
		if l.cooldownLevel == 0 {
			l.cooldownLevel = tmuxSlowCooldownInitial
		} else {
			l.cooldownLevel = min(l.cooldownLevel*2, tmuxSlowCooldownMax)
		}
		l.cooldownUntil = now.Add(l.cooldownLevel)
		if !result.complete(want) {
			return "error"
		}
		return "slow"
	}

	l.cooldownLevel = 0
	l.cooldownUntil = time.Time{}
	return "normal"
}

func (l *tmuxImageLimiter) snapshot(now time.Time) tmuxImageLimiterSnapshot {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill(now)
	var remaining time.Duration
	if now.Before(l.cooldownUntil) {
		remaining = l.cooldownUntil.Sub(now)
	}
	return tmuxImageLimiterSnapshot{
		admittedBytes:     l.admittedBytes,
		tokens:            int64(l.tokens),
		cooldownRemaining: remaining,
		limitedCount:      l.limitedCount,
	}
}

// logTmuxCoverDisabledOnce explains once why covers are suppressed inside
// tmux when the experimental passthrough flag is not opted in, so a silent
// blank cover area is not mistaken for a rendering bug.
func logTmuxCoverDisabledOnce() {
	tmuxCoverDisabledLogOnce.Do(func() {
		slog.Warn("cover: kitty graphics disabled inside tmux (set main.lyric.cover.tmuxPassthrough=true to opt in; ghostty+tmux image passthrough has caused macOS watchdog reboots)")
	})
}

// CoverRenderer is a dedicated UI component for rendering album cover images
// using the Kitty graphics protocol.
type CoverRenderer struct {
	netease      *Netease
	state        playerRendererState
	imageCache   *kitty.ImageCache
	kittySupport bool

	mu sync.Mutex
	// writeMu serializes direct os.Stdout writes from the render thread, the
	// animation goroutine and Close(), so kitty sequences never interleave
	// and corrupt the escape stream.
	writeMu       sync.Mutex
	currentSongId int64  // Track currently displayed song to avoid redundant renders
	cachedSeq     string // Cached kitty sequence
	lastStartRow  int    // Last rendered start row position
	lastStartCol  int    // Last rendered start column position
	imageRendered bool   // Whether the image has been rendered to terminal
	forceRerender bool   // Force re-render on next View call (set after resize)
	skipFrames    int    // Number of View calls to skip before rendering (for resize timing)

	animImageID     uint32      // ID for animated cover
	displayImageID  uint32      // ID for tmux Unicode-placeholder virtual placement
	lastAngle       float64     // Last rendered rotation angle
	lastPlayerState types.State // Track player state to control animation

	renderingID int64              // Song ID currently being rendered in background
	renderChan  chan renderResult  // Channel for async render results
	cancelFunc  context.CancelFunc // Function to cancel the current rendering goroutine

	// Display dimensions
	cols int
	rows int

	// lastBgTransparent tracks whether the previous frame used a transparent
	// app background. When it changes, the cached kitty sequence is invalidated
	// so the cover is re-rendered with the correct z-index.
	lastBgTransparent bool

	// placeFailAt/placeBackoff implement an exponential retry backoff after
	// a failed placement (tmux pane offset query): without it, render
	// retries (song change, resize, mouse-driven rerender — even while
	// music is paused) would regenerate and rewrite megabytes of animation
	// frame data every frame while the offset query keeps failing. Guarded
	// by mu.
	placeFailAt  time.Time
	placeBackoff time.Duration

	tmuxImageLimiter *tmuxImageLimiter
}

func coverDebugEnabled() bool {
	return configs.AppConfig != nil && configs.AppConfig.Main.Debug
}

func coverWriteIsSlow(d time.Duration) bool {
	return d > coverWriteSlowThreshold
}

func coverWriteShouldTraceBegin(debug bool, n int) bool {
	return debug && n >= coverWriteTraceMinBytes
}

type renderResult struct {
	songID   int64
	sequence string
	startRow int
	startCol int
	animID   uint32
	// ok reports whether the animation placement (including the pane offset
	// query it depends on) succeeded. On failure the sequence is empty and
	// the receiver must not apply it to the terminal.
	ok bool
}

// NewCoverRenderer creates a new cover image renderer component.
func NewCoverRenderer(netease *Netease, state playerRendererState) *CoverRenderer {
	kittySupport := kitty.IsSupported()

	r := &CoverRenderer{
		netease:          netease,
		state:            state,
		imageCache:       kitty.NewImageCache(10),
		kittySupport:     kittySupport,
		animImageID:      kitty.NewImageID(),
		renderChan:       make(chan renderResult, 1),
		tmuxImageLimiter: newTmuxImageLimiter(time.Now()),
	}
	r.logCoverEnvSnapshot()
	return r
}

func (r *CoverRenderer) logCoverEnvSnapshot() {
	show, spin, tmuxPass := false, false, false
	var frameRate configs.FrameRate
	visualizerEnable := false
	playerEngine := ""
	debug := false
	if cfg := configs.AppConfig; cfg != nil {
		show = cfg.Main.Lyric.Cover.Show
		spin = cfg.Main.Lyric.Cover.Spin
		tmuxPass = cfg.Main.Lyric.Cover.TmuxPassthrough
		frameRate = cfg.Main.FrameRate
		visualizerEnable = cfg.Main.Visualizer.Enable
		playerEngine = cfg.Player.Engine
		debug = cfg.Main.Debug
	}
	// One-shot startup snapshot (not a periodic probe). Flush once so a
	// hard kill shortly after launch still leaves the env line on disk.
	slog.Info("cover: env snapshot",
		slog.Int("pid", os.Getpid()),
		slog.Uint64("rssBytes", processRSSBytes()),
		slog.String("TERM", os.Getenv("TERM")),
		slog.String("TERM_PROGRAM", os.Getenv("TERM_PROGRAM")),
		slog.Bool("tmux", os.Getenv("TMUX") != ""),
		slog.Bool("kittySupport", r.kittySupport),
		slog.Bool("kittyTmuxPassthrough", kitty.UseTmuxPassthrough()),
		slog.Bool("cover.show", show),
		slog.Bool("cover.spin", spin),
		slog.Bool("cover.tmuxPassthrough", tmuxPass),
		slog.Int("frameRate", int(frameRate)),
		slog.Bool("visualizer.enable", visualizerEnable),
		slog.String("player.engine", playerEngine),
		slog.Bool("debug", debug),
	)
	slogx.Flush()
}

// IsEnabled returns whether cover rendering is enabled and supported.
func (r *CoverRenderer) IsEnabled() bool {
	if !r.kittySupport || !configs.AppConfig.Main.Lyric.Cover.Show {
		return false
	}
	// Kitty graphics via tmux DCS passthrough can stall GPU compositors
	// (observed with Ghostty: WindowServer hang → macOS watchdog reboot).
	// Require an explicit opt-in before sending image payloads through tmux.
	if kitty.UseTmuxPassthrough() && !configs.AppConfig.Main.Lyric.Cover.TmuxPassthrough {
		logTmuxCoverDisabledOnce()
		return false
	}
	return true
}

// Update handles UI messages, primarily for resizing.
func (r *CoverRenderer) Update(msg tea.Msg, a *model.App) {
	if !r.IsEnabled() {
		return
	}

	switch msg.(type) {
	case tea.WindowSizeMsg:
		// Pane offsets change when the window is resized or panes are
		// rearranged; drop the cached geometry so the next render re-queries.
		if kitty.UseTmuxPassthrough() {
			kitty.InvalidateTmuxPaneOffset()
		}
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
	spaceHeight := r.netease.EffectiveWindowHeight() - FixedTopBottomRows - main.MenuBottomRow() - r.netease.SpectrumLines(main)

	if spaceHeight < MinSpaceHeight {
		r.rows = 0
		r.cols = 0
		return
	}

	// Get width ratio from config, default to 0.3 if not set or invalid
	widthRatio := configs.AppConfig.Main.Lyric.Cover.WidthRatio
	if widthRatio <= 0 || widthRatio > 1 {
		widthRatio = DefaultCoverWidthRatio
	}

	windowWidth := r.netease.WindowWidth()
	r.cols = max(int(float64(windowWidth)*widthRatio), MinCoverCols) // Minimum width

	// Calculate rows to maintain square visual aspect ratio
	// Terminal cells are typically 2:1 (twice as tall as wide, e.g., 8x16 pixels)
	// So rows = cols / 2 makes the image appear visually square
	r.rows = max(r.cols/TerminalCellAspectRatio, MinCoverRows)
	// Don't exceed available space
	if r.rows > spaceHeight {
		r.rows = spaceHeight
		// Adjust cols to maintain square aspect ratio (cols = rows * 2)
		r.cols = r.rows * TerminalCellAspectRatio
	}
}

// rectsOverlap returns true if two rectangles overlap.
// All coordinates are 0-indexed.
func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h2 > y2
}

// isAppBackgroundTransparent returns true when the current theme's app
// background is transparent.
func isAppBackgroundTransparent(a *model.App) bool {
	bg := a.StyleSet().AppBackground.GetBackground()
	if bg == nil {
		return true
	}
	_, isNoColor := bg.(lipgloss.NoColor)
	return isNoColor
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

	windowHeight := r.netease.EffectiveWindowHeight()

	lyricStartRow, lyricLines := r.netease.GetLyricPosition()

	// Position cover purely based on lyrics: align the cover's bottom edge to the
	// lyric block's bottom edge so they form a tight horizontal group at the same baseline.
	// The +CoverBottomAlignOffset nudges the cover down to visually match the lyric baseline,
	// compensating for the Kitty image not filling the bottom terminal cell exactly.
	coverStartRow := lyricStartRow + lyricLines - r.rows/2 - 1

	// If cover can't fit at all, skip rendering
	if r.rows > windowHeight-FixedTopBottomRows {
		return "", 0
	}

	// Calculate start column. In centerEverything mode, center the cover and the
	// lyric block together as a single group so they stay visually grouped and
	// centered in the window; otherwise align with the menu arrow.
	coverStartCol := max(main.MenuStartColumn(), 1)
	if main.CenterEverything() {
		if c, _, _ := centeredCoverLyricLayout(r.netease.WindowWidth(), r.cols); c > 0 {
			coverStartCol = c
		}
	}

	song := r.state.CurSong()
	picUrl := getCoverUrl(song)

	if picUrl == "" {
		return "", 0
	}

	// Choose the cover z-index based on whether the app background is
	// transparent. Transparent backgrounds let the cover render below the
	// text grid (z < 0), with popups naturally occluding it. Opaque
	// backgrounds require z > 0 and explicit collision detection.
	isTransparent := isAppBackgroundTransparent(a)

	r.mu.Lock()
	if isTransparent {
		kitty.CoverZIndex = -2000000000
	} else {
		kitty.CoverZIndex = 1
	}
	// Capture the z-index under the lock: the animation goroutine renders
	// asynchronously and must not read the mutable global (it races with the
	// next View() call that may flip it on a theme change).
	zIndex := kitty.CoverZIndex
	// Invalidate cached sequences when background transparency changes,
	// since the kitty sequence embeds a z-index that is now stale.
	// Always clear the image cache regardless of current render state:
	// the next View() pass must regenerate sequences with the correct z-index.
	if r.lastBgTransparent != isTransparent {
		r.imageCache.Clear()
		r.cachedSeq = ""
		r.imageRendered = false
	}
	r.lastBgTransparent = isTransparent

	// Popup collision detection: hide cover when a modal overlaps the cover area.
	// Only needed for opaque backgrounds; transparent backgrounds let the text
	// grid and popup cells naturally cover the cover via z-order.
	if !isTransparent {
		if mx, my, mw, mh, ok := a.TopModalBounds(); ok {
			if rectsOverlap(coverStartCol-1, coverStartRow-1, r.cols, r.rows, mx, my, mw, mh) {
				if r.imageRendered {
					r.deleteDisplayedImagesLocked()
					r.imageRendered = false
					r.cachedSeq = ""
				}
				r.mu.Unlock()
				return "", 0
			}
		}
	}
	r.mu.Unlock()

	// Check if we need to re-render
	r.mu.Lock()
	forceRerender := r.forceRerender
	songChanged := song.Id != r.currentSongId
	positionChanged := r.lastStartRow != coverStartRow || r.lastStartCol != coverStartCol
	// Placement backoff: while active, no new render is spawned even if the
	// conditions below hold (song change / resize / theme change included);
	// the capped backoff guarantees a retry eventually happens.
	backoffActive := r.placeBackoffActive(time.Now())

	spin := configs.AppConfig.Main.Lyric.Cover.Spin
	// Even with an explicit tmuxPassthrough opt-in, never stream rotating
	// cover frames through tmux: hundreds of PNG payloads via DCS are what
	// previously stalled Ghostty / WindowServer into a watchdog reboot.
	if spin && kitty.UseTmuxPassthrough() {
		spin = false
	}

	if spin {
		// Native Animation Mode
		// 1. Check for Async Results (Non-blocking)
		select {
		case res := <-r.renderChan:
			// Verify this result is for the song we still want to show
			if res.songID == song.Id {
				if !res.ok {
					// Placement failed (e.g. pane offset query): keep the
					// currently displayed cover untouched, clear the
					// rendering flag and arm the exponential backoff so the
					// next spawn is delayed instead of regenerating and
					// rewriting megabytes of frame data every frame.
					if r.renderingID == res.songID {
						r.renderingID = 0
					}
					r.recordPlaceFailure(time.Now())
					r.mu.Unlock()
					return "", 0
				}
				// Apply to terminal
				r.writeStdout(res.sequence)

				r.currentSongId = res.songID
				r.animImageID = res.animID
				r.lastStartRow = res.startRow
				r.lastStartCol = res.startCol
				r.imageRendered = true
				r.forceRerender = false
				r.renderingID = 0               // Clear rendering flag
				r.lastPlayerState = playerState // Initialize player state
				r.recordPlaceSuccess()          // Placement succeeded, reset backoff

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
				r.writeKitty(kitty.StopAnimation(r.animImageID))
			} else if playerState == types.Playing && r.lastPlayerState == types.Paused {
				// Resume animation
				r.writeKitty(kitty.StartAnimation(r.animImageID))
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
		// Only if we actually have something to render, and not while a
		// placement failure backoff is active: mouse-driven rerenders can
		// re-enter View every frame (even while music is paused), and without
		// the gate each retry would regenerate and rewrite megabytes of
		// frame data after the failed placement.
		if (songChanged || forceRerender || !r.imageRendered || positionChanged) && song.Id != 0 && !backoffActive {
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

			// IMPORTANT: Render static image IMMEDIATELY while animation is being calculated
			// This avoids a blank cover during the calculation time
			renderStaticForAnimation(ctx, song, picUrl, coverStartRow, coverStartCol, r.cols, r.rows, r, newAnimID, zIndex)

			// Capture variables for closure
			go func(ctx context.Context, bgSong structs.Song, bgUrl string, bgRow, bgCol int, bgCols, bgRows int, bgAnimID uint32, oldBgAnimID uint32, bgZIndex int) {
				// In tmux passthrough mode the pane offset must be known
				// before any frame data is generated or written: on failure
				// the placement cannot be built, and frame data written
				// beforehand (megabytes, wrapped in DCS passthrough) would be
				// wasted and re-written on every retry — an output flood
				// driven even by mouse motion while music is paused. Query
				// once up front; on failure report without any stdout output
				// so View clears renderingID and arms the backoff. On success
				// the captured offset is reused for the placement (a few
				// seconds of staleness while frames are generated is
				// acceptable; the next song change re-queries). The static
				// image rendered just before this goroutine spawned queried
				// the offset separately (it runs before spawn, see
				// renderStaticForAnimation); its result is shared through the
				// short kitty-level offset cache.
				var paneTop, paneLeft int
				if kitty.UseTmuxPassthrough() {
					top, left, ok := kitty.TmuxPaneOffset()
					if !ok {
						select {
						case <-ctx.Done():
							return
						case r.renderChan <- renderResult{songID: bgSong.Id, ok: false}:
						}
						return
					}
					paneTop, paneLeft = top, left
				}

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
				_spinDuration := configs.AppConfig.Main.Lyric.Cover.SpinDuration
				if _spinDuration <= 0 || _spinDuration > 30 {
					_spinDuration = 6
				}

				// Dynamic frame count calculation based on FPS and duration
				frameCount := fps * _spinDuration

				// Calculate step size (degrees per frame) to complete 360 degrees
				step := 360.0 / float64(frameCount)

				// Use ALL available CPU cores
				numWorkers := max(runtime.NumCPU(), 4) // Minimum 4 workers

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

				srcRGBA := kitty.EnsureRGBA(img)

				var wg sync.WaitGroup
				for range numWorkers {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for task := range tasks {
							if task.index%50 == 0 {
								select {
								case <-ctx.Done():
									return
								default:
								}
							}

							seq := func() string {
								rotated := kitty.RotateImagePooled(srcRGBA, task.angle)
								defer kitty.PutPooledRGBA(rotated)
								if task.index == 0 {
									return ""
								}
								s, _ := kitty.TransmitFrame(rotated, bgAnimID, frameDuration)
								return s
							}()

							resultMu.Lock()
							frameSeqs[task.index] = seq
							resultMu.Unlock()
						}
					}()
				}

				// Dispatch tasks (inline to avoid goroutine overhead)
				for i := range frameCount {
					angle := float64(i) * step
					// Workers exit early on ctx.Done (see above); without a
					// cancellation check here the buffered channel fills up and
					// this goroutine blocks forever once all workers have left.
					select {
					case tasks <- frameTask{index: i, angle: angle}:
					case <-ctx.Done():
						return
					}
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

				// Transmit frames (skipping frame 0 which is already visible),
				// chunked so a cancelled generation stops writing promptly
				// instead of flushing one multi-megabyte burst, and so the
				// terminal is not hit with a single huge write. In tmux
				// passthrough mode a batch is wrapped into a single DCS
				// packet, and tmux drops packets larger than its
				// input-buffer-size option (default 1 MiB) entirely — so each
				// batch is additionally flushed once it reaches 512 KiB,
				// keeping the wrapped packet well below the limit. Non-tmux
				// mode keeps the frame-count-only flush policy.
				const transmitBatch = 16
				const tmuxMaxBatchBytes = 512 * 1024
				tmuxMode := kitty.UseTmuxPassthrough()
				var frameData strings.Builder
				for i, seq := range frameSeqs {
					if seq == "" {
						continue
					}
					if i%transmitBatch == 0 || (tmuxMode && frameData.Len() >= tmuxMaxBatchBytes) {
						select {
						case <-ctx.Done():
							return
						default:
						}
						if frameData.Len() > 0 {
							r.writeKitty(frameData.String())
							frameData.Reset()
						}
					}
					frameData.WriteString(seq)
				}
				if frameData.Len() > 0 {
					select {
					case <-ctx.Done():
						return
					default:
					}
					r.writeKitty(frameData.String())
				}

				// Assemble sequence for animation playback; the placement is
				// built from the pane offset captured at goroutine start —
				// do not re-query here.
				sequence, seqOK := buildAnimationSequence(bgAnimID, oldBgAnimID, frameDuration, bgRow, bgCol, bgCols, bgZIndex, paneTop, paneLeft)

				// Send result (only if not cancelled); seqOK=false tells the
				// receiver the placement failed and must not be applied.
				select {
				case <-ctx.Done():
					return
				case r.renderChan <- renderResult{
					songID:   bgSong.Id,
					sequence: sequence,
					startRow: bgRow,
					startCol: bgCol,
					animID:   bgAnimID,
					ok:       seqOK,
				}:
				}
			}(ctx, song, picUrl, coverStartRow, coverStartCol, r.cols, r.rows, newAnimID, oldAnimID, zIndex)

			return "", 0
		}

		r.mu.Unlock()
		return "", 0
	}

	// Static Logic
	tmuxUnicode := kitty.UseTmuxPassthrough()
	// If force rerender is set (e.g., after resize), skip all caching logic
	if !forceRerender {
		alreadyShown := r.imageRendered && song.Id != 0 &&
			((tmuxUnicode && r.displayImageID != 0) || (!tmuxUnicode && r.cachedSeq != ""))
		// If nothing changed and image is already rendered, skip
		if !songChanged && !positionChanged && alreadyShown {
			if tmuxUnicode {
				r.applyCoverBackgroundExclusionLocked(a)
			}
			r.mu.Unlock()
			return "", 0
		}

		// If only position changed but same song, update geometry (tmux
		// Unicode placeholders follow text-grid coordinates) or re-CUP the
		// absolute overlay (non-tmux). Skipped while placement backoff is
		// active; the fall-through gate below returns instead.
		if !songChanged && song.Id != 0 && !backoffActive {
			if tmuxUnicode && r.imageRendered && r.displayImageID != 0 {
				r.lastStartRow = coverStartRow
				r.lastStartCol = coverStartCol
				r.applyCoverBackgroundExclusionLocked(a)
				r.mu.Unlock()
				return "", 0
			}
			if !tmuxUnicode && r.cachedSeq != "" {
				seq := r.cachedSeq
				r.lastStartRow = coverStartRow
				r.lastStartCol = coverStartCol
				r.mu.Unlock()
				written := r.writeToTerminal(seq, coverStartRow, coverStartCol, true)
				r.mu.Lock()
				if written {
					r.imageRendered = true
					r.recordPlaceSuccess()
				}
				r.mu.Unlock()
				return "", 0
			}
		}
	}
	r.mu.Unlock()

	// Throttle the fetch/render retry while the placement backoff is active.
	if backoffActive {
		return "", 0
	}

	if tmuxUnicode {
		return r.renderStaticTmuxUnicode(a, song, picUrl, coverStartRow, coverStartCol)
	}

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
	written := r.writeToTerminal(kittySeq, coverStartRow, coverStartCol, true)

	r.mu.Lock()
	// Only mark success when the write happened (tmux pane offset query may
	// fail); otherwise the next frame retries via the !imageRendered path
	// (throttled by the placement backoff).
	if written {
		r.imageRendered = true
		r.forceRerender = false // Reset forceRerender after successful render
		r.recordPlaceSuccess()
	}
	r.mu.Unlock()

	return "", 0
}

// writeStdout serializes direct terminal writes from the cover renderer
// (View, animation goroutines, Close) so kitty sequences never interleave.
//
// Intentionally does not Sync(): Kitty image payloads can be large, and
// Sync() blocks until the PTY consumer drains. If the terminal GPU path is
// stalled (the Ghostty + tmux failure mode), Sync turns a soft lag into a
// hard hang that can freeze the UI thread and contribute to watchdog
// timeouts. The kernel still delivers writes through the PTY buffer.
func (r *CoverRenderer) writeStdout(s string) coverWriteResult {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	n := len(s)
	debug := coverDebugEnabled()
	if coverWriteShouldTraceBegin(debug, n) {
		// Event-driven only (no periodic probe). Match upstream debug style:
		// log when something happens; do not fsync on every debug line.
		slog.Debug("cover: stdout write begin",
			slog.Int("bytes", n),
			slog.Uint64("rssBytes", processRSSBytes()),
			slog.Int("goroutines", runtime.NumGoroutine()),
		)
	}

	start := time.Now()
	written, err := coverStdoutWrite(s)
	dur := time.Since(start)

	if coverWriteShouldTraceBegin(debug, n) {
		limitState := r.imageLimiter().snapshot(time.Now())
		slog.Debug("cover: stdout write end",
			slog.Int("bytes", n),
			slog.Duration("duration", dur),
			slog.Int("n", written),
			slog.Any("err", err),
			slog.Uint64("rssBytes", processRSSBytes()),
			slog.Int("goroutines", runtime.NumGoroutine()),
			slog.Int64("tmuxAdmittedBytes", limitState.admittedBytes),
			slog.Int64("tmuxLimitedCount", limitState.limitedCount),
			slog.Duration("tmuxCooldownRemaining", limitState.cooldownRemaining),
		)
	}
	if coverWriteIsSlow(dur) {
		slog.Warn("cover: stdout write slow",
			slog.Int("bytes", n),
			slog.Duration("duration", dur),
			slog.Uint64("rssBytes", processRSSBytes()),
		)
		// Flush only on Warn: survives hard kill without turning debug into
		// a continuous fsync load.
		slogx.Flush()
	}
	return coverWriteResult{written: written, duration: dur, err: err}
}

// writeKitty writes a bare kitty APC sequence, wrapping it in tmux DCS
// passthrough when running in tmux passthrough mode (Wrap checks the mode
// itself). For positioning writes use writePositioned instead: pane-relative
// cursor sequences must not be passed through to the outer terminal.
func (r *CoverRenderer) writeKitty(s string) {
	r.writeStdout(kitty.Wrap(s))
}

// deleteDisplayedImagesLocked removes the currently displayed cover image.
// Under tmux Unicode-placeholder mode virtual placements require d=i (by ID);
// d=a does not affect them. Caller must hold r.mu.
func (r *CoverRenderer) deleteDisplayedImagesLocked() {
	if kitty.UseTmuxPassthrough() {
		id := r.displayImageID
		r.displayImageID = 0
		if id != 0 {
			r.writeKitty(kitty.DeleteImage(id))
		}
		return
	}
	r.writeKitty(kitty.DeleteAllImages())
}

// applyCoverBackgroundExclusionLocked registers the cover rect so Main.View
// leaves those cells without AppBackground fill. Absolute Kitty overlays use
// height rows-1 (image does not fill the last cell row). Unicode placeholders
// occupy every placement row, so the full height is excluded. Caller must hold r.mu.
func (r *CoverRenderer) applyCoverBackgroundExclusionLocked(a *model.App) {
	if a == nil || !kitty.UseTmuxPassthrough() || !r.imageRendered {
		return
	}
	if r.cols <= 0 || r.rows <= 0 || r.lastStartRow <= 0 || r.lastStartCol <= 0 {
		return
	}
	a.SetAppBackgroundExclusion(r.lastStartCol-1, r.lastStartRow-1, r.cols, r.rows)
}

// PlaceholderSegment returns a Unicode-placeholder string for the cover
// columns on absolute screen row absRow. absRow is 1-based, matching the
// coverStartRow / lyricStartRow conventions used by CoverRenderer.
// ok is false when not in tmux placeholder mode, no image is active, or
// absRow lies outside the cover rectangle.
func (r *CoverRenderer) PlaceholderSegment(absRow int) (startCol int, cells string, ok bool) {
	if r == nil || !kitty.UseTmuxPassthrough() {
		return 0, "", false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.imageRendered || r.displayImageID == 0 || r.cols <= 0 || r.rows <= 0 || r.lastStartRow <= 0 {
		return 0, "", false
	}
	row := absRow - r.lastStartRow
	if row < 0 || row >= r.rows {
		return 0, "", false
	}
	return r.lastStartCol, kitty.UnicodePlaceholderRow(r.displayImageID, row, r.cols), true
}

// PlaceholderCacheFields returns the tmux Unicode cover identity and geometry
// used by LyricRenderer to invalidate its output cache when the cover changes.
func (r *CoverRenderer) PlaceholderCacheFields() (imageID uint32, startRow, startCol, cols int) {
	if r == nil {
		return 0, 0, 0, 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.displayImageID, r.lastStartRow, r.lastStartCol, r.cols
}

// renderStaticTmuxUnicode transmits an image with a=t, creates a virtual
// placement (U=1), and relies on LyricRenderer PlaceholderSegment cells for
// grid-resident display — no outer-terminal absolute CUP.
func (r *CoverRenderer) renderStaticTmuxUnicode(a *model.App, song structs.Song, picUrl string, coverStartRow, coverStartCol int) (string, int) {
	img, err := r.imageCache.GetImage(context.Background(), picUrl, r.cols, r.rows)
	if err != nil || img == nil {
		slog.Debug("CoverRenderer: failed to fetch image for tmux unicode cover", slog.Any("error", err))
		return "", 0
	}

	imageID := kitty.NewImageID()
	transmit, err := kitty.TransmitImage(img, r.cols, r.rows, imageID)
	if err != nil {
		slog.Debug("CoverRenderer: TransmitImage failed", slog.Any("error", err))
		return "", 0
	}

	payload := transmit + kitty.VirtualPlaceImage(imageID, r.cols, r.rows)
	r.mu.Lock()
	oldID := r.displayImageID
	r.mu.Unlock()
	if oldID != 0 {
		payload = kitty.DeleteImage(oldID) + payload
	}

	if !r.writeTmuxLimited(kitty.Wrap(payload)) {
		return "", 0
	}

	r.mu.Lock()
	r.currentSongId = song.Id
	r.displayImageID = imageID
	r.cachedSeq = ""
	r.lastStartRow = coverStartRow
	r.lastStartCol = coverStartCol
	r.imageRendered = true
	r.forceRerender = false
	r.recordPlaceSuccess()
	r.applyCoverBackgroundExclusionLocked(a)
	r.mu.Unlock()
	return "", 0
}

// writeTmuxLimited rate-limits and writes an already-wrapped tmux DCS payload.
// On limiter rejection or incomplete write it arms placement backoff and
// returns false so callers do not mark the render successful.
func (r *CoverRenderer) writeTmuxLimited(wrapped string) bool {
	now := time.Now()
	limiter := r.imageLimiter()
	decision := limiter.allow(now, len(wrapped))
	if !decision.allowed {
		slog.Debug("cover: tmux image write limited",
			slog.Int("bytes", len(wrapped)),
			slog.String("reason", decision.reason),
			slog.Duration("retryAfter", decision.retryAfter),
		)
		if decision.reason == "single_packet_limit" {
			tmuxImageOversizeLogOnce.Do(func() {
				slog.Warn("cover: tmux image packet exceeds runtime safety limit",
					slog.Int("bytes", len(wrapped)),
					slog.Int("limitBytes", tmuxImageSingleMaxBytes),
				)
			})
		}
		r.mu.Lock()
		r.recordPlaceFailure(now)
		r.mu.Unlock()
		return false
	}

	result := r.writeStdout(wrapped)
	pressure := limiter.report(time.Now(), result, len(wrapped))
	if coverDebugEnabled() {
		var throughput int64
		if result.duration > 0 {
			throughput = int64(float64(result.written) / result.duration.Seconds())
		}
		slog.Debug("cover: tmux image write",
			slog.Int("bytes", len(wrapped)),
			slog.Int("written", result.written),
			slog.Duration("duration", result.duration),
			slog.Int64("throughputBytesPerSec", throughput),
			slog.Int("limitBytesPerSec", tmuxImageRateBytes),
			slog.Int("burstBytes", tmuxImageBurstBytes),
			slog.String("pressureProxy", pressure),
		)
	}
	if !result.complete(len(wrapped)) {
		r.mu.Lock()
		r.recordPlaceFailure(time.Now())
		r.mu.Unlock()
		return false
	}
	return true
}

func (r *CoverRenderer) imageLimiter() *tmuxImageLimiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.tmuxImageLimiter == nil {
		r.tmuxImageLimiter = newTmuxImageLimiter(time.Now())
	}
	return r.tmuxImageLimiter
}

// writePositioned writes the kitty image sequence positioned at the given
// 1-based in-pane row/column and reports whether anything was written. In
// non-tmux mode the behavior is unchanged (optional DeleteAllImages + \e[s +
// CUP + image + \e[u) and it always returns true. In tmux passthrough mode,
// positioning must target the outer terminal's absolute cursor: tmux only
// restores the real cursor to the focused pane on redraw, so pane-relative
// CUP sequences would paint the image at whatever pane currently owns the
// cursor. Instead, the pane offset is queried via `tmux display -p` and the
// whole payload (save outer cursor + absolute CUP + image + restore outer
// cursor) is wrapped into a single DCS passthrough packet to avoid races with
// tmux redrawing. If the pane offset cannot be queried, nothing is written —
// better to skip the cover than paint it into another pane — and callers must
// not mark the render as successful so the next frame retries.
func (r *CoverRenderer) writePositioned(startRow, startCol int, imageSeq string, deleteOld bool) bool {
	if kitty.UseTmuxPassthrough() {
		top, left, ok := coverTmuxPaneOffset()
		if !ok {
			slog.Debug("CoverRenderer: failed to query tmux pane offset, skipping cover render")
			return false
		}
		payload := kitty.BuildTmuxPositionedPayload(top, left, startRow, startCol, imageSeq, deleteOld)
		wrapped := kitty.Wrap(payload)
		if imageSeq == "" {
			r.writeStdout(wrapped)
			return true
		}
		return r.writeTmuxLimited(wrapped)
	}

	// Non-tmux path: unchanged behavior.
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
	output += imageSeq

	// Restore cursor position
	output += "\x1b[u"

	r.writeStdout(output)
	return true
}

// writeToTerminal writes the kitty graphics sequence directly to stdout,
// bypassing bubbletea's rendering pipeline. It reports whether anything was
// written (see writePositioned).
// deleteOld controls whether to delete existing images first (only needed when changing images).
func (r *CoverRenderer) writeToTerminal(kittySeq string, startRow, startCol int, deleteOld bool) bool {
	return r.writePositioned(startRow, startCol, kittySeq, deleteOld)
}

// recordPlaceFailure arms the exponential placement backoff: the retry delay
// starts at placeBackoffInitial and doubles on every consecutive failure,
// capped at placeBackoffMax, so a persistently failing placement (tmux pane
// offset query) is retried at most once per window instead of per frame.
// Caller must hold r.mu.
func (r *CoverRenderer) recordPlaceFailure(now time.Time) {
	if r.placeBackoff == 0 {
		r.placeBackoff = placeBackoffInitial
	} else {
		r.placeBackoff = min(r.placeBackoff*2, placeBackoffMax)
	}
	r.placeFailAt = now
}

// placeBackoffActive reports whether a failed placement still suppresses
// render retries at time now. Caller must hold r.mu.
func (r *CoverRenderer) placeBackoffActive(now time.Time) bool {
	return r.placeBackoff > 0 && now.Sub(r.placeFailAt) < r.placeBackoff
}

// recordPlaceSuccess clears the placement backoff after a successful
// placement (static image or animation). Caller must hold r.mu.
func (r *CoverRenderer) recordPlaceSuccess() {
	r.placeFailAt = time.Time{}
	r.placeBackoff = 0
}

// buildAnimationSequence assembles the final animation playback sequence
// (frame gap setup, animation start, old image delete, new image placement)
// and reports whether the placement succeeded. In non-tmux mode it keeps the
// original layout: SetFrameGap + StartAnimation first, then pane-relative
// cursor save / CUP / PlaceImage / restore, old image delete last; it always
// returns ok=true and ignores paneTop/paneLeft. In tmux passthrough mode the
// whole sequence travels as a single wrapped DCS — including the animation
// control commands: SetFrameGap/StartAnimation are bare Kitty APCs, and a
// bare APC written to the pane is consumed by tmux as a pane title (dropped
// or polluting the title) instead of reaching the outer terminal, so the
// animation would never start. The placement targets the outer terminal's
// absolute cursor (see writePositioned). paneTop/paneLeft must have been
// queried successfully before any frame data was generated (see the render
// goroutine) and are reused here — the function never re-queries. The old
// image delete is merged into the payload so that when the offset query
// fails nothing is sent at all and the old cover stays on screen — a
// standalone delete packet would drop the old cover and leave nothing in
// its place. The delete is a by-ID delete (not a=d,d=a) because the
// placement only references already-transmitted image data: deleting all
// images would also wipe the new animation's data and make a=p fail (the
// static path can afford d=a only because a full re-transmit follows it).
// The sequence must be written with writeStdout (it is already wrapped).
func buildAnimationSequence(animID, oldAnimID uint32, frameDuration, bgRow, bgCol, bgCols, bgZIndex, paneTop, paneLeft int) (string, bool) {
	placement := kitty.PlaceImage(animID, bgCols, 0, bgZIndex)

	if kitty.UseTmuxPassthrough() {
		// Single fully wrapped DCS: animation control + positioning +
		// placement. Merge the old image delete into the payload (see doc
		// comment); the direct-mode path keeps the delete as a separate
		// plain command appended after the placement.
		if oldAnimID != 0 && oldAnimID != animID {
			placement = kitty.DeleteImage(oldAnimID) + placement
		}
		payload := kitty.SetFrameGap(animID, 1, frameDuration) +
			kitty.StartAnimation(animID) +
			kitty.BuildTmuxPositionedPayload(paneTop, paneLeft, bgRow, bgCol, placement, false)
		return kitty.Wrap(payload), true
	}

	// Non-tmux path: unchanged layout.
	var sb strings.Builder
	sb.WriteString(kitty.SetFrameGap(animID, 1, frameDuration))
	sb.WriteString(kitty.StartAnimation(animID))

	// Placement
	sb.WriteString("\x1b[s")
	fmt.Fprintf(&sb, "\x1b[%d;%dH", bgRow, bgCol)
	sb.WriteString(placement)
	sb.WriteString("\x1b[u")

	// Delete OLD ID
	if oldAnimID != 0 && oldAnimID != animID {
		sb.WriteString(kitty.DeleteImage(oldAnimID))
	}

	return sb.String(), true
}

// renderStaticForAnimation renders a static (non-spinning) version of the cover image
// immediately while the animation is being calculated in the background.
// Animation frames will overwrite this static image when ready.
func renderStaticForAnimation(ctx context.Context, song structs.Song, picUrl string, startRow, startCol, cols, rows int, r *CoverRenderer, animID uint32, zIndex int) {
	// In tmux passthrough mode check the pane offset before doing any work:
	// on failure nothing may be written (zero output), not even the image
	// fetch or PNG encode. This query runs before the animation goroutine
	// spawns; the goroutine re-queries separately and shares the result via
	// the short kitty-level offset cache (success or negative cache), with
	// the renderer-level backoff as the flood guard.
	if kitty.UseTmuxPassthrough() {
		if _, _, ok := kitty.TmuxPaneOffset(); !ok {
			return
		}
	}

	img, err := r.imageCache.GetImage(ctx, picUrl, cols, rows)
	if err != nil || img == nil {
		return
	}

	// Transmit and display the static image, with the z-index captured by the
	// caller (the global CoverZIndex is mutable and read from this goroutine).
	kittySeq, err := kitty.TransmitAndDisplayWithIDZ(img, cols, rows, animID, zIndex)
	if err != nil {
		return
	}

	// On failure (tmux pane offset query) return before touching any state,
	// so the renderer keeps treating the old cover as current and retries.
	if !r.writePositioned(startRow, startCol, kittySeq, false) {
		return
	}

	r.mu.Lock()
	r.currentSongId = song.Id
	r.cachedSeq = kittySeq
	r.lastStartRow = startRow
	r.lastStartCol = startCol
	r.imageRendered = true
	r.recordPlaceSuccess()
	r.mu.Unlock()
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
	startCol := max(main.MenuStartColumn(), 1)
	if r.cols == 0 {
		r.calculateDimensions()
	}
	return startCol + r.cols
}

// centeredCoverLyricLayout computes horizontal positions for the cover +
// lyric group in centerEverything mode. It returns the 1-indexed terminal
// column where the cover starts, the number of leading spaces before the
// lyric block, and the lyric block width used to center each lyric line.
//
// The cover, a fixed gap, and the lyric block are treated as one group that is
// centered in the window. lyricWidth is derived from windowWidth (not the
// current lyric text) so the cover's column stays stable while lyrics scroll.
// When coverWidth <= 0 (cover disabled) lyrics center across the full window,
// preserving the previous no-cover behavior.
func centeredCoverLyricLayout(windowWidth, coverWidth int) (coverStartCol, lyricStartCol, lyricWidth int) {
	if coverWidth <= 0 {
		return 1, 0, windowWidth
	}
	gap := CoverRightPadding
	lyricWidth = windowWidth / CenteredLyricBlockDivisor
	if coverWidth+gap+lyricWidth > windowWidth {
		lyricWidth = windowWidth - coverWidth - gap
	}
	if lyricWidth < MinLyricWidth {
		lyricWidth = MinLyricWidth
	}
	groupWidth := coverWidth + gap + lyricWidth
	groupStart := max(0, (windowWidth-groupWidth)/2)
	coverStartCol = groupStart + 1 // 1-indexed terminal column
	lyricStartCol = coverStartCol + coverWidth + gap
	return coverStartCol, lyricStartCol, lyricWidth
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

// ClearDisplayed clears the displayed cover image when switching pages.
func (r *CoverRenderer) ClearDisplayed() {
	if !r.IsEnabled() {
		return
	}

	r.mu.Lock()
	r.deleteDisplayedImagesLocked()
	if r.netease != nil && r.netease.App != nil {
		r.netease.ClearAppBackgroundExclusion()
	}

	if r.cancelFunc != nil {
		r.cancelFunc()
		r.cancelFunc = nil
	}

	r.imageRendered = false
	r.cachedSeq = ""
	r.currentSongId = 0
	r.animImageID = 0
	r.displayImageID = 0
	r.renderingID = 0
	r.lastStartRow = 0
	r.lastStartCol = 0
	r.mu.Unlock()
}

// Close cleans up the cover renderer, clearing any displayed images.
// This should be called when the application exits.
func (r *CoverRenderer) Close() {
	if !r.kittySupport {
		return
	}

	r.mu.Lock()
	wasRendered := r.imageRendered
	tmuxUnicode := kitty.UseTmuxPassthrough()
	displayID := r.displayImageID
	startRow, startCol, rows := r.lastStartRow, r.lastStartCol, r.rows
	r.mu.Unlock()

	// Only attempt cleanup if an image was actually rendered
	if !wasRendered {
		r.ClearCache()
		return
	}

	if tmuxUnicode {
		// Virtual placements require d=i; grid-resident placeholders clear
		// with the text buffer on exit — no outer CUP clear needed.
		if displayID != 0 {
			r.writeKitty(kitty.DeleteImage(displayID))
		}
		if r.netease != nil && r.netease.App != nil {
			r.netease.ClearAppBackgroundExclusion()
		}
		r.mu.Lock()
		r.displayImageID = 0
		r.imageRendered = false
		r.mu.Unlock()
		r.ClearCache()
		return
	}

	// Delete all Kitty graphics images
	r.writeKitty(kitty.DeleteAllImages())

	// In non-alt-screen mode, we need to be more aggressive with cleanup.
	// Move cursor to where the image was and clear that area.
	if startRow > 0 && startCol > 0 && rows > 0 {
		// Build the clear-rows payload (clear each line, moving down).
		var clearLines strings.Builder
		for i := 0; i < rows; i++ {
			clearLines.WriteString("\x1b[2K") // Clear entire line
			if i < rows-1 {
				clearLines.WriteString("\x1b[B") // Move down one line
			}
		}

		// Non-tmux path: unchanged behavior.
		var cleanup strings.Builder

		// Save cursor position
		cleanup.WriteString("\x1b[s")

		// Move to where the image started
		fmt.Fprintf(&cleanup, "\x1b[%d;%dH", startRow, startCol)

		// Clear the area where the image was
		cleanup.WriteString(clearLines.String())

		// Restore cursor position
		cleanup.WriteString("\x1b[u")

		r.writeStdout(cleanup.String())
	}

	// Small delay to ensure terminal processes the commands
	// This is especially important in non-alt-screen mode
	time.Sleep(10 * time.Millisecond)

	r.ClearCache()
}
