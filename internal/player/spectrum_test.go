package player

import (
	"math"
	"testing"
	"time"

	"github.com/charmbracelet/harmonica"
)

func TestPCMAnalyzerDetectsSineWave(t *testing.T) {
	analyzer := NewPCMAnalyzer(100 * time.Millisecond)
	defer analyzer.Close()

	const (
		sampleRate = 44100.0
		frequency  = 440.0
	)
	samples := make([]float32, spectrumFFTSize)
	for index := range samples {
		samples[index] = float32(math.Sin(2 * math.Pi * frequency * float64(index) / sampleRate))
	}
	analyzer.NewConsumer()(sampleRate, samples)

	expectedBand := int(math.Log(frequency/60) / math.Log(16000.0/60) * SpectrumBandCount)
	deadline := time.After(time.Second)
	for {
		frame := analyzer.Spectrum()
		peakBand, peakLevel := 0, 0.0
		for index, level := range frame.Levels {
			if level > peakLevel {
				peakBand, peakLevel = index, level
			}
		}
		if peakLevel > 0 {
			if distance := abs(peakBand - expectedBand); distance > 2 {
				t.Fatalf("peak band = %d, want near %d", peakBand, expectedBand)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("analyzer did not publish a spectrum frame")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestPCMAnalyzerClearsFrameForNewConsumer(t *testing.T) {
	analyzer := NewPCMAnalyzer(100 * time.Millisecond)
	defer analyzer.Close()

	consumer := analyzer.NewConsumer()
	consumer(44100, []float32{1, 1, 1, 1})
	consumer = analyzer.NewConsumer()
	consumer(44100, nil)

	frame := analyzer.Spectrum()
	for _, level := range frame.Levels {
		if level != 0 {
			t.Fatalf("new consumer left stale level %f", level)
		}
	}
}

func TestPCMAnalyzerSpringApproachesWithoutOvershoot(t *testing.T) {
	analyzer := NewPCMAnalyzer(100 * time.Millisecond)
	defer analyzer.Close()

	target := SpectrumFrame{}
	target.Levels[0] = 1
	previous := 0.0
	for frame := 0; frame < 10; frame++ {
		level := analyzer.springFrame(target).Levels[0]
		if level < previous || level > target.Levels[0] {
			t.Fatalf("frame %d level = %f, previous = %f, target = %f", frame, level, previous, target.Levels[0])
		}
		previous = level
	}
	if previous <= 0 || previous >= target.Levels[0] {
		t.Fatalf("level = %f, want a fractional value between 0 and %f", previous, target.Levels[0])
	}
}

func TestPCMAnalyzerAdvancesSpringWithoutNewPCM(t *testing.T) {
	analyzer := NewPCMAnalyzer(time.Hour)
	defer analyzer.Close()
	analyzer.spring = harmonica.NewSpring(harmonica.FPS(60), 9, 1)
	analyzer.target.Levels[0] = 1

	analyzer.analyzeLatest()
	first := analyzer.Spectrum().Levels[0]
	analyzer.analyzeLatest()
	second := analyzer.Spectrum().Levels[0]
	if first <= 0 || second <= first || second >= 1 {
		t.Fatalf("spring levels = %f, %f, want increasing fractional values", first, second)
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
