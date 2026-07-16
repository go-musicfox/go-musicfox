package player

import (
	"github.com/charmbracelet/harmonica"

	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// SpectrumBandCount is the fixed analysis resolution exposed to renderers.
	SpectrumBandCount = 64
	spectrumFFTSize   = 1024
	spectrumSlots     = 3
)

const (
	spectrumSlotFree uint32 = iota
	spectrumSlotWriting
	spectrumSlotReady
	spectrumSlotReading
)

// SpectrumFrame contains normalized terminal bar levels in the range [0, 1].
// Renderers quantize only at their final terminal-cell boundary.
type SpectrumFrame struct {
	Levels [SpectrumBandCount]float64
}

// SpectrumProvider is implemented by playback engines that expose live PCM analysis.
type SpectrumProvider interface {
	Spectrum() SpectrumFrame
}

type spectrumSlot struct {
	state      atomic.Uint32
	generation uint64
	sequence   uint64
	sampleRate float64
	count      int
	samples    [spectrumFFTSize]float32
}

// PCMAnalyzer accepts PCM from a real-time callback and analyzes it away from
// the audio thread. Full input slots are dropped instead of blocking playback.
type PCMAnalyzer struct {
	slots      [spectrumSlots]spectrumSlot
	sequence   atomic.Uint64
	generation atomic.Uint64

	closeOnce sync.Once
	stop      chan struct{}
	done      chan struct{}

	frameMu sync.RWMutex
	frame   SpectrumFrame

	window     [spectrumFFTSize]float64
	fft        [spectrumFFTSize]complex128
	interval   time.Duration
	spring     harmonica.Spring
	positions  [SpectrumBandCount]float64
	velocities [SpectrumBandCount]float64
	processed  uint64
	target     SpectrumFrame
}

func NewPCMAnalyzer(interval time.Duration) *PCMAnalyzer {
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	a := &PCMAnalyzer{
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		interval: interval,
		spring:   harmonica.NewSpring(interval.Seconds(), 9, 1),
	}
	for i := range a.window {
		a.window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(spectrumFFTSize-1)))
	}
	go a.run()
	return a
}

// NewConsumer resets the visible frame and returns a callback for one audio
// source. Frames from a replaced player item are discarded by generation.
func (a *PCMAnalyzer) NewConsumer() func(sampleRate float64, samples []float32) {
	generation := a.generation.Add(1)
	a.frameMu.Lock()
	a.frame = SpectrumFrame{}
	a.frameMu.Unlock()

	return func(sampleRate float64, samples []float32) {
		a.consume(generation, sampleRate, samples)
	}
}

func (a *PCMAnalyzer) consume(generation uint64, sampleRate float64, samples []float32) {
	if generation != a.generation.Load() || sampleRate <= 0 || len(samples) == 0 {
		return
	}
	for index := range a.slots {
		slot := &a.slots[index]
		if !slot.state.CompareAndSwap(spectrumSlotFree, spectrumSlotWriting) {
			continue
		}
		count := min(len(samples), spectrumFFTSize)
		copy(slot.samples[:count], samples[len(samples)-count:])
		slot.generation = generation
		slot.sequence = a.sequence.Add(1)
		slot.sampleRate = sampleRate
		slot.count = count
		slot.state.Store(spectrumSlotReady)
		return
	}
}

func (a *PCMAnalyzer) run() {
	ticker := time.NewTicker(a.interval)
	defer func() {
		ticker.Stop()
		close(a.done)
	}()
	for {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
			a.analyzeLatest()
		}
	}
}

func (a *PCMAnalyzer) analyzeLatest() {
	generation := a.generation.Load()
	if generation != a.processed {
		a.positions = [SpectrumBandCount]float64{}
		a.velocities = [SpectrumBandCount]float64{}
		a.target = SpectrumFrame{}
		a.processed = generation
	}

	var latest *spectrumSlot
	for index := range a.slots {
		slot := &a.slots[index]
		if !slot.state.CompareAndSwap(spectrumSlotReady, spectrumSlotReading) {
			continue
		}
		if slot.generation != generation {
			slot.state.Store(spectrumSlotFree)
			continue
		}
		if latest == nil || slot.sequence > latest.sequence {
			if latest != nil {
				latest.state.Store(spectrumSlotFree)
			}
			latest = slot
			continue
		}
		slot.state.Store(spectrumSlotFree)
	}
	if latest != nil {
		defer latest.state.Store(spectrumSlotFree)
		a.transform(latest.samples[:latest.count])
		if latest.generation != a.generation.Load() {
			return
		}
		a.target = a.targetLevels(latest.sampleRate)
	}

	a.publish(a.springFrame(a.target))
}

