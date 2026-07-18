//go:build darwin

package avcore

import (
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func TestAudioTapLayout(t *testing.T) {
	if got, want := unsafe.Sizeof(AudioStreamBasicDescription{}), uintptr(40); got != want {
		t.Fatalf("AudioStreamBasicDescription size = %d, want %d", got, want)
	}
	if got, want := unsafe.Sizeof(audioBuffer{}), uintptr(16); got != want {
		t.Fatalf("AudioBuffer size = %d, want %d", got, want)
	}
	if got, want := unsafe.Sizeof(audioBufferList{}), uintptr(24); got != want {
		t.Fatalf("AudioBufferList size = %d, want %d", got, want)
	}
}

func TestAudioTapObservesInterleavedFloat32(t *testing.T) {
	input := []float32{1, 0, 0.5, -0.25} // L,R,L,R
	var (
		observedL []float32
		observedR []float32
	)
	tap := AudioTap{
		format: AudioStreamBasicDescription{
			SampleRate:       44100,
			FormatID:         audioFormatLinearPCM,
			FormatFlags:      audioFormatFlagIsFloat | audioFormatFlagIsPacked,
			BytesPerFrame:    8,
			ChannelsPerFrame: 2,
			BitsPerChannel:   32,
		},
		samples:  make([]float32, 2),
		samplesR: make([]float32, 2),
		handler: func(_ float64, samplesL, samplesR []float32) {
			observedL = append(observedL, samplesL...)
			observedR = append(observedR, samplesR...)
		},
	}
	buffers := audioBufferList{
		NumberBuffers: 1,
		Buffers: [1]audioBuffer{{
			NumberChannels: 2,
			DataByteSize:   uint32(len(input) * 4),
			Data:           unsafe.Pointer(&input[0]),
		}},
	}

	tap.observe(&buffers, 2)

	// Left channel: [1, 0.5]
	if len(observedL) != 2 {
		t.Fatalf("observedL frames = %d, want 2", len(observedL))
	}
	if observedL[0] != 1 || observedL[1] != 0.5 {
		t.Fatalf("observedL = %v, want [1 0.5]", observedL)
	}
	// Right channel: [0, -0.25]
	if len(observedR) != 2 {
		t.Fatalf("observedR frames = %d, want 2", len(observedR))
	}
	if observedR[0] != 0 || observedR[1] != -0.25 {
		t.Fatalf("observedR = %v, want [0 -0.25]", observedR)
	}
}

func TestAudioTapCapturesPlaybackPCM(t *testing.T) {
	var frames atomic.Int64
	tap, err := NewAudioTap(func(_ float64, samplesL, samplesR []float32) {
		frames.Add(int64(len(samplesL)))
		_ = samplesR
	})
	if err != nil {
		t.Fatal(err)
	}

	_, path, _, _ := runtime.Caller(0)
	file := core.String("/" + filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path)))), "testdata", "a.mp3"))
	defer file.Release()
	url := core.NSURL_fileURLWithPath(file)
	defer url.Release()
	item := AVPlayerItem_playerItemWithURL(url)
	defer item.Release()
	if !item.AttachAudioTap(tap) {
		tap.Close()
		t.Fatal("attach audio tap")
	}

	player := AVPlayer_alloc().Init()
	defer player.Release()
	player.ReplaceCurrentItemWithPlayerItem(item)
	player.Play()
	defer player.Pause()

	deadline := time.After(5 * time.Second)
	for frames.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("audio tap did not receive PCM frames")
		case <-time.After(25 * time.Millisecond):
		}
	}
}
