package configs

import "testing"

func TestVisualizerCharactersDefaultAndOverride(t *testing.T) {
	halfBlock, fullBlock, emptyBlock := (VisualizerConfig{}).Characters()
	if halfBlock != '▌' || fullBlock != '█' || emptyBlock != ' ' {
		t.Fatalf("default characters = %q %q %q", halfBlock, fullBlock, emptyBlock)
	}

	halfBlock, fullBlock, emptyBlock = (VisualizerConfig{
		FullCharHalfBlock: "~",
		FullCharFullBlock: "@",
		EmptyCharBlock:    ".",
	}).Characters()
	if halfBlock != '~' || fullBlock != '@' || emptyBlock != '.' {
		t.Fatalf("configured characters = %q %q %q", halfBlock, fullBlock, emptyBlock)
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
