package desktop_lyrics

import "testing"

func TestLyricLabelHeight(t *testing.T) {
	for _, tc := range []struct {
		size              float64
		primary, fallback fontInk
		large             bool
		want              float64
	}{
		{24, fontInk{36, 2}, fontInk{20, 10}, false, 50},
		{24, fontInk{36, 2}, fontInk{20, 10}, true, 100},
		{24, fontInk{10, 2}, fontInk{10, 2}, false, 32},
		{24, fontInk{10, 2}, fontInk{10, 2}, true, 64},
		{24, fontInk{10, 2}, fontInk{}, false, 32},
		// Fallback dominates primary in both directions (e.g. Helvetica Neue + PingFang SC CJK).
		{24, fontInk{20, 6}, fontInk{26, 9}, false, 39},
		{24, fontInk{20, 6}, fontInk{26, 9}, true, 78},
	} {
		if got := lyricLabelHeight(tc.size, 4, tc.primary, tc.fallback, tc.large); got != tc.want {
			t.Errorf("case %+v: got %v, want %v", tc, got, tc.want)
		}
	}
}
