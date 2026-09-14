package ui

import (
	"sync"
	"time"
)

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
