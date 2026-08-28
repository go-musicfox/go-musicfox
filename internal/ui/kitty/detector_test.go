package kitty

import (
	"testing"
)

// clearTerminalEnv clears all environment variables that participate in the
// kitty graphics detection, so tests are deterministic regardless of the
// environment the test suite runs in.
func clearTerminalEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"TMUX", "STY", "TERM", "TERM_PROGRAM",
		"KITTY_WINDOW_ID", "WEZTERM_EXECUTABLE",
		"GHOSTTY_RESOURCES_DIR", "KONSOLE_VERSION",
	} {
		t.Setenv(key, "")
	}
}

// resetDetectionState resets package-level detection state. IsSupported caches
// its result in a sync.Once which cannot be reset, so tests call
// detectKittySupport directly and reset tmuxPassthrough manually.
func resetDetectionState(t *testing.T) {
	t.Helper()
	origPassthrough := tmuxPassthrough
	origSupported := kittySupported
	tmuxPassthrough = false
	kittySupported = false
	t.Cleanup(func() {
		tmuxPassthrough = origPassthrough
		kittySupported = origSupported
	})
}

func TestDetectKittySupportInsideTmuxWithGhostty(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-nonexistent-socket,123,0")
	t.Setenv("TERM", "tmux-256color")
	t.Setenv("TERM_PROGRAM", "ghostty")
	resetDetectionState(t)

	if !detectKittySupport() {
		t.Error("expected kitty support inside tmux with ghostty outer terminal")
	}
	if !UseTmuxPassthrough() {
		t.Error("expected tmux passthrough to be enabled")
	}
}

func TestDetectKittySupportInsideTmuxWithUnsupportedTerminal(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-nonexistent-socket,123,0")
	t.Setenv("TERM", "tmux-256color")
	resetDetectionState(t)

	if detectKittySupport() {
		t.Error("expected no kitty support inside tmux with unsupported outer terminal")
	}
	if UseTmuxPassthrough() {
		t.Error("expected tmux passthrough to stay disabled")
	}
}

func TestDetectKittySupportInsideScreen(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("STY", "12345.pts-0.host")
	t.Setenv("TERM", "screen-256color")
	t.Setenv("TERM_PROGRAM", "ghostty")
	resetDetectionState(t)

	if detectKittySupport() {
		t.Error("expected no kitty support inside GNU screen")
	}
	if UseTmuxPassthrough() {
		t.Error("expected tmux passthrough to stay disabled")
	}
}

func TestDetectKittySupportDirectGhostty(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "xterm-ghostty")
	resetDetectionState(t)

	if !detectKittySupport() {
		t.Error("expected kitty support in direct ghostty terminal")
	}
	if UseTmuxPassthrough() {
		t.Error("expected tmux passthrough to stay disabled outside tmux")
	}
}

func TestDetectKittySupportDirectUnsupported(t *testing.T) {
	clearTerminalEnv(t)
	t.Setenv("TERM", "xterm-256color")
	resetDetectionState(t)

	if detectKittySupport() {
		t.Error("expected no kitty support in unsupported direct terminal")
	}
	if UseTmuxPassthrough() {
		t.Error("expected tmux passthrough to stay disabled outside tmux")
	}
}

func TestDetectDirectTerminalSupportTable(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"kitty window id", map[string]string{"KITTY_WINDOW_ID": "1"}, true},
		{"kitty term", map[string]string{"TERM": "xterm-kitty"}, true},
		{"kitty term program", map[string]string{"TERM_PROGRAM": "kitty"}, true},
		{"wezterm term program", map[string]string{"TERM_PROGRAM": "WezTerm"}, true},
		{"wezterm executable", map[string]string{"WEZTERM_EXECUTABLE": "/usr/bin/wezterm"}, true},
		{"ghostty term", map[string]string{"TERM": "xterm-ghostty"}, true},
		{"ghostty term program", map[string]string{"TERM_PROGRAM": "ghostty"}, true},
		{"ghostty resources dir", map[string]string{"GHOSTTY_RESOURCES_DIR": "/opt/ghostty"}, true},
		{"konsole", map[string]string{"KONSOLE_VERSION": "220403"}, true},
		{"unknown terminal", map[string]string{"TERM": "xterm-256color", "TERM_PROGRAM": "iTerm.app"}, false},
		{"empty env", map[string]string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(k string) string { return tt.env[k] }
			if got := detectDirectTerminalSupport(getenv); got != tt.want {
				t.Errorf("detectDirectTerminalSupport(%v) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}

func TestParseTmuxShowEnvironment(t *testing.T) {
	output := "TERM=xterm-ghostty\n" +
		"TERM_PROGRAM=ghostty\n" +
		"-KITTY_WINDOW_ID\n" +
		"\n" +
		"COLORFGBG=15;0\n" +
		"PATH_WITH_EQ=A=B=C\n" +
		"NOEQUALS\n"
	env := parseTmuxShowEnvironment(output)

	want := map[string]string{
		"TERM":         "xterm-ghostty",
		"TERM_PROGRAM": "ghostty",
		"COLORFGBG":    "15;0",
		"PATH_WITH_EQ": "A=B=C",
	}
	if len(env) != len(want) {
		t.Fatalf("parsed %d entries, want %d: %v", len(env), len(want), env)
	}
	for k, v := range want {
		if env[k] != v {
			t.Errorf("env[%q] = %q, want %q", k, env[k], v)
		}
	}
}
