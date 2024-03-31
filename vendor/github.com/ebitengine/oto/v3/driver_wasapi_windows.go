// Copyright 2022 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oto

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/ebitengine/oto/v3/internal/mux"
)

type comThread struct {
	funcCh chan func()
}

func newCOMThread() (*comThread, error) {
	funcCh := make(chan func())
	errCh := make(chan error)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// S_FALSE is returned when CoInitializeEx is nested. This is a successful case.
		if err := windows.CoInitializeEx(0, windows.COINIT_MULTITHREADED); err != nil && !errors.Is(err, syscall.Errno(windows.S_FALSE)) {
			errCh <- err
		}
		// CoUninitialize should be called even when CoInitializeEx returns S_FALSE.
		defer windows.CoUninitialize()

		close(errCh)

		for f := range funcCh {
			f()
		}
	}()

	if err := <-errCh; err != nil {
		return nil, err
	}

	return &comThread{
		funcCh: funcCh,
	}, nil
}

func (c *comThread) Run(f func()) {
	ch := make(chan struct{})
	c.funcCh <- func() {
		f()
		close(ch)
	}
	<-ch
}

type wasapiContext struct {
	sampleRate        int
	channelCount      int
	mux               *mux.Mux
	bufferSizeInBytes int

	comThread     *comThread
	err           atomicError
	suspended     bool
	suspendedCond *sync.Cond

	sampleReadyEvent windows.Handle
	client           *_IAudioClient2
	bufferFrames     uint32
	renderClient     *_IAudioRenderClient
	currentDeviceID  string
	enumerator       *_IMMDeviceEnumerator

	buf []float32

	m sync.Mutex
}

var (
	errDeviceSwitched     = errors.New("oto: device switched")
	errFormatNotSupported = errors.New("oto: the specified format is not supported (there is the closest format instead)")
)

func newWASAPIContext(sampleRate, channelCount int, mux *mux.Mux, bufferSizeInBytes int) (context *wasapiContext, ferr error) {
	t, err := newCOMThread()
	if err != nil {
		return nil, err
	}

	c := &wasapiContext{
		sampleRate:        sampleRate,
		channelCount:      channelCount,
		mux:               mux,
		bufferSizeInBytes: bufferSizeInBytes,
		comThread:         t,
		suspendedCond:     sync.NewCond(&sync.Mutex{}),
	}

	ev, err := windows.CreateEventEx(nil, nil, 0, windows.EVENT_ALL_ACCESS)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ferr != nil {
			windows.CloseHandle(ev)
		}
	}()
	c.sampleReadyEvent = ev

	if err := c.start(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *wasapiContext) isDeviceSwitched() (bool, error) {
	// If the audio is suspended, do nothing.
	if c.isSuspended() {
		return false, nil
	}

	var switched bool
	var cerr error
	c.comThread.Run(func() {
		device, err := c.enumerator.GetDefaultAudioEndPoint(eRender, eConsole)
		if err != nil {
			cerr = err
			return
		}
		defer device.Release()

		id, err := device.GetId()
		if err != nil {
			cerr = err
			return
		}

		if c.currentDeviceID == id {
			return
		}
		switched = true
	})

	return switched, cerr
}

func (c *wasapiContext) start() error {
	var cerr error
	c.comThread.Run(func() {
		if err := c.startOnCOMThread(); err != nil {
			cerr = err
			return
		}
	})
	if cerr != nil {
		return cerr
	}

	go func() {
		if err := c.loop(); err != nil {
			if !errors.Is(err, _AUDCLNT_E_DEVICE_INVALIDATED) && !errors.Is(err, _AUDCLNT_E_RESOURCES_INVALIDATED) && !errors.Is(err, errDeviceSwitched) {
				c.err.TryStore(err)
				return
			}

			if err := c.restart(); err != nil {
				c.err.TryStore(err)
				return
			}
		}
	}()

	return nil
}

