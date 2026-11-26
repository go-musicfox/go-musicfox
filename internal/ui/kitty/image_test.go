package kitty

import "testing"

func TestImageTargetSize(t *testing.T) {
	tests := []struct {
		name string
		spin bool
		tmux bool
		want int
	}{
		{name: "direct static", want: 512},
		{name: "tmux static", tmux: true, want: DefaultRotationSize},
		{name: "spin", spin: true, want: DefaultRotationSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imageTargetSize(tt.spin, tt.tmux); got != tt.want {
				t.Fatalf("imageTargetSize(%v, %v) = %d, want %d", tt.spin, tt.tmux, got, tt.want)
			}
		})
	}
}
