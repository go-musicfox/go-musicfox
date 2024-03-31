package effects

import (
	"math"

	"github.com/gopxl/beep"
)

// TransitionFunc defines a function used in a transition to describe the progression curve
// from one value to the next. The input 'percent' always ranges from 0.0 to 1.0, where 0.0
// represents the starting point and 1.0 represents the end point of the transition.
//
// The returned value from TransitionFunc is expected to be in the normalized range of [0.0, 1.0].
// However, it may exceed this range, providing flexibility to generate curves with momentum.
// The Transition() function then maps this normalized output to the actual desired range.
type TransitionFunc func(percent float64) float64

// TransitionLinear transitions the gain linearly from the start to end value.
func TransitionLinear(percent float64) float64 {
	return percent
}

// TransitionEqualPower transitions the gain of a streamer in such a way that the total perceived volume stays
// constant if mixed together with another streamer doing the inverse transition.
//
// See https://www.oreilly.com/library/view/web-audio-api/9781449332679/ch03.html#s03_2 for more information.
func TransitionEqualPower(percent float64) float64 {
	return math.Cos((1.0 - percent) * 0.5 * math.Pi)
}

// Transition gradually adjusts the gain of the source streamer 's' from 'startGain' to 'endGain'
// over the entire duration of the stream, defined by the number of samples 'len'.
// The transition is defined by the provided 'transitionFunc' function, which determines the
// gain at each point during the transition.
func Transition(s beep.Streamer, len int, startGain, endGain float64, transitionfunc TransitionFunc) *TransitionStreamer {
	return &TransitionStreamer{
		s:              s,
		len:            len,
		startGain:      startGain,
		endGain:        endGain,
		transitionFunc: transitionfunc,
	}
}

type TransitionStreamer struct {
	s                  beep.Streamer
	pos                int
	len                int
	startGain, endGain float64
	transitionFunc     TransitionFunc
}

// Stream fills samples with the gain-adjusted samples of the source streamer.
func (t *TransitionStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = t.s.Stream(samples)

	for i := 0; i < n; i++ {
		pos := t.pos + i
		progress := float64(pos) / float64(t.len)
		if progress < 0 {
			progress = 0
		} else if progress > 1 {
			progress = 1
		}
		value := t.transitionFunc(progress)
		gain := t.startGain + (t.endGain-t.startGain)*value

		samples[i][0] *= gain
		samples[i][1] *= gain
	}

	t.pos += n

	return
}

// Err propagates the original Streamer's errors.
func (t *TransitionStreamer) Err() error {
	return t.s.Err()
}
