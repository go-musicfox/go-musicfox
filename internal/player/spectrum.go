package player

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/harmonica"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

const (
	SpectrumBandCount   = 64
	spectrumFFTSize     = 1024
	spectrumSlots       = 3
	rawSampleBufferSize = 4096
)

const (
	spectrumSlotFree uint32 = iota
	spectrumSlotWriting
	spectrumSlotReady
	spectrumSlotReading
)

var fftTwiddle [spectrumFFTSize / 2][2]float64 // [real, imag]
var fftTwiddleOnce sync.Once

func initFFTTwiddle() {
	fftTwiddleOnce.Do(func() {
		for k := 0; k < spectrumFFTSize/2; k++ {
			sin, cos := math.Sincos(-2.0 * math.Pi * float64(k) / float64(spectrumFFTSize))
			fftTwiddle[k][0] = cos // real part
			fftTwiddle[k][1] = sin // imag part
		}
	})
}

// SpectrumFrame holds per-band levels.
// Levels  = combined (average of L+R, used by bar/mirror_bar).
// LevelsL = left channel (used by line/dot for dual display).
// LevelsR = right channel.
// PhasesL / PhasesR = phase angle (radians) of dominant FFT bin per band.
type SpectrumFrame struct {
	Levels  [SpectrumBandCount]float64
	LevelsL [SpectrumBandCount]float64
	LevelsR [SpectrumBandCount]float64
	PhasesL [SpectrumBandCount]float64
	PhasesR [SpectrumBandCount]float64
}

// RawSampleFrame holds raw PCM time-domain samples for oscilloscope/vectorscope.
// SamplesL and SamplesR are freshly-allocated snapshots of the ring buffer;
// Count is the number of valid samples in each slice.
type RawSampleFrame struct {
	SamplesL   []float64
	SamplesR   []float64
	Count      int
	SampleRate float64
}

type SpectrumProvider interface {
	Spectrum() SpectrumFrame
	RawSamples() RawSampleFrame
}

type spectrumSlot struct {
	state      atomic.Uint32
	generation uint64
	sequence   uint64
	sampleRate float64
	count      int
	hasStereo  bool
	samples    [spectrumFFTSize]float32
	samplesR   [spectrumFFTSize]float32
}

// PCMAnalyzer accepts stereo PCM callbacks and analyzes each channel independently.
type PCMAnalyzer struct {
	slots      [spectrumSlots]spectrumSlot
	sequence   atomic.Uint64
	generation atomic.Uint64

	closeOnce sync.Once
	stop      chan struct{}
	done      chan struct{}

	frameMu sync.RWMutex
	frame   SpectrumFrame

	// Raw PCM ring buffer for oscilloscope/vectorscope.
	rawRingL      [rawSampleBufferSize]float64
	rawRingR      [rawSampleBufferSize]float64
	rawRingPos    atomic.Uint32
	rawRingCount  atomic.Uint32
	rawSampleRate atomic.Uint64 // math.Float64bits(sampleRate)

	rawMu   sync.RWMutex
	rawSnap RawSampleFrame // pre-allocated snapshot buffers

	window       [spectrumFFTSize]float64
	fft          [spectrumFFTSize]complex128
	fftR         [spectrumFFTSize]complex128
	interval     time.Duration

	springL      harmonica.Spring
	positionsL   [SpectrumBandCount]float64
	velocitiesL  [SpectrumBandCount]float64
	springR      harmonica.Spring
	positionsR   [SpectrumBandCount]float64
	velocitiesR  [SpectrumBandCount]float64

	// FFT frame averaging (scope-tui spectroscope average feature).
	avgLevelsL [SpectrumBandCount]float64 // previous averaged L levels (EMA)
	avgLevelsR [SpectrumBandCount]float64 // previous averaged R levels (EMA)

	target    SpectrumFrame // last known target, persists across dry ticks
	processed uint64
}

