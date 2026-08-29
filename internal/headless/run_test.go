package headless

import "testing"

// TestParseOnce verifies the --once string split: the first space-separated
// token is the command and the (trimmed) remainder is the query.
func TestParseOnce(t *testing.T) {
	cases := []struct {
		in      string
		wantCmd string
		wantQ   string
	}{
		{"", "", ""},
		{"   ", "", ""},
		{"status", "status", ""},
		{"play", "play", ""},
		{"play 周杰伦", "play", "周杰伦"},
		{"play  周杰伦  晴天  ", "play", "周杰伦  晴天"},
		{"volume 60", "volume", "60"},
		{"seek 30.5", "seek", "30.5"},
	}
	for _, tc := range cases {
		cmd, query := parseOnce(tc.in)
		if cmd != tc.wantCmd || query != tc.wantQ {
			t.Errorf("parseOnce(%q) = (%q, %q), want (%q, %q)", tc.in, cmd, query, tc.wantCmd, tc.wantQ)
		}
	}
}
