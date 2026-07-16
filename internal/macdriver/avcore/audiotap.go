//go:build darwin

package avcore

import (
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ebitengine/purego"
)

const mTAudioProcessingTapCreationFlagPreEffects uint32 = 1

const (
	audioFormatLinearPCM            uint32 = 0x6c70636d // 'lpcm'
	audioFormatFlagIsFloat          uint32 = 1 << 0
	audioFormatFlagIsBigEndian      uint32 = 1 << 1
	audioFormatFlagIsSignedInteger  uint32 = 1 << 2
	audioFormatFlagIsPacked         uint32 = 1 << 3
	audioFormatFlagIsAlignedHigh    uint32 = 1 << 4
	audioFormatFlagIsNonInterleaved uint32 = 1 << 5
)

const (
	audioTapCallbacksSize    = 52
	audioTapClientInfoOffset = 4
	audioTapInitOffset       = 12
	audioTapFinalizeOffset   = 20
	audioTapPrepareOffset    = 28
	audioTapUnprepareOffset  = 36
	audioTapProcessOffset    = 44
)

// AudioTap observes decoded PCM frames. samples is valid only for the duration
// of the callback and must be copied by the receiver if it needs to retain it.
type AudioTap struct {
	id       uint64
	handler  func(sampleRate float64, samples []float32)
	format   AudioStreamBasicDescription
	samples  []float32
	ref      uintptr
	released atomic.Bool
}

// AudioStreamBasicDescription mirrors CoreAudio's structure.
type AudioStreamBasicDescription struct {
	SampleRate       float64
	FormatID         uint32
	FormatFlags      uint32
	BytesPerPacket   uint32
	FramesPerPacket  uint32
	BytesPerFrame    uint32
	ChannelsPerFrame uint32
	BitsPerChannel   uint32
	Reserved         uint32
}

type audioBuffer struct {
	NumberChannels uint32
	DataByteSize   uint32
	Data           unsafe.Pointer
}

type audioBufferList struct {
	NumberBuffers uint32
	_             uint32
	Buffers       [1]audioBuffer
}

var (
	audioTapOnce       sync.Once
	coreFoundationOnce sync.Once
	audioTapRegistry   sync.Map
	audioTapID         atomic.Uint64

	mTAudioProcessingTapCreate         func(allocator unsafe.Pointer, callbacks unsafe.Pointer, flags uint32, tapOut *uintptr) int32
	mTAudioProcessingTapGetStorage     func(tap uintptr) unsafe.Pointer
	mTAudioProcessingTapGetSourceAudio func(tap uintptr, numberFrames int64, bufferList *audioBufferList, flagsOut *uint32, timeRangeOut unsafe.Pointer, numberFramesOut *int64) int32
	cfRelease                          func(unsafe.Pointer)

	audioTapInitCallback      = purego.NewCallback(audioTapInit)
	audioTapFinalizeCallback  = purego.NewCallback(audioTapFinalize)
	audioTapPrepareCallback   = purego.NewCallback(audioTapPrepare)
	audioTapUnprepareCallback = purego.NewCallback(audioTapUnprepare)
	audioTapProcessCallback   = purego.NewCallback(audioTapProcess)
)

