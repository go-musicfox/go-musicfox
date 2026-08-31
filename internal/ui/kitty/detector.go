package kitty

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
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

// SetTmuxPassthroughForTest overrides the package-level tmux passthrough
// mode. Test-only: it mutates global package state, so callers must restore
// the previous value (typically via t.Cleanup) and must not use
// t.Parallel() in tests that touch it.
func SetTmuxPassthroughForTest(on bool) {
	tmuxPassthrough = on
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
		// `set -g allow-passthrough on`, which is verified below).
		if detectDirectTerminalSupport(os.Getenv) {
			return enableTmuxPassthrough()
		}
		// tmux 3.4+ overwrites pane environment variables (TERM=tmux-256color,
		// TERM_PROGRAM=tmux), so pane env may not identify the outer terminal.
		// Fall back to the session-level environment (`tmux show-environment`),
		// which tracks the attaching client and does.
		if m := tmuxSessionEnv(); m != nil {
			if detectDirectTerminalSupport(func(k string) string { return m[k] }) {
				return enableTmuxPassthrough()
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
// does. Note: when multiple clients are attached to the same session, the
// session environment reflects the most recently attaching client's outer
// terminal and may therefore misdetect; the pane environment probe above
// takes precedence, so this is only a fallback. Never panics.
func tmuxSessionEnv() map[string]string {
	tmux := os.Getenv("TMUX")
	if tmux == "" {
		return nil
	}

	output, err := tmuxExec(context.Background(), "show-environment")
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

// enableTmuxPassthrough turns on tmux DCS passthrough mode after verifying
// that the running tmux server actually has `allow-passthrough on`: tmux 3.3+
// defaults it to off, and with it off tmux silently drops every passthrough
// packet — claiming support anyway would make the cover image fail silently
// and frustrate debugging. Fail-safe: on query failure (including tmux
// versions without the option) or an off/unknown value, passthrough stays
// disabled, matching the pre-passthrough behavior. Caller must be inside
// tmux with a supported outer terminal.
func enableTmuxPassthrough() bool {
	if !tmuxAllowPassthroughEnabled() {
		return false
	}
	tmuxPassthrough = true
	return true
}

// tmuxAllowPassthroughEnabled reports whether the tmux server has
// `allow-passthrough on` ("1"/"on"). It is a variable so tests can override
// the subprocess-backed probe.
var tmuxAllowPassthroughEnabled = func() bool {
	output, err := tmuxExec(context.Background(), "display", "-p", "#{allow-passthrough}")
	if err != nil {
		slog.Debug("kitty: tmux allow-passthrough query failed, disabling passthrough", slog.Any("error", err))
		return false
	}
	value := strings.TrimSpace(string(output))
	if value != "1" && value != "on" {
		slog.Debug("kitty: tmux allow-passthrough is not enabled, disabling passthrough", slog.String("value", value))
		return false
	}
	return true
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
