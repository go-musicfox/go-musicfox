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
	analyzer.NewConsumer()(sampleRate, samples, nil)

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
	consumer(44100, []float32{1, 1, 1, 1}, nil)
	consumer = analyzer.NewConsumer()
	consumer(44100, nil, nil)

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
	target.LevelsL[0] = 1
	previous := 0.0
	for frame := 0; frame < 10; frame++ {
		level := analyzer.springFrame(target).LevelsL[0]
		if level < previous || level > target.LevelsL[0] {
			t.Fatalf("frame %d level = %f, previous = %f, target = %f", frame, level, previous, target.LevelsL[0])
		}
		previous = level
	}
	if previous <= 0 || previous >= target.LevelsL[0] {
		t.Fatalf("level = %f, want a fractional value between 0 and %f", previous, target.LevelsL[0])
	}
}

func TestPCMAnalyzerAdvancesSpringWithoutNewPCM(t *testing.T) {
	analyzer := NewPCMAnalyzer(time.Hour)
	defer analyzer.Close()
	analyzer.springL = harmonica.NewSpring(harmonica.FPS(60), 9, 1)

	// Seed spring with an initial position and feed a high target.
	target := SpectrumFrame{}
	target.LevelsL[0] = 1
	first := analyzer.springFrame(target).LevelsL[0]
	second := analyzer.springFrame(target).LevelsL[0]
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

func TestPCMAnalyzerStereoSeparatesChannels(t *testing.T) {
	analyzer := NewPCMAnalyzer(100 * time.Millisecond)
	defer analyzer.Close()

	const sampleRate = 44100.0

	// Left: 440 Hz sine, Right: 1000 Hz sine
	samplesL := make([]float32, spectrumFFTSize)
	samplesR := make([]float32, spectrumFFTSize)
	for i := range samplesL {
		samplesL[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / sampleRate))
		samplesR[i] = float32(math.Sin(2 * math.Pi * 1000 * float64(i) / sampleRate))
	}
	analyzer.NewConsumer()(sampleRate, samplesL, samplesR)

	expectedL := int(math.Log(440.0/60) / math.Log(16000.0/60) * SpectrumBandCount)
	expectedR := int(math.Log(1000.0/60) / math.Log(16000.0/60) * SpectrumBandCount)

	deadline := time.After(time.Second)
	for {
		frame := analyzer.Spectrum()

		peakL, peakR := findPeakBand(frame.LevelsL), findPeakBand(frame.LevelsR)
		if frame.LevelsL[peakL] > 0 && frame.LevelsR[peakR] > 0 {
			if abs(peakL-expectedL) > 2 {
				t.Fatalf("L peak band = %d, want near %d", peakL, expectedL)
			}
			if abs(peakR-expectedR) > 2 {
				t.Fatalf("R peak band = %d, want near %d", peakR, expectedR)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("stereo analyzer did not produce separate L/R levels")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func findPeakBand(levels [SpectrumBandCount]float64) int {
	peakBand, peakLevel := 0, 0.0
	for band, level := range levels {
		if level > peakLevel {
			peakBand, peakLevel = band, level
		}
	}
	return peakBand
}