func importAudioTapFrameworks() {
	audioTapOnce.Do(func() {
		lib, err := purego.Dlopen("/System/Library/Frameworks/MediaToolbox.framework/MediaToolbox", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
		purego.RegisterLibFunc(&mTAudioProcessingTapCreate, lib, "MTAudioProcessingTapCreate")
		purego.RegisterLibFunc(&mTAudioProcessingTapGetStorage, lib, "MTAudioProcessingTapGetStorage")
		purego.RegisterLibFunc(&mTAudioProcessingTapGetSourceAudio, lib, "MTAudioProcessingTapGetSourceAudio")
	})
	coreFoundationOnce.Do(func() {
		lib, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
		purego.RegisterLibFunc(&cfRelease, lib, "CFRelease")
	})
}

// NewAudioTap creates a pre-effects tap. Its callback copies only decoded PCM
// data into the supplied handler and never changes the audio buffers.
func NewAudioTap(handler func(sampleRate float64, samples []float32)) (*AudioTap, error) {
	if handler == nil {
		return nil, fmt.Errorf("audio tap handler is nil")
	}
	importAudioTapFrameworks()

	id := audioTapID.Add(1)
	tap := &AudioTap{id: id, handler: handler}
	audioTapRegistry.Store(id, tap)

	var callbacks [audioTapCallbacksSize]byte
	binary.LittleEndian.PutUint64(callbacks[audioTapClientInfoOffset:], uint64(uintptr(unsafe.Pointer(uintptr(id)))))
	binary.LittleEndian.PutUint64(callbacks[audioTapInitOffset:], uint64(audioTapInitCallback))
	binary.LittleEndian.PutUint64(callbacks[audioTapFinalizeOffset:], uint64(audioTapFinalizeCallback))
	binary.LittleEndian.PutUint64(callbacks[audioTapPrepareOffset:], uint64(audioTapPrepareCallback))
	binary.LittleEndian.PutUint64(callbacks[audioTapUnprepareOffset:], uint64(audioTapUnprepareCallback))
	binary.LittleEndian.PutUint64(callbacks[audioTapProcessOffset:], uint64(audioTapProcessCallback))

	status := mTAudioProcessingTapCreate(nil, unsafe.Pointer(&callbacks[0]), mTAudioProcessingTapCreationFlagPreEffects, &tap.ref)
	runtime.KeepAlive(&callbacks)
	if status != 0 || tap.ref == 0 {
		audioTapRegistry.Delete(id)
		return nil, fmt.Errorf("create audio tap: OSStatus %d", status)
	}
	return tap, nil
}

// Close releases the caller's creation reference. AVAudioMixInputParameters
// retains a successfully attached tap until its player item is released.
func (t *AudioTap) Close() {
	if t == nil || t.ref == 0 || !t.released.CompareAndSwap(false, true) {
		return
	}
	cfRelease(unsafe.Pointer(t.ref))
}

func audioTapInit(_ uintptr, clientInfo unsafe.Pointer, storageOut *unsafe.Pointer) {
	if storageOut != nil {
		*storageOut = clientInfo
	}
}

func audioTapFinalize(tapRef uintptr) {
	storage := mTAudioProcessingTapGetStorage(tapRef)
	if storage != nil {
		audioTapRegistry.Delete(uint64(uintptr(storage)))
	}
}

func audioTapPrepare(tapRef uintptr, maxFrames int64, format *AudioStreamBasicDescription) {
	tap := audioTapFromRef(tapRef)
	if tap == nil || format == nil || maxFrames <= 0 || maxFrames > int64(^uint(0)>>1) {
		return
	}
	tap.format = *format
	tap.samples = make([]float32, int(maxFrames))
}

func audioTapUnprepare(_ uintptr) {}

func audioTapProcess(tapRef uintptr, numberFrames int64, _ uint32, bufferList *audioBufferList, numberFramesOut *int64, flagsOut *uint32) {
	var (
		sourceFrames int64
		sourceFlags  uint32
	)
	status := mTAudioProcessingTapGetSourceAudio(tapRef, numberFrames, bufferList, &sourceFlags, nil, &sourceFrames)
	if status == 0 && sourceFrames > 0 {
		if tap := audioTapFromRef(tapRef); tap != nil {
			tap.observe(bufferList, int(sourceFrames))
		}
	}
	if numberFramesOut != nil {
		*numberFramesOut = sourceFrames
	}
	if flagsOut != nil {
		*flagsOut = sourceFlags
	}
}

func audioTapFromRef(tapRef uintptr) *AudioTap {
	storage := mTAudioProcessingTapGetStorage(tapRef)
	if storage == nil {
		return nil
	}
	value, ok := audioTapRegistry.Load(uint64(uintptr(storage)))
	if !ok {
		return nil
	}
	return value.(*AudioTap)
}

func (t *AudioTap) observe(bufferList *audioBufferList, frames int) {
	if bufferList == nil || frames <= 0 || len(t.samples) == 0 || t.format.FormatID != audioFormatLinearPCM || t.format.FormatFlags&audioFormatFlagIsBigEndian != 0 {
		return
	}

	channels := int(t.format.ChannelsPerFrame)
	sampleBytes := int(t.format.BitsPerChannel / 8)
	if channels <= 0 || sampleBytes <= 0 || t.format.BytesPerFrame == 0 || !supportedPCMFormat(t.format) {
		return
	}
	if frames > len(t.samples) {
		frames = len(t.samples)
	}

	if t.format.FormatFlags&audioFormatFlagIsNonInterleaved != 0 {
		if int(bufferList.NumberBuffers) < channels {
			return
		}
		for frame := 0; frame < frames; frame++ {
			var sum float64
			for channel := 0; channel < channels; channel++ {
				buffer := audioBufferAt(bufferList, channel)
				if buffer == nil || buffer.Data == nil || int(buffer.DataByteSize) < (frame+1)*int(t.format.BytesPerFrame) {
					return
				}
				sum += pcmSample(unsafe.Add(buffer.Data, frame*int(t.format.BytesPerFrame)), t.format)
			}
			t.samples[frame] = float32(sum / float64(channels))
		}
	} else {
		buffer := audioBufferAt(bufferList, 0)
		if buffer == nil || buffer.Data == nil || int(buffer.DataByteSize) < frames*int(t.format.BytesPerFrame) {
			return
		}
		for frame := 0; frame < frames; frame++ {
			frameStart := unsafe.Add(buffer.Data, frame*int(t.format.BytesPerFrame))
			var sum float64
			for channel := 0; channel < channels; channel++ {
				sum += pcmSample(unsafe.Add(frameStart, channel*sampleBytes), t.format)
			}
			t.samples[frame] = float32(sum / float64(channels))
		}
	}
	t.handler(t.format.SampleRate, t.samples[:frames])
}

func supportedPCMFormat(format AudioStreamBasicDescription) bool {
	if format.FormatFlags&audioFormatFlagIsFloat != 0 {
		return format.BitsPerChannel == 32 || format.BitsPerChannel == 64
	}
	return format.FormatFlags&audioFormatFlagIsSignedInteger != 0 && (format.BitsPerChannel == 16 || format.BitsPerChannel == 32)
}

func pcmSample(data unsafe.Pointer, format AudioStreamBasicDescription) float64 {
	if format.FormatFlags&audioFormatFlagIsFloat != 0 {
		if format.BitsPerChannel == 32 {
			return float64(*(*float32)(data))
		}
		return *(*float64)(data)
	}
	if format.BitsPerChannel == 16 {
		return float64(*(*int16)(data)) / math.MaxInt16
	}
	return float64(*(*int32)(data)) / math.MaxInt32
}

func audioBufferAt(list *audioBufferList, index int) *audioBuffer {
	if index < 0 || index >= int(list.NumberBuffers) {
		return nil
	}
	return (*audioBuffer)(unsafe.Add(unsafe.Pointer(&list.Buffers[0]), uintptr(index)*unsafe.Sizeof(audioBuffer{})))
}
