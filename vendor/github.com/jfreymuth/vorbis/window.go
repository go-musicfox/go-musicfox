package vorbis

import "math"

type windowType struct {
	size, prev, next int
}

func (d *Decoder) applyWindow(t *windowType, samples [][]float32) {
	center := t.size / 2
	prevOffset := t.size/4 - t.prev/4
	nextOffset := t.size/4 - t.next/4
	var prevType, nextType int
	if t.prev == d.blocksize[1] {
		prevType = 1
	}
	if t.next == d.blocksize[1] {
		nextType = 1
	}
	for ch := range samples {
		s := samples[ch][:prevOffset]
		for i := range s {
			s[i] = 0
		}
		s = samples[ch][prevOffset : prevOffset+t.prev/2]
		w := d.windows[prevType][:len(s)]
		for i := range s {
			s[i] *= w[i]
		}
		s = samples[ch][center+nextOffset : center+nextOffset+t.next/2]
		w = d.windows[nextType][t.next/2:]
		w = w[:len(s)]
		for i := range s {
			s[i] *= w[i]
		}
		s = samples[ch][t.size-nextOffset:]
		for i := range s {
			s[i] = 0
		}
	}
}

func makeWindow(size int) []float32 {
	window := make([]float32, size)
	for i := range window {
		window[i] = windowFunc((float32(i) + .5) / float32(size/2) * math.Pi / 2)
	}
	return window
}

func windowFunc(x float32) float32 {
	sinx := math.Sin(float64(x))
	return float32(math.Sin(math.Pi / 2 * sinx * sinx))
}
