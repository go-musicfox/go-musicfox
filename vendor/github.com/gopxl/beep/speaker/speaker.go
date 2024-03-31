// Package speaker implements playback of beep.Streamer values through physical speakers.
package speaker

import (
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/pkg/errors"

	"github.com/gopxl/beep"
)

const channelCount = 2
const bitDepthInBytes = 2
const bytesPerSample = bitDepthInBytes * channelCount
const otoFormat = oto.FormatSignedInt16LE

var (
	mu      sync.Mutex
	mixer   beep.Mixer
	context *oto.Context
	player  *oto.Player

	bufferDuration time.Duration
)

// Init initializes audio playback through speaker. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(sampleRate beep.SampleRate, bufferSize int) error {
	if context != nil {
		return errors.New("speaker cannot be initialized more than once")
	}

	mixer = beep.Mixer{}

	// We split the total amount of buffer size between the driver and the player.
	// This seems to be a decent ratio on my machine, but it may have different
	// results on other OS's because of different underlying implementations.
	// Both buffers try to keep themselves filled, so the total buffered
	// number of samples should be some number less than bufferSize.
	driverBufferSize := bufferSize / 2
	playerBufferSize := bufferSize / 2

	var err error
	var readyChan chan struct{}
	context, readyChan, err = oto.NewContext(&oto.NewContextOptions{
		SampleRate:   int(sampleRate),
		ChannelCount: channelCount,
		Format:       otoFormat,
		BufferSize:   sampleRate.D(driverBufferSize),
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}
	<-readyChan

	player = context.NewPlayer(newReaderFromStreamer(&mixer))
	player.SetBufferSize(playerBufferSize * bytesPerSample)
	player.Play()

	bufferDuration = sampleRate.D(bufferSize)

	return nil
}

// Close closes audio playback. However, the underlying driver context keeps existing, because
// closing it isn't supported (https://github.com/hajimehoshi/oto/issues/149). In most cases,
// there is certainly no need to call Close even when the program doesn't play anymore, because
// in properly set systems, the default mixer handles multiple concurrent processes.
func Close() {
	if player != nil {
		player.Close()
		player = nil
		Clear()
	}
}

// Lock locks the speaker. While locked, speaker won't pull new data from the playing Streamers. Lock
// if you want to modify any currently playing Streamers to avoid race conditions.
//
// Always lock speaker for as little time as possible, to avoid playback glitches.
func Lock() {
	mu.Lock()
}

// Unlock unlocks the speaker. Call after modifying any currently playing Streamer.
func Unlock() {
	mu.Unlock()
}

// Play starts playing all provided Streamers through the speaker.
func Play(s ...beep.Streamer) {
	mu.Lock()
	mixer.Add(s...)
	mu.Unlock()
}

// PlayAndWait plays all provided Streamers through the speaker and waits until they have all finished playing.
func PlayAndWait(s ...beep.Streamer) {
	mu.Lock()
	var wg sync.WaitGroup
	wg.Add(len(s))
	for _, e := range s {
		mixer.Add(beep.Seq(e, beep.Callback(func() {
			wg.Done()
		})))
	}
	mu.Unlock()

	// Wait for the streamers to drain.
	wg.Wait()

	// Wait the expected time it takes for the samples to reach the driver.
	time.Sleep(bufferDuration)
}

// Suspend suspends the entire audio play.
//
// This function is intended to save resources when no audio is playing.
// To suspend individual streams, use the beep.Ctrl.
func Suspend() error {
	err := context.Suspend()
	if err != nil {
		return errors.Wrap(err, "failed to suspend the speaker")
	}
	return nil
}

// Resume resumes the entire audio play, which was suspended by Suspend.
func Resume() error {
	err := context.Resume()
	if err != nil {
		return errors.Wrap(err, "failed to resume the speaker")
	}
	return nil
}

// Clear removes all currently playing Streamers from the speaker.
// Previously buffered samples may still be played.
func Clear() {
	mu.Lock()
	mixer.Clear()
	mu.Unlock()
}

// sampleReader is a wrapper for beep.Streamer to implement io.Reader.
type sampleReader struct {
	s   beep.Streamer
	buf [][2]float64
}

func newReaderFromStreamer(s beep.Streamer) *sampleReader {
	return &sampleReader{
		s: s,
	}
}

// Read pulls samples from the streamer and fills buf with the encoded
// samples. Read expects the size of buf be divisible by the length
// of a sample (= channel count * bit depth in bytes).
func (s *sampleReader) Read(buf []byte) (n int, err error) {
	// Read samples from streamer
	if len(buf)%bytesPerSample != 0 {
		return 0, errors.New("requested number of bytes do not align with the samples")
	}
	ns := len(buf) / bytesPerSample
	if len(s.buf) < ns {
		s.buf = make([][2]float64, ns)
	}
	ns, ok := s.stream(s.buf[:ns])
	if !ok {
		if s.s.Err() != nil {
			return 0, errors.Wrap(s.s.Err(), "streamer returned error when requesting samples")
		}
		if ns == 0 {
			return 0, io.EOF
		}
	}

	// Convert samples to bytes
	for i := range s.buf[:ns] {
		for c := range s.buf[i] {
			val := s.buf[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			buf[i*bytesPerSample+c*bitDepthInBytes+0] = low
			buf[i*bytesPerSample+c*bitDepthInBytes+1] = high
		}
	}

	return ns * bytesPerSample, nil
}

// stream pull samples from the streamer while preventing concurrency
// problems by locking the global mixer.
func (s *sampleReader) stream(samples [][2]float64) (n int, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	return s.s.Stream(samples)
}
