package app

import (
	"math"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
)

// XScrollBar scrolls content horizontally when it is longer than the viewport.
// The offset advances with real time (1 cell per 200ms — the historical 5 FPS
// behavior of one cell per frame), so the marquee speed stays constant no
// matter what the UI frame rate is. The first/last 3 cells of the content are
// dwelled on (~600ms each), matching the original per-frame compensation.
type XScrollBar struct {
	lastContent string

	lastTime time.Time // last Tick() timestamp
	offsetF  float64   // fractional cell offset, wraps over len(content)+3

	nowFn func() time.Time // injectable clock for tests
	l     sync.Mutex
}

func NewXScrollBar() *XScrollBar {
	return &XScrollBar{
		nowFn: time.Now,
	}
}

func (b *XScrollBar) Tick(width int, content string) string {
	b.l.Lock()
	defer b.l.Unlock()

	now := b.nowFn()
	if b.lastTime.IsZero() {
		b.lastTime = now
	}
	elapsed := now.Sub(b.lastTime)
	b.lastTime = now
	if elapsed <= 0 || elapsed > time.Second {
		// First call or a long gap between renders (e.g. playback paused):
		// keep the scroll position frozen instead of jumping forward.
		elapsed = 0
	}

	if b.lastContent != content {
		b.lastContent = content
		b.offsetF = 0
		// Content reset must not inherit the inter-frame elapsed, or the new
		// line's first visible frame jumps elapsed*5 cells (cutting the head
		// after a render pause, e.g. playback resuming after <1s).
		elapsed = 0
	}

	length := runewidth.StringWidth(content)
	// Period of len(content)+3 cells: 3 cells of dwell at the end before
	// wrapping back to the start.
	period := float64(length + 3)
	if period < 1 {
		period = 1
	}
	b.offsetF = math.Mod(b.offsetF+elapsed.Seconds()*5, period)

	// a maps the old frame-based counter (starting at 1) to the same cell
	// offsets: offsetF=0 → a=-2, offsetF=3 → a=1 (scroll starts).
	a := int(b.offsetF+1) - 3

	var tmp string
	if length < width || a < 1 {
		tmp = runewidth.TruncateLeft(b.lastContent, 0, "")
	} else if a+width <= length {
		tmp = runewidth.TruncateLeft(b.lastContent, a, "")
	} else {
		tmp = runewidth.TruncateLeft(b.lastContent, length-width, "")
	}
	return runewidth.Truncate(runewidth.FillRight(tmp, width), width, "")
}
