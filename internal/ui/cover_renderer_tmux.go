package ui

import (
	"context"
	"log/slog"
	"time"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
)

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
func (r *CoverRenderer) PlaceholderCacheFields() (imageID uint32, startRow, startCol, cols, rows int) {
	if r == nil {
		return 0, 0, 0, 0, 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.displayImageID, r.lastStartRow, r.lastStartCol, r.cols, r.rows
}

// renderStaticTmuxUnicode transmits an image with a=t, creates a virtual
// placement (U=1), and relies on LyricRenderer PlaceholderSegment cells for
// grid-resident display — no outer-terminal absolute CUP.
func (r *CoverRenderer) renderStaticTmuxUnicode(a *model.App, song structs.Song, picURL string, coverStartRow, coverStartCol int) (string, int) {
	if r.rows <= 0 || r.cols <= 0 {
		return "", 0
	}
	img, err := r.imageCache.GetImage(context.Background(), picURL, r.cols, r.rows)
	if err != nil || img == nil {
		slog.Debug("CoverRenderer: failed to fetch image for tmux unicode cover", slog.Any("error", err))
		return "", 0
	}

	r.mu.Lock()
	imageID := pickTmuxUnicodeImageID(r.displayImageID)
	r.mu.Unlock()

	transmit, err := kitty.TransmitImage(img, r.cols, r.rows, imageID)
	if err != nil {
		slog.Debug("CoverRenderer: TransmitImage failed", slog.Any("error", err))
		return "", 0
	}

	// Reuse the same i= via a=t replace — do not DeleteImage first.
	payload := transmit + kitty.VirtualPlaceImage(imageID, r.cols, r.rows)

	if !r.writeTmuxLimited(kitty.Wrap(payload)) {
		return "", 0
	}

	r.mu.Lock()
	r.currentSongID = song.Id
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
	r.limiterOnce.Do(func() {
		r.tmuxImageLimiter = newTmuxImageLimiter(time.Now())
	})
	return r.tmuxImageLimiter
}