func NewPCMAnalyzer(interval time.Duration) *PCMAnalyzer {
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	a := &PCMAnalyzer{
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		interval: interval,
		springL:  harmonica.NewSpring(interval.Seconds(), 9, 1),
		springR:  harmonica.NewSpring(interval.Seconds(), 9, 1),
	}
	a.rawSnap.SamplesL = make([]float64, rawSampleBufferSize)
	a.rawSnap.SamplesR = make([]float64, rawSampleBufferSize)
	for i := range a.window {
		a.window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(spectrumFFTSize-1)))
	}
	go a.run()
	return a
}

func (a *PCMAnalyzer) NewConsumer() func(sampleRate float64, samplesL, samplesR []float32) {
	generation := a.generation.Add(1)
	a.frameMu.Lock()
	a.frame = SpectrumFrame{}
	a.target = SpectrumFrame{}
	a.frameMu.Unlock()

	// Clear raw ring buffer for new audio source.
	a.rawRingPos.Store(0)
	a.rawRingCount.Store(0)
	a.rawSampleRate.Store(0)
	for i := range a.rawRingL {
		a.rawRingL[i] = 0
		a.rawRingR[i] = 0
	}
	// Reset FFT averaging state.
	a.avgLevelsL = [SpectrumBandCount]float64{}
	a.avgLevelsR = [SpectrumBandCount]float64{}

	return func(sampleRate float64, samplesL, samplesR []float32) {
		a.consume(generation, sampleRate, samplesL, samplesR)
	}
}

func (a *PCMAnalyzer) consume(generation uint64, sampleRate float64, samplesL, samplesR []float32) {
	if generation != a.generation.Load() || sampleRate <= 0 || len(samplesL) == 0 {
		return
	}

	// Copy to raw PCM ring buffer for oscilloscope/vectorscope.
	a.rawSampleRate.Store(math.Float64bits(sampleRate))
	count := min(len(samplesL), rawSampleBufferSize)
	pos := int(a.rawRingPos.Load())
	for i := 0; i < count; i++ {
		idx := (pos + i) % rawSampleBufferSize
		a.rawRingL[idx] = float64(samplesL[len(samplesL)-count+i])
		if len(samplesR) >= count {
			a.rawRingR[idx] = float64(samplesR[len(samplesR)-count+i])
		}
	}
	newPos := uint32((pos + count) % rawSampleBufferSize)
	a.rawRingPos.Store(newPos)
	if c := a.rawRingCount.Load(); c < rawSampleBufferSize {
		a.rawRingCount.Store(min(c+uint32(count), rawSampleBufferSize))
	}

	for i := range a.slots {
		slot := &a.slots[i]
		if !slot.state.CompareAndSwap(spectrumSlotFree, spectrumSlotWriting) {
			continue
		}
		slotCount := min(len(samplesL), spectrumFFTSize)
		copy(slot.samples[:slotCount], samplesL[len(samplesL)-slotCount:])
		slot.hasStereo = len(samplesR) >= slotCount
		if slot.hasStereo {
			copy(slot.samplesR[:slotCount], samplesR[len(samplesR)-slotCount:])
		}
		slot.generation = generation
		slot.sequence = a.sequence.Add(1)
		slot.sampleRate = sampleRate
		slot.count = slotCount
		slot.state.Store(spectrumSlotReady)
		return
	}
}

