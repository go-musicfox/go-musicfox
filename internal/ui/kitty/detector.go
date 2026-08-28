package kitty

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	kittySupported    bool
	kittyDetectedOnce sync.Once
	// tmuxPassthrough indicates that Kitty graphics sequences must be wrapped
	// in tmux DCS passthrough (\x1bPtmux;...\x1b\\) because we are running
	// inside tmux. Requires tmux >= 3.3 with `set -g allow-passthrough on`.
	tmuxPassthrough bool
)

// UseTmuxPassthrough returns whether Kitty graphics sequences should be
// wrapped in tmux DCS passthrough.
func UseTmuxPassthrough() bool {
	return tmuxPassthrough
}

// IsSupported returns whether the current terminal supports the Kitty graphics protocol.
// The result is cached after the first call.
func IsSupported() bool {
	kittyDetectedOnce.Do(func() {
		kittySupported = detectKittySupport()
	})
	return kittySupported
}

// detectKittySupport checks environment variables to determine if the terminal
// supports the Kitty graphics protocol.
func detectKittySupport() bool {
	tmux := os.Getenv("TMUX")
	sty := os.Getenv("STY") // screen session

	if sty != "" {
		// GNU screen's DCS passthrough is too limited (string sequences are
		// capped and unreliable), so kitty graphics are not supported.
		return false
	}

	if tmux != "" {
		// Inside tmux: TERM usually reports tmux-256color and is unreliable.
		// Probe the outer terminal instead. First try the pane environment
		// (multiplexer vars like KITTY_WINDOW_ID / TERM_PROGRAM /
		// WEZTERM_EXECUTABLE are typically inherited from the outer terminal).
		// If the outer terminal supports kitty graphics, sequences are wrapped
		// in DCS tmux; passthrough (requires tmux >= 3.3 with
		// `set -g allow-passthrough on`).
		if detectDirectTerminalSupport(os.Getenv) {
			tmuxPassthrough = true
			return true
		}
		// tmux 3.4+ overwrites pane environment variables (TERM=tmux-256color,
		// TERM_PROGRAM=tmux), so pane env may not identify the outer terminal.
		// Fall back to the session-level environment (`tmux show-environment`),
		// which tracks the attaching client and does.
		if m := tmuxSessionEnv(); m != nil {
			if detectDirectTerminalSupport(func(k string) string { return m[k] }) {
				tmuxPassthrough = true
				return true
			}
		}
		return false
	}

	// Not inside tmux or screen: direct terminal detection.
	tmuxPassthrough = false
	return detectDirectTerminalSupport(os.Getenv)
}

// detectDirectTerminalSupport checks environment variables (read via getenv)
// to determine if the directly attached terminal (outside of tmux/screen)
// supports the Kitty graphics protocol. Environment variables set by the
// terminal itself (KITTY_WINDOW_ID, TERM_PROGRAM, etc.) are usually inherited
// by the tmux server, so this also works to identify the outer terminal
// inside tmux.
func detectDirectTerminalSupport(getenv func(string) string) bool {
	term := getenv("TERM")
	termProgram := getenv("TERM_PROGRAM")
	kittyWindowID := getenv("KITTY_WINDOW_ID")

	// Kitty terminal
	if kittyWindowID != "" {
		return true
	}
	if strings.Contains(strings.ToLower(term), "kitty") {
		return true
	}
	if strings.Contains(strings.ToLower(termProgram), "kitty") {
		return true
	}

	// WezTerm - supports Kitty graphics protocol
	if strings.Contains(termProgram, "WezTerm") {
		return true
	}
	weztermExecutable := getenv("WEZTERM_EXECUTABLE")
	if weztermExecutable != "" {
		return true
	}

	// Ghostty - supports Kitty graphics protocol
	if term == "xterm-ghostty" {
		return true
	}
	if termProgram == "ghostty" {
		return true
	}
	ghosttyResourcesDir := getenv("GHOSTTY_RESOURCES_DIR")
	if ghosttyResourcesDir != "" {
		return true
	}

	// Konsole (KDE) - supports Kitty graphics protocol since version 22.04
	konsoleVersion := getenv("KONSOLE_VERSION")
	if konsoleVersion != "" {
		return true
	}

	return false
}

// tmuxSessionEnv queries the tmux session-level environment via
// `tmux show-environment` and returns it as a name/value map, or nil if we
// are not inside tmux or the query fails. tmux 3.4+ sets TERM_PROGRAM=tmux in
// pane environments, so pane env (os.Getenv) can't identify the outer
// terminal; the session-level environment tracks the attaching client and
// does. Never panics.
func tmuxSessionEnv() map[string]string {
	tmux := os.Getenv("TMUX")
	if tmux == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// $TMUX format: socket_path,pid,session_id
	var cmd *exec.Cmd
	if idx := strings.Index(tmux, ","); idx > 0 {
		cmd = exec.CommandContext(ctx, "tmux", "-S", tmux[:idx], "show-environment")
	} else {
		// Socket path extraction failed; fall back to the default socket.
		cmd = exec.CommandContext(ctx, "tmux", "show-environment")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseTmuxShowEnvironment(string(output))
}

// parseTmuxShowEnvironment parses the output of `tmux show-environment`.
// Lines starting with "-" mark variables removed from the session environment
// and are skipped; remaining lines are split on the first "=" into
// name/value pairs. Empty lines are skipped.
func parseTmuxShowEnvironment(output string) map[string]string {
	env := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		env[line[:idx]] = line[idx+1:]
	}
	return env
}

// ForceEnable forces kitty support to be enabled (useful for testing).
func ForceEnable() {
	kittyDetectedOnce.Do(func() {})
	kittySupported = true
}

// ForceDisable forces kitty support to be disabled (useful for testing).
func ForceDisable() {
	kittyDetectedOnce.Do(func() {})
	kittySupported = false
}
