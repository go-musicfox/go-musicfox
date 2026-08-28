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

// tmuxPaneOffsetTTL is how long a successful pane geometry query is cached.
// The animation render path queries the geometry once per song change, but a
// short TTL guards against high-frequency callers without spamming `tmux`.
const tmuxPaneOffsetTTL = 2 * time.Second

var (
	tmuxPaneOffsetMu     sync.Mutex
	tmuxPaneOffsetTop    int
	tmuxPaneOffsetLeft   int
	tmuxPaneOffsetCached bool
	tmuxPaneOffsetAt     time.Time
)

// TmuxPaneOffset returns the origin of the current tmux pane relative to the
// window top-left (window top = outer terminal row 1), as reported by
// `tmux display -p '#{pane_top},#{pane_left}'`. pane_top/pane_left are 0-based
// offsets, so the outer-terminal 1-based position of the pane's first cell is
// (top+1, left+1). Only meaningful in tmux passthrough mode (caller's
// responsibility), and defensively safe: returns ok=false when not inside
// tmux or when the query fails. Successful results are cached for
// tmuxPaneOffsetTTL. Never panics.
func TmuxPaneOffset() (top, left int, ok bool) {
	tmux := os.Getenv("TMUX")
	tmuxPane := os.Getenv("TMUX_PANE")
	if tmux == "" || tmuxPane == "" {
		return 0, 0, false
	}

	tmuxPaneOffsetMu.Lock()
	defer tmuxPaneOffsetMu.Unlock()

	if tmuxPaneOffsetCached && time.Since(tmuxPaneOffsetAt) < tmuxPaneOffsetTTL {
		return tmuxPaneOffsetTop, tmuxPaneOffsetLeft, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if idx := strings.Index(tmux, ","); idx > 0 {
		// $TMUX format: socket_path,pid,session_id
		cmd = exec.CommandContext(ctx, "tmux", "-S", tmux[:idx], "display", "-p", "-t", tmuxPane, "#{pane_top},#{pane_left}")
	} else {
		// Socket path extraction failed; fall back to the default socket.
		cmd = exec.CommandContext(ctx, "tmux", "display", "-p", "-t", tmuxPane, "#{pane_top},#{pane_left}")
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, false
	}
	top, left, ok = parseTmuxPaneGeometry(string(output))
	if !ok {
		return 0, 0, false
	}
	tmuxPaneOffsetTop = top
	tmuxPaneOffsetLeft = left
	tmuxPaneOffsetCached = true
	tmuxPaneOffsetAt = time.Now()
	return top, left, true
}

// parseTmuxPaneGeometry parses a "#{pane_top},#{pane_left}" expansion.
// Leading/trailing whitespace (including a trailing newline) is trimmed;
// both fields must be non-negative integers. Returns ok=false on any
// malformed output.
func parseTmuxPaneGeometry(output string) (top, left int, ok bool) {
	line := strings.TrimSpace(output)
	if i := strings.IndexAny(line, "\r\n"); i >= 0 {
		line = line[:i]
	}
	parts := strings.Split(line, ",")
	if len(parts) != 2 {
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