func (c *wasapiContext) startOnCOMThread() (ferr error) {
	if c.enumerator == nil {
		e, err := _CoCreateInstance(&uuidMMDeviceEnumerator, nil, uint32(_CLSCTX_ALL), &uuidIMMDeviceEnumerator)
		if err != nil {
			return err
		}
		c.enumerator = (*_IMMDeviceEnumerator)(e)
		defer func() {
			if ferr != nil {
				c.enumerator.Release()
				c.enumerator = nil
			}
		}()
	}

	device, err := c.enumerator.GetDefaultAudioEndPoint(eRender, eConsole)
	if err != nil {
		if errors.Is(err, _E_NOTFOUND) {
			return errDeviceNotFound
		}
		return err
	}
	defer device.Release()

	id, err := device.GetId()
	if err != nil {
		return err
	}
	c.currentDeviceID = id

	if c.client != nil {
		c.client.Release()
		c.client = nil
	}

	client, err := device.Activate(&uuidIAudioClient2, uint32(_CLSCTX_ALL), nil)
	if err != nil {
		return err
	}
	c.client = (*_IAudioClient2)(client)

	if err := c.client.SetClientProperties(&_AudioClientProperties{
		cbSize:     uint32(unsafe.Sizeof(_AudioClientProperties{})),
		bIsOffload: 0,                    // false
		eCategory:  _AudioCategory_Other, // In the example, AudioCategory_ForegroundOnlyMedia was used, but this value is deprecated.
	}); err != nil {
		return err
	}

	// Check the format is supported by WASAPI.
	// Stereo with 48000 [Hz] is likely supported, but mono and/or other sample rates are unlikely supported.
	// Fallback to WinMM in this case anyway.
	const bitsPerSample = 32
	nBlockAlign := c.channelCount * bitsPerSample / 8
	var channelMask uint32
	switch c.channelCount {
	case 1:
		channelMask = _SPEAKER_FRONT_CENTER
	case 2:
		channelMask = _SPEAKER_FRONT_LEFT | _SPEAKER_FRONT_RIGHT
	}
	f := &_WAVEFORMATEXTENSIBLE{
		wFormatTag:      _WAVE_FORMAT_EXTENSIBLE,
		nChannels:       uint16(c.channelCount),
		nSamplesPerSec:  uint32(c.sampleRate),
		nAvgBytesPerSec: uint32(c.sampleRate * nBlockAlign),
		nBlockAlign:     uint16(nBlockAlign),
		wBitsPerSample:  bitsPerSample,
		cbSize:          0x16,
		Samples:         bitsPerSample,
		dwChannelMask:   channelMask,
		SubFormat:       _KSDATAFORMAT_SUBTYPE_IEEE_FLOAT,
	}

	var bufferSizeIn100ns _REFERENCE_TIME
	if c.bufferSizeInBytes != 0 {
		bufferSizeInFrames := int64(c.bufferSizeInBytes) / int64(nBlockAlign)
		bufferSizeIn100ns = _REFERENCE_TIME(1e7 * bufferSizeInFrames / int64(c.sampleRate))
	} else {
		// The default buffer size can be too small and might cause glitch noises.
		// Specify 50[ms] as the buffer size.
		bufferSizeIn100ns = _REFERENCE_TIME(50 * time.Millisecond / 100)
	}

	// Even if the sample rate and/or the number of channels are not supported by the audio driver,
	// AUDCLNT_STREAMFLAGS_AUTOCONVERTPCM should convert the sample rate automatically (#215).
	if err := c.client.Initialize(_AUDCLNT_SHAREMODE_SHARED,
		_AUDCLNT_STREAMFLAGS_EVENTCALLBACK|_AUDCLNT_STREAMFLAGS_NOPERSIST|_AUDCLNT_STREAMFLAGS_AUTOCONVERTPCM,
		bufferSizeIn100ns, 0, f, nil); err != nil {
		return err
	}

	frames, err := c.client.GetBufferSize()
	if err != nil {
		return err
	}
	c.bufferFrames = frames

	if c.renderClient != nil {
		c.renderClient.Release()
		c.renderClient = nil
	}

	renderClient, err := c.client.GetService(&uuidIAudioRenderClient)
	if err != nil {
		return err
	}
	c.renderClient = (*_IAudioRenderClient)(renderClient)

	if err := c.client.SetEventHandle(c.sampleReadyEvent); err != nil {
		return err
	}

	// TODO: Should some errors be allowed? See WASAPIManager.cpp in the official example SimpleWASAPIPlaySound.

	if err := c.client.Start(); err != nil {
		return err
	}

	return nil
}

