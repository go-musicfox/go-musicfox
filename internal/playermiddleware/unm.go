package playermiddleware

import (
	"context"
	"log/slog"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/utils/netease"
)

// NewUNMMiddleware builds the UNM (UnblockNeteaseMusic) playback middleware.
//
// It reproduces the three interaction constraints go-musicfox currently keeps
// around the vendored UNM engine, so behavior is identical to today:
//
//  1. V1 main path SkipUNM — the fetch layer (utils/netease/songinfo.go)
//     already resolves URLs with SkipUNM: true, so the engine is not triggered
//     there; this middleware only guards the resolved URL, never re-triggers
//     the engine on the main path.
//  2. ProxyURL non-empty forces the engine off — enforced inside the vendored
//     SDK (util/config.go:40-42); the middleware does not re-implement it.
//  3. banned-path interception — when UNM.SkipInvalidTracks is enabled, URLs
//     whose suffix matches the banned-link features (netease.HasBannedPathSuffix)
//     are rejected with ErrBlockedTrack, exactly like today's player.go check.
//
// The engine itself stays inside the vendored SDK (Phase 1 keeps the glue
// only); this middleware only performs the same orchestration decisions the
// current code makes.
func NewUNMMiddleware() URLMiddleware {
	return func(ctx context.Context, urlMusic *player.URLMusic, next MiddlewareNext) error {
		// Constraint 3: banned-path interception (SkipInvalidTracks).
		if configs.AppConfig.UNM.SkipInvalidTracks {
			skip, err := netease.HasBannedPathSuffix(urlMusic.URL)
			if err != nil {
				slog.Warn("banned path check failed", "url", urlMusic.URL, "error", err)
				return next(ctx, urlMusic)
			}
			if skip {
				return ErrBlockedTrack
			}
		}

		// Constraints 1 and 2 are enforced upstream (fetch layer and the
		// vendored SDK engine switch respectively), nothing to do here.
		return next(ctx, urlMusic)
	}
}
