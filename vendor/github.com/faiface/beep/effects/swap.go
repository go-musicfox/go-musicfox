package effects

import "github.com/faiface/beep"

// Swap swaps the left and right channel of the wrapped Streamer.
//
// The returned Streamer propagates s's errors through Err.
func Swap(s beep.Streamer) beep.Streamer {
	return &swap{s}
}

type swap struct {
	Streamer beep.Streamer
}

func (s *swap) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = s.Streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0], samples[i][1] = samples[i][1], samples[i][0]
	}
	return n, ok
}

func (s *swap) Err() error {
	return s.Streamer.Err()
}