func (c *wasapiContext) loop() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// S_FALSE is returned when CoInitializeEx is nested. This is a successful case.
	if err := windows.CoInitializeEx(0, windows.COINIT_MULTITHREADED); err != nil && !errors.Is(err, syscall.Errno(windows.S_FALSE)) {
		_, _ = c.client.Stop()
		return err
	}
	// CoUninitialize should be called even when CoInitializeEx returns S_FALSE.
	defer windows.CoUninitialize()

	if err := c.loopOnRenderThread(); err != nil {
		_, _ = c.client.Stop()
		return err
	}

	return nil
}

func (c *wasapiContext) loopOnRenderThread() error {
	last := time.Now()
	for {
		c.suspendedCond.L.Lock()
		for c.suspended {
			c.suspendedCond.Wait()
		}
		c.suspendedCond.L.Unlock()

		evt, err := windows.WaitForSingleObject(c.sampleReadyEvent, windows.INFINITE)
		if err != nil {
			return err
		}
		if evt != windows.WAIT_OBJECT_0 {
			return fmt.Errorf("oto: WaitForSingleObject failed: returned value: %d", evt)
		}

		if err := c.writeOnRenderThread(); err != nil {
			return err
		}

		// Checking the current default audio device might be an expensive operation.
		// Check this repeatedly but with some time interval.
		if now := time.Now(); now.Sub(last) >= 500*time.Millisecond {
			switched, err := c.isDeviceSwitched()
			if err != nil {
				return err
			}
			if switched {
				return errDeviceSwitched
			}
			last = now
		}
	}
}

func (c *wasapiContext) writeOnRenderThread() error {
	c.m.Lock()
	defer c.m.Unlock()

	paddingFrames, err := c.client.GetCurrentPadding()
	if err != nil {
		return err
	}

	frames := c.bufferFrames - paddingFrames
	if frames <= 0 {
		return nil
	}

	// Get the destination buffer.
	dstBuf, err := c.renderClient.GetBuffer(frames)
	if err != nil {
		return err
	}

	// Calculate the buffer size.
	if buflen := int(frames) * c.channelCount; cap(c.buf) < buflen {
		c.buf = make([]float32, buflen)
	} else {
		c.buf = c.buf[:buflen]
	}

	// Read the buffer from the players.
	c.mux.ReadFloat32s(c.buf)

	// Copy the read buf to the destination buffer.
	copy(unsafe.Slice((*float32)(unsafe.Pointer(dstBuf)), len(c.buf)), c.buf)

	// Release the buffer.
	if err := c.renderClient.ReleaseBuffer(frames, 0); err != nil {
		return err
	}

	c.buf = c.buf[:0]
	return nil
}

func (c *wasapiContext) Suspend() error {
	c.suspendedCond.L.Lock()
	c.suspended = true
	c.suspendedCond.L.Unlock()
	c.suspendedCond.Signal()

	return nil
}

func (c *wasapiContext) Resume() error {
	c.suspendedCond.L.Lock()
	c.suspended = false
	c.suspendedCond.L.Unlock()
	c.suspendedCond.Signal()

	return nil
}

func (c *wasapiContext) isSuspended() bool {
	c.suspendedCond.L.Lock()
	defer c.suspendedCond.L.Unlock()
	return c.suspended
}

func (c *wasapiContext) Err() error {
	return c.err.Load()
}

func (c *wasapiContext) restart() error {
	// Probably the driver is missing temporarily e.g. plugging out the headset.
	// Recreate the device.

retry:
	c.suspendedCond.L.Lock()
	for c.suspended {
		c.suspendedCond.Wait()
	}
	c.suspendedCond.L.Unlock()

	if err := c.start(); err != nil {
		// When a device is switched, the new device might not support the desired format,
		// or all the audio devices might be disconnected.
		// Instead of aborting this context, let's wait for the next device switch.
		if !errors.Is(err, errFormatNotSupported) && !errors.Is(err, errDeviceNotFound) {
			return err
		}

		// Just read the buffer and discard it. Then, retry to search the device.
		var buf32 [4096]float32
		sleep := time.Duration(float64(time.Second) * float64(len(buf32)) / float64(c.channelCount) / float64(c.sampleRate))
		c.mux.ReadFloat32s(buf32[:])
		time.Sleep(sleep)
		goto retry
	}
	return nil
}