func (a *PCMAnalyzer) transform(samples []float32) {
	padding := spectrumFFTSize - len(samples)
	for index := 0; index < padding; index++ {
		a.fft[index] = 0
	}
	for index, sample := range samples {
		a.fft[padding+index] = complex(float64(sample)*a.window[padding+index], 0)
	}

	for index := 1; index < spectrumFFTSize; index++ {
		reversed := reverseSpectrumBits(index)
		if index < reversed {
			a.fft[index], a.fft[reversed] = a.fft[reversed], a.fft[index]
		}
	}
	for size := 2; size <= spectrumFFTSize; size <<= 1 {
		angle := -2 * math.Pi / float64(size)
		step := complex(math.Cos(angle), math.Sin(angle))
		half := size / 2
		for offset := 0; offset < spectrumFFTSize; offset += size {
			weight := complex(1, 0)
			for index := 0; index < half; index++ {
				even := a.fft[offset+index]
				odd := weight * a.fft[offset+index+half]
				a.fft[offset+index] = even + odd
				a.fft[offset+index+half] = even - odd
				weight *= step
			}
		}
	}
}

func reverseSpectrumBits(value int) int {
	reversed := 0
	for bit := 0; bit < 10; bit++ {
		reversed = reversed<<1 | value&1
		value >>= 1
	}
	return reversed
}

func (a *PCMAnalyzer) targetLevels(sampleRate float64) SpectrumFrame {
	var target SpectrumFrame
	lowFrequency := 60.0
	highFrequency := min(16000.0, sampleRate/2)
	if highFrequency <= lowFrequency {
		return target
	}
	ratio := highFrequency / lowFrequency
	for band := range target.Levels {
		startFrequency := lowFrequency * math.Pow(ratio, float64(band)/SpectrumBandCount)
		endFrequency := lowFrequency * math.Pow(ratio, float64(band+1)/SpectrumBandCount)
		startBin := max(1, int(math.Floor(startFrequency*float64(spectrumFFTSize)/sampleRate)))
		endBin := min(spectrumFFTSize/2, int(math.Ceil(endFrequency*float64(spectrumFFTSize)/sampleRate)))
		if endBin <= startBin {
			endBin = min(spectrumFFTSize/2, startBin+1)
		}

		magnitude := 0.0
		for bin := startBin; bin < endBin; bin++ {
			value := a.fft[bin]
			magnitude = max(magnitude, math.Hypot(real(value), imag(value)))
		}
		db := 20 * math.Log10(magnitude*4/float64(spectrumFFTSize)+1e-9)
		target.Levels[band] = min(1.0, max(0.0, (db+72)/72))
	}
	return target
}

func (a *PCMAnalyzer) springFrame(target SpectrumFrame) SpectrumFrame {
	var frame SpectrumFrame
	for band, targetLevel := range target.Levels {
		position, velocity := a.spring.Update(a.positions[band], a.velocities[band], targetLevel)
		a.positions[band] = min(1.0, max(0.0, position))
		a.velocities[band] = velocity
		frame.Levels[band] = a.positions[band]
	}
	return frame
}

func (a *PCMAnalyzer) publish(frame SpectrumFrame) {
	a.frameMu.Lock()
	a.frame = frame
	a.frameMu.Unlock()
}

func (a *PCMAnalyzer) Spectrum() SpectrumFrame {
	a.frameMu.RLock()
	defer a.frameMu.RUnlock()
	return a.frame
}

func (a *PCMAnalyzer) Close() {
	a.closeOnce.Do(func() {
		close(a.stop)
		<-a.done
	})
}
