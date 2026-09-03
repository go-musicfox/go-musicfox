package kitty

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	xansi "github.com/charmbracelet/x/ansi"
)

const (
	// tmuxPaneOffsetTTL is how long a successful pane geometry query is
	// cached. The animation render path queries the geometry once per song
	// change, but a short TTL guards against high-frequency callers without
	// spamming `tmux`.
	tmuxPaneOffsetTTL = 2 * time.Second
	// tmuxPaneOffsetFailTTL is how long a failed pane geometry query is
	// remembered. Within this window TmuxPaneOffset returns not-ok without
	// re-executing `tmux`, so render loops skip drawing instead of forking
	// the subprocess on every frame.
	tmuxPaneOffsetFailTTL = 1 * time.Second
	// tmuxExecTimeout bounds every `tmux` subprocess invocation.
	tmuxExecTimeout = 2 * time.Second
)

var (
	tmuxPaneOffsetMu     sync.Mutex
	tmuxPaneOffsetTop    int
	tmuxPaneOffsetLeft   int
	tmuxPaneOffsetCached bool
	tmuxPaneOffsetAt     time.Time
	tmuxPaneOffsetFailAt time.Time

	// tmuxExecFn is the subprocess runner used by pane-offset and client
	// count probes. Tests may override it via SetTmuxExecForTest.
	tmuxExecFn = tmuxExec
)

// tmuxExec runs `tmux` against the socket reported by $TMUX (the first
// comma-separated field of "socket_path,pid,session_id"; the default socket
// is used when $TMUX is empty or malformed) with a tmuxExecTimeout context
// and returns its stdout.
func tmuxExec(ctx context.Context, args ...string) ([]byte, error) {
	runCtx, cancel := context.WithTimeout(ctx, tmuxExecTimeout)
	defer cancel()

	tmux := os.Getenv("TMUX")
	var cmd *exec.Cmd
	if idx := strings.Index(tmux, ","); idx > 0 {
		cmd = exec.CommandContext(runCtx, "tmux", append([]string{"-S", tmux[:idx]}, args...)...)
	} else {
		// Socket path extraction failed; fall back to the default socket.
		cmd = exec.CommandContext(runCtx, "tmux", args...)
	}
	return cmd.Output()
}

// SetTmuxExecForTest overrides the tmux subprocess runner used by pane-offset
// and client-count probes. Pass nil to restore the real runner. Test-only:
// mutates package globals; callers must restore (typically via t.Cleanup)
// and must not use t.Parallel() in tests that touch it.
func SetTmuxExecForTest(fn func(ctx context.Context, args ...string) ([]byte, error)) {
	if fn == nil {
		tmuxExecFn = tmuxExec
		return
	}
	tmuxExecFn = fn
}

// TmuxClientCount returns the number of clients attached to the current tmux
// server. Returns 0 when not inside tmux or when the query fails (fail-open
// for cover gating). Uncached: callers are expected to be rare edge-case
// checks, not a hot path that needs TTL machinery.
func TmuxClientCount() int {
	if os.Getenv("TMUX") == "" {
		return 0
	}
	out, err := tmuxExecFn(context.Background(), "list-clients", "-F", "#{client_tty}")
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		n++
	}
	return n
}

// successCacheValid reports whether a cached successful pane geometry query
// is still fresh at time now.
func successCacheValid(cached bool, at, now time.Time) bool {
	return cached && now.Sub(at) < tmuxPaneOffsetTTL
}

// failureCacheValid reports whether a recently failed pane geometry query is
// still fresh enough to skip re-executing `tmux` (negative caching).
func failureCacheValid(failAt, now time.Time) bool {
	return !failAt.IsZero() && now.Sub(failAt) < tmuxPaneOffsetFailTTL
}

// TmuxPaneOffset returns the origin of the current tmux pane relative to the
// outer terminal's top-left (row 1 = outer terminal's first row), as reported
// by `tmux display -p
// '#{pane_top},#{pane_left},#{status-position},#{status-height}'`.
// pane_top/pane_left are 0-based offsets relative to the window, so the
// returned top is adjusted by the status line height when the status line
// sits at the top of the outer terminal (status-position top); with the
// default bottom status the returned top equals pane_top. The outer-terminal
// 1-based position of the pane's first cell is (top+1, left+1). Only
// meaningful in tmux passthrough mode (caller's responsibility), and
// defensively safe: returns ok=false when not inside tmux or when the query
// fails. Successful results are cached for tmuxPaneOffsetTTL; failures —
// including missing TMUX/TMUX_PANE — are negatively cached for
// tmuxPaneOffsetFailTTL so callers get a consistent backoff. Never panics.
func TmuxPaneOffset() (top, left int, ok bool) {
	tmuxPaneOffsetMu.Lock()
	defer tmuxPaneOffsetMu.Unlock()

	now := time.Now()
	if successCacheValid(tmuxPaneOffsetCached, tmuxPaneOffsetAt, now) {
		return tmuxPaneOffsetTop, tmuxPaneOffsetLeft, true
	}
	if failureCacheValid(tmuxPaneOffsetFailAt, now) {
		return 0, 0, false
	}

	// Missing TMUX/TMUX_PANE is treated as a failed query and feeds the same
	// negative cache: callers inside tmux passthrough mode get the backoff
	// instead of re-checking the environment (and re-failing) every frame.
	tmux := os.Getenv("TMUX")
	tmuxPane := os.Getenv("TMUX_PANE")
	if tmux == "" || tmuxPane == "" {
		tmuxPaneOffsetFailAt = now
		return 0, 0, false
	}

	output, err := tmuxExecFn(context.Background(), "display", "-p", "-t", tmuxPane, "#{pane_top},#{pane_left},#{status-position},#{status-height}")
	if err != nil {
		tmuxPaneOffsetFailAt = now
		return 0, 0, false
	}
	top, left, ok = parseTmuxPaneGeometry(string(output))
	if !ok {
		tmuxPaneOffsetFailAt = now
		return 0, 0, false
	}
	tmuxPaneOffsetTop = top
	tmuxPaneOffsetLeft = left
	tmuxPaneOffsetCached = true
	tmuxPaneOffsetAt = now
	return top, left, true
}

