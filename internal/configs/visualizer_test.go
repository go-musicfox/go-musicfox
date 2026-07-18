package configs

import "testing"

func TestVisualizerBarCharactersDefaultAndOverride(t *testing.T) {
	halfBlock, fullBlock, emptyBlock := (VisualizerConfig{}).BarCharacters()
	if halfBlock != '▌' || fullBlock != '█' || emptyBlock != ' ' {
		t.Fatalf("default bar characters = %q %q %q", halfBlock, fullBlock, emptyBlock)
	}

	halfBlock, fullBlock, emptyBlock = (VisualizerConfig{
		BarHalfBlock: "~",
		BarFullBlock: "@",
		BarEmptyBlock: ".",
	}).BarCharacters()
	if halfBlock != '~' || fullBlock != '@' || emptyBlock != '.' {
		t.Fatalf("configured bar characters = %q %q %q", halfBlock, fullBlock, emptyBlock)
	}
}

func TestVisualizerMirrorBarCharactersFallback(t *testing.T) {
	// All empty: falls back to defaults.
	half, full, empty := (VisualizerConfig{}).MirrorBarCharacters()
	if half != '▌' || full != '█' || empty != ' ' {
		t.Fatalf("default mirror bar characters = %q %q %q", half, full, empty)
	}

	// Mirror-specific values override.
	half, full, empty = (VisualizerConfig{
		MirrorBarFullBlock: "*",
	}).MirrorBarCharacters()
	if full != '*' {
		t.Fatalf("override mirror bar full = %q, want *", full)
	}

	// Mirror falls back to bar values when mirror values are empty.
	half, full, empty = (VisualizerConfig{
		BarHalfBlock: "A",
		BarFullBlock: "B",
	}).MirrorBarCharacters()
	if half != 'A' || full != 'B' {
		t.Fatalf("mirror bar fallback = %q %q, want A B", half, full)
	}
}

func TestVisualizerMaxBarHeightBounds(t *testing.T) {
	for _, test := range []struct {
		configured int
		want       int
	}{
		{configured: 0, want: 0},
		{configured: 3, want: 3},
		{configured: 12, want: 12},
	} {
		if got := (VisualizerConfig{MaxHeight: test.configured}).MaxBarHeight(); got != test.want {
			t.Fatalf("max height(%d) = %d, want %d", test.configured, got, test.want)
		}
	}
}