func (a *PCMAnalyzer) run() {
	ticker := time.NewTicker(a.interval)
	defer func() { ticker.Stop(); close(a.done) }()
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
	gen := a.generation.Load()
	if gen != a.processed {
		a.positionsL, a.velocitiesL = [SpectrumBandCount]float64{}, [SpectrumBandCount]float64{}
		a.positionsR, a.velocitiesR = [SpectrumBandCount]float64{}, [SpectrumBandCount]float64{}
		a.avgLevelsL = [SpectrumBandCount]float64{}
		a.avgLevelsR = [SpectrumBandCount]float64{}
		a.target = SpectrumFrame{}
		a.processed = gen
	}

	var latest *spectrumSlot
	for i := range a.slots {
		slot := &a.slots[i]
		if !slot.state.CompareAndSwap(spectrumSlotReady, spectrumSlotReading) {
			continue
		}
		if slot.generation != gen {
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
	if latest == nil {
		frame := a.springFrame(a.target)

		// Apply cava-inspired smoothing after springFrame
		if configs.AppConfig != nil {
			cfg := configs.AppConfig.Main.Visualizer
			if cfg.Monstercat > 0 {
				a.monstercatFilter(&frame.Levels)
			} else if cfg.Waves {
				a.wavesFilter(&frame.Levels)
			}
		}

		a.publish(frame)
		return
	}
	defer latest.state.Store(spectrumSlotFree)

	// Left channel FFT
	a.transform(a.fft[:], latest.samples[:latest.count])
	if latest.generation != a.generation.Load() {
		return
	}
	logScale := true
	if configs.AppConfig != nil {
		logScale = configs.AppConfig.Main.Visualizer.SpectrumLogScale
	}
	rawL, rawPhasesL := a.channelLevelsAndPhases(a.fft[:], latest.sampleRate, logScale)
	a.target.LevelsL = a.applyAvg(rawL, a.avgLevelsL[:])
	a.target.PhasesL = rawPhasesL
	a.target.Levels = a.target.LevelsL

	// Right channel FFT
	if latest.hasStereo {
		if configs.AppConfig != nil && configs.AppConfig.Main.Visualizer.IsMono() {
			// Mono mode: skip right FFT, use left channel for all
			a.target.LevelsR = a.target.LevelsL
			a.target.PhasesR = a.target.PhasesL
		} else {
			a.transform(a.fftR[:], latest.samplesR[:latest.count])
			if latest.generation != a.generation.Load() {
				return
			}
			rawR, rawPhasesR := a.channelLevelsAndPhases(a.fftR[:], latest.sampleRate, logScale)
			a.target.LevelsR = a.applyAvg(rawR, a.avgLevelsR[:])
			a.target.PhasesR = rawPhasesR
			for band := range a.target.Levels {
				a.target.Levels[band] = (a.target.LevelsL[band] + a.target.LevelsR[band]) / 2
			}
		}
	}

	// Apply overshoot to target levels before spring interpolation
	if configs.AppConfig != nil {
		overshoot := configs.AppConfig.Main.Visualizer.EffectiveOvershoot()
		if overshoot > 0 {
			for band := range a.target.LevelsL {
				a.target.LevelsL[band] = min(1.0, a.target.LevelsL[band]*(1.0+overshoot))
			}
			for band := range a.target.LevelsR {
				a.target.LevelsR[band] = min(1.0, a.target.LevelsR[band]*(1.0+overshoot))
			}
			for band := range a.target.Levels {
				a.target.Levels[band] = min(1.0, a.target.Levels[band]*(1.0+overshoot))
			}
		}
	}

	frame := a.springFrame(a.target)

	// Apply cava-inspired smoothing after springFrame
	if configs.AppConfig != nil {
		cfg := configs.AppConfig.Main.Visualizer
		if cfg.Monstercat > 0 {
			a.monstercatFilter(&frame.Levels)
		} else if cfg.Waves {
			a.wavesFilter(&frame.Levels)
		}
	}

	a.publish(frame)
}

// --- FFT ---

func (a *PCMAnalyzer) transform(fftBuf []complex128, samples []float32) {
	initFFTTwiddle()
	padding := spectrumFFTSize - len(samples)
	for i := 0; i < padding; i++ {
		fftBuf[i] = 0
	}
	for i, s := range samples {
		fftBuf[padding+i] = complex(float64(s)*a.window[padding+i], 0)
	}
	for i := 1; i < spectrumFFTSize; i++ {
		rev := reverseSpectrumBits(i)
		if i < rev {
			fftBuf[i], fftBuf[rev] = fftBuf[rev], fftBuf[i]
		}
	}
	for size := 2; size <= spectrumFFTSize; size <<= 1 {
		half := size / 2
		stride := spectrumFFTSize / size
		for offset := 0; offset < spectrumFFTSize; offset += size {
			for i := 0; i < half; i++ {
				tw := fftTwiddle[i*stride]
				w := complex(tw[0], tw[1])
				even := fftBuf[offset+i]
				odd := w * fftBuf[offset+i+half]
				fftBuf[offset+i] = even + odd
				fftBuf[offset+i+half] = even - odd
			}
		}
	}
}

func reverseSpectrumBits(v int) int {
	rev := 0
	for bit := 0; bit < 10; bit++ {
		rev = rev<<1 | v&1
		v >>= 1
	}
	return rev
}

// --- Level extraction ---

// channelLevelsAndPhases extracts per-band magnitude levels and phase angles from FFT data.
// When useLogScale is true, magnitudes are converted to dB (log scale);
// when false, linear amplitude is used.
func (a *PCMAnalyzer) channelLevelsAndPhases(fftBuf []complex128, sampleRate float64, useLogScale bool) (levels, phases [SpectrumBandCount]float64) {
	lo := 60.0
	hi := min(16000.0, sampleRate/2)
	if hi <= lo {
		return levels, phases
	}
	ratio := hi / lo
	fftScale := 4.0 / float64(spectrumFFTSize)
	for band := range levels {
		startFreq := lo * math.Pow(ratio, float64(band)/SpectrumBandCount)
		endFreq := lo * math.Pow(ratio, float64(band+1)/SpectrumBandCount)
		startBin := max(1, int(math.Floor(startFreq*float64(spectrumFFTSize)/sampleRate)))
		endBin := min(spectrumFFTSize/2, int(math.Ceil(endFreq*float64(spectrumFFTSize)/sampleRate)))
		if endBin <= startBin {
			endBin = min(spectrumFFTSize/2, startBin+1)
		}
		mag := 0.0
		bestPhase := 0.0
		for bin := startBin; bin < endBin; bin++ {
			v := fftBuf[bin]
			m := math.Hypot(real(v), imag(v))
			if m > mag {
				mag = m
				bestPhase = math.Atan2(imag(v), real(v))
			}
		}
		if useLogScale {
			db := 20 * math.Log10(mag*fftScale+1e-9)
			levels[band] = clamp01((db + 72) / 72)
		} else {
			levels[band] = clamp01(mag * fftScale)
		}
		phases[band] = bestPhase
	}
	return levels, phases
}

// --- Spring animation ---

// applyAvg applies exponential moving average smoothing to a level array.
// The prev slice is updated in-place. Returns the averaged levels.
func (a *PCMAnalyzer) applyAvg(raw [SpectrumBandCount]float64, prev []float64) [SpectrumBandCount]float64 {
	alpha := a.computeAvgAlpha()
	if alpha >= 1.0 {
		return raw
	}
	var out [SpectrumBandCount]float64
	for band := range raw {
		prev[band] = raw[band]*alpha + prev[band]*(1-alpha)
		out[band] = prev[band]
	}
	return out
}

func (a *PCMAnalyzer) computeAvgAlpha() float64 {
	if configs.AppConfig == nil {
		return 1.0
	}
	avg := configs.AppConfig.Main.Visualizer.SpectrumAverage
	if avg <= 1 {
		return 1.0
	}
	return 1.0 / float64(avg)
}

func (a *PCMAnalyzer) springFrame(target SpectrumFrame) SpectrumFrame {
	var frame SpectrumFrame

	if configs.AppConfig != nil && configs.AppConfig.Main.Visualizer.IsMono() {
		// Mono mode: only animate left channel, use it for all
		for band, t := range target.LevelsL {
			pos, vel := a.springL.Update(a.positionsL[band], a.velocitiesL[band], t)
			a.positionsL[band] = clamp01(pos)
			a.velocitiesL[band] = vel
			frame.LevelsL[band] = a.positionsL[band]
		}
		frame.LevelsR = frame.LevelsL
		frame.Levels = frame.LevelsL
		frame.PhasesL = target.PhasesL
		frame.PhasesR = target.PhasesL
		return frame
	}

	// Left channel spring
	for band, t := range target.LevelsL {
		pos, vel := a.springL.Update(a.positionsL[band], a.velocitiesL[band], t)
		a.positionsL[band] = clamp01(pos)
		a.velocitiesL[band] = vel
		frame.LevelsL[band] = a.positionsL[band]
	}

	// Right channel spring
	for band, t := range target.LevelsR {
		pos, vel := a.springR.Update(a.positionsR[band], a.velocitiesR[band], t)
		a.positionsR[band] = clamp01(pos)
		a.velocitiesR[band] = vel
		frame.LevelsR[band] = a.positionsR[band]
	}

	// Combined: average of L+R (scales correctly for both mono and stereo)
	hasR := false
	for _, v := range frame.LevelsR {
		if v > 1e-9 {
			hasR = true
			break
		}
	}
	for band := range frame.Levels {
		if hasR {
			frame.Levels[band] = (frame.LevelsL[band] + frame.LevelsR[band]) / 2
		} else {
			frame.Levels[band] = frame.LevelsL[band]
		}
	}

	// Pass phases through unchanged (spring doesn't apply to phase).
	frame.PhasesL = target.PhasesL
	frame.PhasesR = target.PhasesR

	return frame
}

func clamp01(v float64) float64 {
	return min(1.0, max(0.0, v))
}

// monstercatFilter applies cava-style monstercat smoothing to spectrum levels.
// Lower monstercat values = more spread (more fusion between bars).
// The typical cava formula uses monstercat * 1.5 as the base.
// Only operates on the combined Levels array.
func (a *PCMAnalyzer) monstercatFilter(levels *[SpectrumBandCount]float64) {
	if configs.AppConfig == nil {
		return
	}
	monstercat := configs.AppConfig.Main.Visualizer.Monstercat
	if monstercat <= 0 {
		return
	}
	base := monstercat * 1.5
	for i := range levels {
		if levels[i] == 0 {
			continue
		}
		for d := 1; d < SpectrumBandCount; d++ {
			spread := levels[i] / math.Pow(base, float64(d))
			if spread < 0.01 {
				break
			}
			if i-d >= 0 {
				levels[i-d] = max(levels[i-d], spread)
			}
			if i+d < SpectrumBandCount {
				levels[i+d] = max(levels[i+d], spread)
			}
		}
	}
}

// wavesFilter applies cava-style waves (quadratic decay) smoothing to spectrum levels.
// Mutually exclusive with monstercatFilter.
// Only operates on the combined Levels array.
func (a *PCMAnalyzer) wavesFilter(levels *[SpectrumBandCount]float64) {
	const heightNormalizer = 0.02
	for i := range levels {
		if levels[i] == 0 {
			continue
		}
		for d := 1; d < SpectrumBandCount; d++ {
			spread := levels[i] - heightNormalizer*float64(d*d)
			if spread <= 0 {
				break
			}
			if i-d >= 0 {
				levels[i-d] = max(levels[i-d], spread)
			}
			if i+d < SpectrumBandCount {
				levels[i+d] = max(levels[i+d], spread)
			}
		}
	}
}

// --- Publish / Read ---

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

// RawSamples returns a linearized snapshot of the raw PCM ring buffer
// for oscilloscope and vectorscope rendering.
func (a *PCMAnalyzer) RawSamples() RawSampleFrame {
	pos := int(a.rawRingPos.Load())
	count := int(a.rawRingCount.Load())
	sr := math.Float64frombits(a.rawSampleRate.Load())

	a.rawMu.Lock()
	defer a.rawMu.Unlock()

	snap := &a.rawSnap
	// Linearize the circular buffer: oldest = count samples before pos.
	start := (pos - count + rawSampleBufferSize) % rawSampleBufferSize
	for i := 0; i < count; i++ {
		idx := (start + i) % rawSampleBufferSize
		snap.SamplesL[i] = a.rawRingL[idx]
		snap.SamplesR[i] = a.rawRingR[idx]
	}
	snap.Count = count
	snap.SampleRate = sr
	return *snap
}

func (a *PCMAnalyzer) Close() {
	a.closeOnce.Do(func() {
		close(a.stop)
		<-a.done
	})
}