// InvalidateTmuxPaneOffset resets the pane geometry caches (both the
// successful result and the failure backoff), forcing the next TmuxPaneOffset
// call to re-execute `tmux`. Call after terminal size changes: pane_top /
// pane_left may change when the window is resized or panes are rearranged.
func InvalidateTmuxPaneOffset() {
	tmuxPaneOffsetMu.Lock()
	defer tmuxPaneOffsetMu.Unlock()
	tmuxPaneOffsetCached = false
	tmuxPaneOffsetFailAt = time.Time{}
}

// parseTmuxPaneGeometry parses a
// "#{pane_top},#{pane_left},#{status-position},#{status-height}" expansion
// and folds the status line offset into the returned top: when the status
// line sits at the top of the outer terminal (status-position "top"), the
// window starts status_height rows below the terminal's first row, so
// effectiveTop = pane_top + status_height. The pane geometry fields (first
// two) must be non-negative integers, otherwise ok=false. The status fields
// are best-effort: unparsable status-height, a missing fourth field (e.g.
// older tmux), or a status-position other than "top" degrade to a 0
// adjustment, matching the default bottom-status behavior.
func parseTmuxPaneGeometry(output string) (top, left int, ok bool) {
	line := strings.TrimSpace(output)
	if i := strings.IndexAny(line, "\r\n"); i >= 0 {
		line = line[:i]
	}
	parts := strings.Split(line, ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	top, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || top < 0 {
		return 0, 0, false
	}
	left, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || left < 0 {
		return 0, 0, false
	}
	if len(parts) >= 4 && strings.TrimSpace(parts[2]) == "top" {
		if height, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil && height > 0 {
			top += height
		}
	}
	return top, left, true
}

// Wrap wraps a Kitty APC sequence in tmux DCS passthrough when running
// inside tmux (tmux >= 3.3 with `set -g allow-passthrough on`); otherwise it
// returns the sequence unchanged. Doubling the ESC bytes of the inner
// sequence(s) is safe even when the payload contains multiple APC chunks and
// CSI sequences. Cursor positioning sequences (save/cup/restore) that must be
// consumed by tmux itself must NOT be passed through Wrap; use
// BuildTmuxPositionedPayload to embed outer-absolute positioning inside a
// single passthrough instead.
func Wrap(seq string) string {
	if tmuxPassthrough {
		return xansi.TmuxPassthrough(seq)
	}
	return seq
}

// BuildTmuxPositionedPayload builds a bare (unwrapped) payload that displays
// imageSeq at outer-terminal absolute coordinates computed from the pane
// origin (top, left, 0-based) and the 1-based in-pane position
// (startRow, startCol): outerRow = top + startRow, outerCol = left + startCol.
// The payload saves the outer cursor (ESC 7), issues an absolute CUP, plays
// the image sequence (and optionally deletes all images first), then restores
// the outer cursor (ESC 8). Callers must pass the result through Wrap exactly
// once so it becomes a single DCS passthrough packet, avoiding races with
// tmux redrawing the focused pane. Never send pane-relative \e[s/CUP/\e[u
// through passthrough: the real outer cursor may sit in another pane.
func BuildTmuxPositionedPayload(top, left, startRow, startCol int, imageSeq string, deleteOld bool) string {
	var b strings.Builder
	if deleteOld {
		b.WriteString(fmt.Sprintf("%sa=d,d=a,q=2%s", apcStart, st))
	}
	// Save outer cursor position
	b.WriteString("\x1b7")
	// Move to the outer-terminal absolute cover position (row, col)
	fmt.Fprintf(&b, "\x1b[%d;%dH", top+startRow, left+startCol)
	// Image (or other kitty/CSI) payload
	b.WriteString(imageSeq)
	// Restore outer cursor position
	b.WriteString("\x1b8")
	return b.String()
}
