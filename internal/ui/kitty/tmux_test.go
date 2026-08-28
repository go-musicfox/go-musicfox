package kitty

import (
	"strings"
	"testing"
)

func TestParseTmuxPaneGeometry(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantTop    int
		wantLeft   int
		wantOK     bool
	}{
		{"normal", "3,74", 3, 74, true},
		{"zeroes", "0,0", 0, 0, true},
		{"with spaces", " 3 , 74 ", 3, 74, true},
		{"trailing newline", "3,74\n", 3, 74, true},
		{"crlf", "3,74\r\n", 3, 74, true},
		{"non numeric", "abc,74", 0, 0, false},
		{"missing column", "3", 0, 0, false},
		{"extra column", "3,74,1", 0, 0, false},
		{"negative", "-1,74", 0, 0, false},
		{"empty", "", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top, left, ok := parseTmuxPaneGeometry(tt.output)
			if ok != tt.wantOK {
				t.Fatalf("parseTmuxPaneGeometry(%q) ok = %v, want %v", tt.output, ok, tt.wantOK)
			}
			if ok && (top != tt.wantTop || left != tt.wantLeft) {
				t.Errorf("parseTmuxPaneGeometry(%q) = (%d, %d), want (%d, %d)", tt.output, top, left, tt.wantTop, tt.wantLeft)
			}
		})
	}
}

func TestBuildTmuxPositionedPayload(t *testing.T) {
	// pane_top=0, pane_left=74, in-pane start row 5 / col 1 (1-based) must map
	// to outer-terminal absolute row 5 / col 75.
	payload := BuildTmuxPositionedPayload(0, 74, 5, 1, "IMGSEQ", false)
	want := "\x1b7\x1b[5;75HIMGSEQ\x1b8"
	if payload != want {
		t.Errorf("payload = %q, want %q", payload, want)
	}

	// deleteOld prefixes the delete-all APC command.
	withDelete := BuildTmuxPositionedPayload(2, 0, 1, 1, "IMGSEQ", true)
	wantDelete := "\x1b_Ga=d,d=a,q=2\x1b\\\x1b7\x1b[3;1HIMGSEQ\x1b8"
	if withDelete != wantDelete {
		t.Errorf("payload with delete = %q, want %q", withDelete, wantDelete)
	}

	// The payload is bare (no DCS tmux; wrapper); callers wrap it exactly once.
	setTmuxPassthrough(t, true)
	if strings.Contains(payload, "\x1bPtmux;") {
		t.Errorf("payload must be bare, got %q", payload)
	}
	wrapped := Wrap(payload)
	if !strings.HasPrefix(wrapped, "\x1bPtmux;") || !strings.HasSuffix(wrapped, "\x1b\\") {
		t.Errorf("wrapped payload = %q, want DCS tmux; passthrough", wrapped)
	}
}
