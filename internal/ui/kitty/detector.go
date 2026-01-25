package kitty

import (
	"os"
	"strings"
	"sync"
)

var (
	kittySupported    bool
	kittyDetectedOnce sync.Once
)

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
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	kittyWindowID := os.Getenv("KITTY_WINDOW_ID")

	// Check for terminal multiplexers that don't pass through kitty graphics
	// tmux and screen typically don't support kitty graphics passthrough
	tmux := os.Getenv("TMUX")
	sty := os.Getenv("STY") // screen session
	if tmux != "" || sty != "" {
		// Running inside tmux or screen - kitty graphics won't work
		return false
	}

	// Kitty terminal
	if kittyWindowID != "" {
		return true
	}
	if strings.Contains(strings.ToLower(term), "kitty") {
		return true
	}

	// WezTerm - supports Kitty graphics protocol
	if strings.Contains(termProgram, "WezTerm") {
		return true
	}
	weztermExecutable := os.Getenv("WEZTERM_EXECUTABLE")
	if weztermExecutable != "" {
		return true
	}

	// Ghostty - supports Kitty graphics protocol
	if term == "xterm-ghostty" {
		return true
	}
	ghosttyResourcesDir := os.Getenv("GHOSTTY_RESOURCES_DIR")
	if ghosttyResourcesDir != "" {
		return true
	}

	// Konsole (KDE) - supports Kitty graphics protocol since version 22.04
	konsoleVersion := os.Getenv("KONSOLE_VERSION")
	if konsoleVersion != "" {
		return true
	}

	return false
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
