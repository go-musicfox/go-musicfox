package effects

import (
	"math"

	"github.com/faiface/beep"
)

type (

	// This parametric equalizer is based on the GK Nilsen's post at:
	// https://octovoid.com/2017/11/04/coding-a-parametric-equalizer-for-audio-applications/
	equalizer struct {
		streamer beep.Streamer
		sections []section
	}

	section struct {
		a, b         [2][]float64
		xPast, yPast [][2]float64
	}

	// EqualizerSections is the interfacd that is passed into NewEqualizer
	EqualizerSections interface {
		sections(fs float64) []section
	}

	StereoEqualizerSection struct {
		Left  MonoEqualizerSection
		Right MonoEqualizerSection
	}

	MonoEqualizerSection struct {
		// F0 (center frequency) sets the mid-point of the section’s
		// frequency range and is given in Hertz [Hz].
		F0 float64

		// Bf (bandwidth) represents the width of the section across
		// frequency and is measured in Hertz [Hz]. A low bandwidth
		// corresponds to a narrow frequency range meaning that the
		// section will concentrate its operation to only the
		// frequencies close to the center frequency. On the other hand,
		// a high bandwidth yields a section of wide frequency range —
		// affecting a broader range of frequencies surrounding the
		// center frequency.
		Bf float64

		// GB (bandwidth gain) is given in decibels [dB] and represents
		// the level at which the bandwidth is measured. That is, to
		// have a meaningful measure of bandwidth, we must define the
		// level at which it is measured.
		GB float64

		// G0 (reference gain) is given in decibels [dB] and simply
		// represents the level of the section’s offset.
		G0 float64

		// G (boost/cut gain) is given in decibels [dB] and prescribes
		// the effect imposed on the audio loudness for the section’s
		// frequency range. A boost/cut level of 0 dB corresponds to
		// unity (no operation), whereas negative numbers corresponds to
		// cut (volume down) and positive numbers to boost (volume up).
		G float64
	}

	// StereoEqualizerSections implements EqualizerSections and can be passed into NewEqualizer
	StereoEqualizerSections []StereoEqualizerSection

	// MonoEqualizerSections implements EqualizerSections and can be passed into NewEqualizer
	MonoEqualizerSections []MonoEqualizerSection
)

// NewEqualizer returns a beep.Streamer that modifies the stream based on the EqualizerSection slice that is passed in.
// The SampleRate (sr) must match that of the Streamer.
func NewEqualizer(st beep.Streamer, sr beep.SampleRate, s EqualizerSections) beep.Streamer {
	return &equalizer{
		streamer: st,
		sections: s.sections(float64(sr)),
	}
}

func (m MonoEqualizerSections) sections(fs float64) []section {
	out := make([]section, len(m))
	for i, s := range m {
		out[i] = s.section(fs)
	}
	return out
}

func (m StereoEqualizerSections) sections(fs float64) []section {
	out := make([]section, len(m))
	for i, s := range m {
		out[i] = s.section(fs)
	}
	return out
}

// Stream streams the wrapped Streamer modified by Equalizer.
func (e *equalizer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = e.streamer.Stream(samples)
	for _, s := range e.sections {
		s.apply(samples)
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (e *equalizer) Err() error {
	return e.streamer.Err()
}

func (m MonoEqualizerSection) section(fs float64) section {
	beta := math.Tan(m.Bf/2.0*math.Pi/(fs/2.0)) *
		math.Sqrt(math.Abs(math.Pow(math.Pow(10, m.GB/20.0), 2.0)-
			math.Pow(math.Pow(10.0, m.G0/20.0), 2.0))) /
		math.Sqrt(math.Abs(math.Pow(math.Pow(10.0, m.G/20.0), 2.0)-
			math.Pow(math.Pow(10.0, m.GB/20.0), 2.0)))

	b := []float64{
		(math.Pow(10.0, m.G0/20.0) + math.Pow(10.0, m.G/20.0)*beta) / (1 + beta),
		(-2 * math.Pow(10.0, m.G0/20.0) * math.Cos(m.F0*math.Pi/(fs/2.0))) / (1 + beta),
		(math.Pow(10.0, m.G0/20) - math.Pow(10.0, m.G/20.0)*beta) / (1 + beta),
	}

	a := []float64{
		1.0,
		-2 * math.Cos(m.F0*math.Pi/(fs/2.0)) / (1 + beta),
		(1 - beta) / (1 + beta),
	}

	return section{
		a: [2][]float64{a, a},
		b: [2][]float64{b, b},
	}
}

func (s StereoEqualizerSection) section(fs float64) section {
	l := s.Left.section(fs)
	r := s.Right.section(fs)

	return section{
		a: [2][]float64{l.a[0], r.a[0]},
		b: [2][]float64{l.b[0], r.b[0]},
	}
}

func (s *section) apply(x [][2]float64) {
	ord := len(s.a[0]) - 1
	np := len(x) - 1

	if np < ord {
		x = append(x, make([][2]float64, ord-np)...)
		np = ord
	}

	y := make([][2]float64, len(x))

	if len(s.xPast) < len(x) {
		s.xPast = append(s.xPast, make([][2]float64, len(x)-len(s.xPast))...)
	}

	if len(s.yPast) < len(x) {
		s.yPast = append(s.yPast, make([][2]float64, len(x)-len(s.yPast))...)
	}

	for i := 0; i < len(x); i++ {
		for j := 0; j < ord+1; j++ {
			if i-j < 0 {
				y[i][0] = y[i][0] + s.b[0][j]*s.xPast[len(s.xPast)-j][0]
				y[i][1] = y[i][1] + s.b[1][j]*s.xPast[len(s.xPast)-j][1]
			} else {
				y[i][0] = y[i][0] + s.b[0][j]*x[i-j][0]
				y[i][1] = y[i][1] + s.b[1][j]*x[i-j][1]
			}
		}

		for j := 0; j < ord; j++ {
			if i-j-1 < 0 {
				y[i][0] = y[i][0] - s.a[0][j+1]*s.yPast[len(s.yPast)-j-1][0]
				y[i][1] = y[i][1] - s.a[1][j+1]*s.yPast[len(s.yPast)-j-1][1]
			} else {
				y[i][0] = y[i][0] - s.a[0][j+1]*y[i-j-1][0]
				y[i][1] = y[i][1] - s.a[1][j+1]*y[i-j-1][1]
			}
		}

		y[i][0] = y[i][0] / s.a[0][0]
		y[i][1] = y[i][1] / s.a[1][0]
	}

	s.xPast = x[:]
	s.yPast = y[:]
	copy(x, y)
}
