package effects

import "github.com/faiface/beep"

// Mono converts the wrapped Streamer to a mono buffer
// by downmixing the left and right channels together.
//
// The returned Streamer propagates s's errors through Err.
func Mono(s beep.Streamer) beep.Streamer {
	return &mono{s}
}

type mono struct {
	Streamer beep.Streamer
}

func (m *mono) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = m.Streamer.Stream(samples)
	for i := range samples[:n] {
		mix := (samples[i][0] + samples[i][1]) / 2
		samples[i][0], samples[i][1] = mix, mix
	}
	return n, ok
}

func (m *mono) Err() error {
	return m.Streamer.Err()
}
