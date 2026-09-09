package ui

import (
	"log/slog"
	"sync"
	"time"
)

type imageLimitDecision struct {
	allowed    bool
	reason     string
	retryAfter time.Duration
}

type imageLimiterSnapshot struct {
	admittedBytes     int64
	tokens            int64
	cooldownRemaining time.Duration
	limitedCount      int64
}

type imageLimiter struct {
	mu sync.Mutex

	tokens        float64
	lastRefill    time.Time
	cooldownUntil time.Time
	cooldownLevel time.Duration
	admittedBytes int64
	limitedCount  int64
}

func newImageLimiter(now time.Time) *imageLimiter {
	return &imageLimiter{
		tokens:     imageBurstBytes,
		lastRefill: now,
	}
}

func (l *imageLimiter) refill(now time.Time) {
	if now.After(l.lastRefill) {
		l.tokens = min(
			float64(imageBurstBytes),
			l.tokens+now.Sub(l.lastRefill).Seconds()*float64(imageRateBytes),
		)
		l.lastRefill = now
	}
}

func (l *imageLimiter) allow(now time.Time, bytes int) imageLimitDecision {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill(now)
	if bytes > imageSingleMaxBytes {
		l.limitedCount++
		return imageLimitDecision{reason: "single_packet_limit"}
	}
	if now.Before(l.cooldownUntil) {
		l.limitedCount++
		return imageLimitDecision{
			reason:     "cooldown",
			retryAfter: l.cooldownUntil.Sub(now),
		}
	}
	if float64(bytes) > l.tokens {
		l.limitedCount++
		retryAfter := time.Duration((float64(bytes) - l.tokens) / float64(imageRateBytes) * float64(time.Second))
		return imageLimitDecision{
			reason:     "rate_limit",
			retryAfter: max(retryAfter, time.Nanosecond),
		}
	}

	l.tokens -= float64(bytes)
	l.admittedBytes += int64(bytes)
	return imageLimitDecision{allowed: true}
}

func (l *imageLimiter) report(now time.Time, result coverWriteResult, want int) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !result.complete(want) || coverWriteIsSlow(result.duration) {
		if l.cooldownLevel == 0 {
			l.cooldownLevel = imageSlowCooldownInitial
		} else {
			l.cooldownLevel = min(l.cooldownLevel*2, imageSlowCooldownMax)
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

func (l *imageLimiter) snapshot(now time.Time) imageLimiterSnapshot {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill(now)
	var remaining time.Duration
	if now.Before(l.cooldownUntil) {
		remaining = l.cooldownUntil.Sub(now)
	}
	return imageLimiterSnapshot{
		admittedBytes:     l.admittedBytes,
		tokens:            int64(l.tokens),
		cooldownRemaining: remaining,
		limitedCount:      l.limitedCount,
	}
}

// writeImageLimited applies the shared budget to static tmux and Herdr images.
// The caller supplies the complete payload, including any tmux wrapping.
// On limiter rejection or incomplete write it arms placement backoff and
// returns false so callers do not mark the render successful.
func (r *CoverRenderer) writeImageLimited(payload string) bool {
	now := time.Now()
	limiter := r.imageLimiter()
	decision := limiter.allow(now, len(payload))
	if !decision.allowed {
		slog.Debug("cover: image write limited",
			slog.Int("bytes", len(payload)),
			slog.String("reason", decision.reason),
			slog.Duration("retryAfter", decision.retryAfter),
		)
		if decision.reason == "single_packet_limit" {
			imageOversizeLogOnce.Do(func() {
				slog.Warn("cover: image packet exceeds runtime safety limit",
					slog.Int("bytes", len(payload)),
					slog.Int("limitBytes", imageSingleMaxBytes),
				)
			})
		}
		r.mu.Lock()
		r.recordPlaceFailure(now)
		r.mu.Unlock()
		return false
	}

	result := r.writeStdout(payload)
	pressure := limiter.report(time.Now(), result, len(payload))
	if coverDebugEnabled() {
		var throughput int64
		if result.duration > 0 {
			throughput = int64(float64(result.written) / result.duration.Seconds())
		}
		slog.Debug("cover: image write",
			slog.Int("bytes", len(payload)),
			slog.Int("written", result.written),
			slog.Duration("duration", result.duration),
			slog.Int64("throughputBytesPerSec", throughput),
			slog.Int("limitBytesPerSec", imageRateBytes),
			slog.Int("burstBytes", imageBurstBytes),
			slog.String("pressureProxy", pressure),
		)
	}
	if !result.complete(len(payload)) {
		r.mu.Lock()
		r.recordPlaceFailure(time.Now())
		r.mu.Unlock()
		return false
	}
	return true
}

func (r *CoverRenderer) imageLimiter() *imageLimiter {
	r.limiterOnce.Do(func() {
		r.writeLimiter = newImageLimiter(time.Now())
	})
	return r.writeLimiter
}
