package ui

import "testing"

func TestCoverBackgroundExclusionLeavesBottomTerminalRowPainted(t *testing.T) {
	_, _, _, height := coverBackgroundExclusion(8, 12, 24, 6)
	if got, want := height, 5; got != want {
		t.Fatalf("cover background exclusion height = %d, want %d so the bottom terminal row is painted", got, want)
	}
}
