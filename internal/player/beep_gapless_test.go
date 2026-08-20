package player

import (
	"bytes"
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

func TestStreamAcrossBoundaryKeepsExactBufferBoundaryAlive(t *testing.T) {
	current := &sampleStreamer{samples: [][2]float64{{1, 1}, {2, 2}}}
	next := &sampleStreamer{samples: [][2]float64{{3, 3}}}
	buffer := make([][2]float64, 2)

	n, ok, switched := streamAcrossBoundary(buffer, current, func() beep.Streamer { return next })
	if n != len(buffer) || !ok || !switched {
		t.Fatalf("n=%d ok=%v switched=%v, want exact-boundary switch", n, ok, switched)
	}
	if next.pos != 0 {
		t.Fatalf("next stream advanced to %d, want it untouched until the next callback", next.pos)
	}
}

func TestTrimmedMP3RemovesEncoderPadding(t *testing.T) {
	raw := &sampleSeekStreamer{samples: [][2]float64{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}}}
	stream := &trimmedMP3{StreamSeekCloser: raw, start: 1, end: 2, trimEnd: true}
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

func TestTrimmedMP3DoesNotUseFixedLengthWhileStreaming(t *testing.T) {
	raw := &sampleSeekStreamer{samples: [][2]float64{{0, 0}, {1, 1}, {2, 2}, {3, 3}}}
	stream := &trimmedMP3{StreamSeekCloser: raw, start: 1, end: 2}
	buffer := make([][2]float64, 3)
	n, ok := stream.Stream(buffer)
	if n != 3 || ok {
		t.Fatalf("n=%d ok=%v, want 3 false at the underlying EOF", n, ok)
	}
	if stream.Len() != 3 {
		t.Fatalf("streaming logical length=%d, want underlying length minus start", stream.Len())
	}
}

func TestEstimateMP3PaddingIsConservative(t *testing.T) {
	const rate = beep.SampleRate(44100)
	maxPadding := int(maxEstimatedMP3Padding.Seconds() * float64(rate))
	for _, test := range []struct {
		name        string
		decoded     int
		target      int
		wantPadding int
	}{
		{name: "typical padding", decoded: 101000, target: 100000, wantPadding: 1000},
		{name: "maximum padding", decoded: 100000 + maxPadding, target: 100000, wantPadding: maxPadding},
		{name: "duration mismatch", decoded: 100001 + maxPadding, target: 100000, wantPadding: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := estimateMP3Padding(test.decoded, test.target, rate); got != test.wantPadding {
				t.Fatalf("padding=%d, want %d", got, test.wantPadding)
			}
		})
	}
}

func TestMP3GaplessSamplesTreatsExplicitZeroPaddingAsReliable(t *testing.T) {
	lame := make([]byte, 64)
	copy(lame[4:], "LAME")
	copy(lame[4+21:], []byte{0x45, 0x10, 0x00}) // delay=0x451, padding=0

	for _, test := range []struct {
		name string
		data []byte
	}{
		{name: "LAME", data: lame},
		{name: "iTunSMPB", data: []byte("iTunSMPB 00000000 00000451 00000000")},
	} {
		t.Run(test.name, func(t *testing.T) {
			delay, padding, hasPadding := mp3GaplessSamples(bytes.NewReader(test.data))
			if delay != 0x451 || padding != 0 || !hasPadding {
				t.Fatalf("delay=%d padding=%d hasPadding=%v", delay, padding, hasPadding)
			}
		})
	}
}

func TestMP3GaplessSamplesRejectsInvalidITunSMPB(t *testing.T) {
	delay, padding, hasPadding := mp3GaplessSamples(bytes.NewReader([]byte("iTunSMPB 00000000 invalid invalid")))
	if delay != 0 || padding != 0 || hasPadding {
		t.Fatalf("delay=%d padding=%d hasPadding=%v, want invalid metadata ignored", delay, padding, hasPadding)
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
