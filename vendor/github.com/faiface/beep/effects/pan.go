package effects

import "github.com/faiface/beep"

// Pan balances the wrapped Streamer between the left and the right channel. The Pan field value of
// -1 means that both original channels go through the left channel. The value of +1 means the same
// for the right channel. The value of 0 changes nothing.
type Pan struct {
	Streamer beep.Streamer
	Pan      float64
}

// Stream streams the wrapped Streamer balanced by Pan.
func (p *Pan) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = p.Streamer.Stream(samples)
	switch {
	case p.Pan < 0:
		for i := range samples[:n] {
			r := samples[i][1]
			samples[i][0] += -p.Pan * r
			samples[i][1] -= -p.Pan * r
		}
	case p.Pan > 0:
		for i := range samples[:n] {
			l := samples[i][0]
			samples[i][0] -= p.Pan * l
			samples[i][1] += p.Pan * l
		}
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (p *Pan) Err() error {
	return p.Streamer.Err()
}
