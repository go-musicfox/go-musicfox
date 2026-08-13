package player

import (
	"testing"

	"github.com/gopxl/beep"
)

func TestStreamAcrossBoundaryFillsSameBufferWithoutSilence(t *testing.T) {
	current := &sampleStreamer{samples: [][2]float64{{1, 1}, {2, 2}}}
	next := &sampleStreamer{samples: [][2]float64{{3, 3}, {4, 4}, {5, 5}}}
	buffer := make([][2]float64, 4)

	n, ok, switched := streamAcrossBoundary(buffer, current, func() beep.Streamer { return next })
	if n != 4 || !ok || !switched {
		t.Fatalf("n=%d ok=%v switched=%v", n, ok, switched)
	}
	for i, want := range []float64{1, 2, 2, 4} {
		if buffer[i][0] != want || buffer[i][1] != want {
			t.Fatalf("sample %d = %v, want {%v %v}", i, buffer[i], want, want)
		}
	}
}

func TestTrimmedMP3RemovesEncoderPadding(t *testing.T) {
	raw := &sampleSeekStreamer{samples: [][2]float64{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}}}
	stream := &trimmedMP3{StreamSeekCloser: raw, start: 1, end: 2}
	buffer := make([][2]float64, 3)
	n, ok := stream.Stream(buffer)
	if n != 3 || ok {
		t.Fatalf("n=%d ok=%v", n, ok)
	}
	for i, want := range []float64{1, 2, 3} {
		if buffer[i][0] != want {
			t.Fatalf("sample %d = %v, want %v", i, buffer[i][0], want)
		}
	}
	if stream.Len() != 3 || stream.Position() != 3 {
		t.Fatalf("len=%d position=%d", stream.Len(), stream.Position())
	}
	if err := stream.Seek(1); err != nil {
		t.Fatal(err)
	}
	var one [1][2]float64
	if n, _ := stream.Stream(one[:]); n != 1 || one[0][0] != 2 {
		t.Fatalf("seeked sample = %v", one[0][0])
	}
}

func TestDeClickBoundarySmoothsOnlyTheBoundary(t *testing.T) {
	samples := [][2]float64{{0, 0}, {1, 1}, {10, 10}, {10, 10}, {10, 10}}
	deClickBoundary(samples, 2, 3)
	if samples[2][0] != 1 {
		t.Fatalf("first next sample = %v, want 1", samples[2][0])
	}
	if samples[3][0] != 5.5 || samples[4][0] != 10 {
		t.Fatalf("boundary correction = %v, %v, want 5.5, 10", samples[3][0], samples[4][0])
	}
}

type sampleStreamer struct {
	samples [][2]float64
	pos     int
}

func (s *sampleStreamer) Stream(dst [][2]float64) (int, bool) {
	n := copy(dst, s.samples[s.pos:])
	s.pos += n
	return n, s.pos < len(s.samples)
}

func (s *sampleStreamer) Err() error { return nil }

var _ beep.Streamer = (*sampleStreamer)(nil)

type sampleSeekStreamer struct {
	samples [][2]float64
	pos     int
}

func (s *sampleSeekStreamer) Stream(dst [][2]float64) (int, bool) {
	n := copy(dst, s.samples[s.pos:])
	s.pos += n
	return n, s.pos < len(s.samples)
}

func (s *sampleSeekStreamer) Err() error         { return nil }
func (s *sampleSeekStreamer) Len() int           { return len(s.samples) }
func (s *sampleSeekStreamer) Position() int      { return s.pos }
func (s *sampleSeekStreamer) Seek(pos int) error { s.pos = pos; return nil }
func (s *sampleSeekStreamer) Close() error       { return nil }

var _ beep.StreamSeekCloser = (*sampleSeekStreamer)(nil)
