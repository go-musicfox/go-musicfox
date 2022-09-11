// Copyright 2019 The Oto Authors
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

//go:build !js
// +build !js

package oto

// #cgo LDFLAGS: -framework AudioToolbox
//
// #import <AudioToolbox/AudioToolbox.h>
//
// void oto_render(void* inUserData, AudioQueueRef inAQ, AudioQueueBufferRef inBuffer);
//
// void oto_setNotificationHandler();
// bool oto_isBackground(void);
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

const baseQueueBufferSize = 1024

type audioInfo struct {
	channelNum      int
	bitDepthInBytes int
}

type driver struct {
	ctx           context.Context
	cancel        context.CancelFunc
	audioQueue    C.AudioQueueRef
	buf           []byte
	bufSize       int
	sampleRate    int
	audioInfo     *audioInfo
	buffers       []C.AudioQueueBufferRef
	paused        bool
	lastPauseTime time.Time

	err error

	chWrite   chan []byte
	chWritten chan int

	m sync.Mutex
}

var (
	theDriver *driver
	driverM   sync.Mutex
)

func setDriver(d *driver) {
	driverM.Lock()
	defer driverM.Unlock()

	if theDriver != nil && d != nil {
		panic("oto: at most one driver object can exist")
	}
	theDriver = d

	if d != nil {
		setNotificationHandler(d)
	}
}

func getDriver() *driver {
	driverM.Lock()
	defer driverM.Unlock()

	return theDriver
}

// TOOD: Convert the error code correctly.
// See https://stackoverflow.com/questions/2196869/how-do-you-convert-an-iphone-osstatus-code-to-something-useful

func newDriver(sampleRate, channelNum, bitDepthInBytes, bufferSizeInBytes int) (tryWriteCloser, error) {
	flags := C.kAudioFormatFlagIsPacked
	if bitDepthInBytes != 1 {
		flags |= C.kAudioFormatFlagIsSignedInteger
	}
	desc := C.AudioStreamBasicDescription{
		mSampleRate:       C.double(sampleRate),
		mFormatID:         C.kAudioFormatLinearPCM,
		mFormatFlags:      C.UInt32(flags),
		mBytesPerPacket:   C.UInt32(channelNum * bitDepthInBytes),
		mFramesPerPacket:  1,
		mBytesPerFrame:    C.UInt32(channelNum * bitDepthInBytes),
		mChannelsPerFrame: C.UInt32(channelNum),
		mBitsPerChannel:   C.UInt32(8 * bitDepthInBytes),
	}

	audioInfo := &audioInfo{
		channelNum:      channelNum,
		bitDepthInBytes: bitDepthInBytes,
	}

	var audioQueue C.AudioQueueRef
	if osstatus := C.AudioQueueNewOutput(
		&desc,
		(C.AudioQueueOutputCallback)(C.oto_render),
		unsafe.Pointer(audioInfo),
		(C.CFRunLoopRef)(0),
		(C.CFStringRef)(0),
		0,
		&audioQueue); osstatus != C.noErr {
		return nil, fmt.Errorf("oto: AudioQueueNewFormat with StreamFormat failed: %d", osstatus)
	}

	queueBufferSize := baseQueueBufferSize * channelNum * bitDepthInBytes
	nbuf := bufferSizeInBytes / queueBufferSize
	if nbuf <= 1 {
		nbuf = 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	d := &driver{
		ctx:        ctx,
		cancel:     cancel,
		audioQueue: audioQueue,
		sampleRate: sampleRate,
		audioInfo:  audioInfo,
		bufSize:    nbuf * queueBufferSize,
		buffers:    make([]C.AudioQueueBufferRef, nbuf),
		chWrite:    make(chan []byte),
		chWritten:  make(chan int),
	}
	runtime.SetFinalizer(d, (*driver).Close)
	// Set the driver before setting the rendering callback.
	setDriver(d)

	for i := 0; i < len(d.buffers); i++ {
		var buf C.AudioQueueBufferRef
		if osstatus := C.AudioQueueAllocateBuffer(audioQueue, C.UInt32(queueBufferSize), &buf); osstatus != C.noErr {
			return nil, fmt.Errorf("oto: AudioQueueAllocateBuffer failed: %d", osstatus)
		}
		d.buffers[i] = buf
		d.buffers[i].mAudioDataByteSize = C.UInt32(queueBufferSize)
		for j := 0; j < queueBufferSize; j++ {
			*(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(d.buffers[i].mAudioData)) + uintptr(j))) = 0
		}
		if osstatus := C.AudioQueueEnqueueBuffer(audioQueue, d.buffers[i], 0, nil); osstatus != C.noErr {
			return nil, fmt.Errorf("oto: AudioQueueEnqueueBuffer failed: %d", osstatus)
		}
	}

	for C.oto_isBackground() {
		time.Sleep(time.Second)
	}

	if osstatus := C.AudioQueueStart(audioQueue, nil); osstatus != C.noErr {
		return nil, fmt.Errorf("oto: AudioQueueStart failed: %d", osstatus)
	}

	return d, nil
}

//export oto_render
func oto_render(inUserData unsafe.Pointer, inAQ C.AudioQueueRef, inBuffer C.AudioQueueBufferRef) {
	audioInfo := (*audioInfo)(inUserData)
	queueBufferSize := baseQueueBufferSize * audioInfo.channelNum * audioInfo.bitDepthInBytes

	d := getDriver()

	var buf []byte

	// Set the timer. When the input does not come, the audio must be paused.
	s := time.Second * time.Duration(queueBufferSize) / time.Duration(d.sampleRate*d.audioInfo.channelNum*d.audioInfo.bitDepthInBytes)
	t := time.NewTicker(s)
	defer t.Stop()
	ch := t.C

	for len(buf) < queueBufferSize && d.ctx.Err() == nil {
		select {
		case dbuf := <-d.chWrite:
			for !d.resume(false) {
				d.m.Lock()
				err := d.err
				d.m.Unlock()
				if err != nil {
					return
				}

				time.Sleep(time.Second)
			}
			n := queueBufferSize - len(buf)
			if n > len(dbuf) {
				n = len(dbuf)
			}
			buf = append(buf, dbuf[:n]...)
			d.chWritten <- n
		case <-ch:
			d.pause()
			ch = nil
		case <-d.ctx.Done():
			// AudioQueue was closed, return immediately
			return
		}
	}

	// oto_render is a callback for AudioQueueNewOutput.
	// According to the observation, it may still called once after drvier.Close called.
	//
	// In most case, the assumption len(buf) == queueBufferSize may always be true.
	//
	// However, here is a special case:
	// After the `Close` called, and it is receiving chWrite and it may noticed
	// `d.ctx.Err() != nil`, it will jump out from loop directly.
	//
	// At this moment, since d.audioQueue is nil and the inBuffer may not processed,
	// We don't need to do any process to the inBuffer.
	//
	// Another consideration is when len(buf) != queueBufferSize, we may return directly.
	// In this case we still need to check whether d.audioQueue is valid inside enqueueBuffer
	// It may worthless.
	for i := 0; i < len(buf); i++ {
		*(*byte)(unsafe.Pointer(uintptr(inBuffer.mAudioData) + uintptr(i))) = buf[i]
	}
	// Do not update mAudioDataByteSize, or the buffer is not used correctly any more.

	d.enqueueBuffer(inBuffer)
}

func (d *driver) TryWrite(data []byte) (int, error) {
	d.m.Lock()
	err := d.err
	d.m.Unlock()
	if err != nil {
		return 0, err
	}

	n := d.bufSize - len(d.buf)
	if n > len(data) {
		n = len(data)
	}
	d.buf = append(d.buf, data[:n]...)
	// Use the buffer only when the buffer length is enough to avoid choppy sound.
	queueBufferSize := baseQueueBufferSize * d.audioInfo.channelNum * d.audioInfo.bitDepthInBytes
	for len(d.buf) >= queueBufferSize {
		d.chWrite <- d.buf
		n := <-d.chWritten
		d.buf = d.buf[n:]
	}
	return n, nil
}

func (d *driver) Close() error {
	d.m.Lock()
	defer d.m.Unlock()

	runtime.SetFinalizer(d, nil)

	// notify to close any (oto_render in this case) running progress
	d.cancel()

	if osstatus := C.AudioQueueStop(d.audioQueue, C.false); osstatus != C.noErr {
		return fmt.Errorf("oto: AudioQueueStop failed: %d", osstatus)
	}
	if osstatus := C.AudioQueueDispose(d.audioQueue, C.false); osstatus != C.noErr {
		return fmt.Errorf("oto: AudioQueueDispose failed: %d", osstatus)
	}
	d.audioQueue = nil
	setDriver(nil)
	return nil
}

func (d *driver) enqueueBuffer(buffer C.AudioQueueBufferRef) {
	d.m.Lock()
	defer d.m.Unlock()

	// avoid to enqueue buffer to a closed audio queue
	if d.ctx.Err() != nil {
		return
	}

	if osstatus := C.AudioQueueEnqueueBuffer(d.audioQueue, buffer, 0, nil); osstatus != C.noErr && d.err == nil {
		d.err = fmt.Errorf("oto: AudioQueueEnqueueBuffer failed: %d", osstatus)
		return
	}
}

func (d *driver) resume(afterSleep bool) bool {
	d.m.Lock()
	defer d.m.Unlock()

	// Audio doesn't work soon after recovering from sleeping. Wait for a while
	// (hajimehoshi/ebiten#1259).
	if afterSleep {
		// After short-time sleeping, 500ms more sleeping is enough. However, after long-time sleeping, it
		// looks like 1 second more sleeping are required (hajimehoshi/ebiten#1280).
		// This is tested on MacBook Pro 2020 macOS 10.15.6.
		if time.Now().Sub(d.lastPauseTime) < 30*time.Second {
			time.Sleep(500 * time.Millisecond)
		} else {
			time.Sleep(time.Second)
		}
	}

	if C.oto_isBackground() {
		return false
	}

	if osstatus := C.AudioQueueStart(d.audioQueue, nil); osstatus != C.noErr && d.err == nil {
		d.err = fmt.Errorf("oto: AudioQueueStart for resuming failed: %d", osstatus)
		return false
	}
	d.paused = false
	return true
}

func (d *driver) pause() {
	d.m.Lock()
	defer d.m.Unlock()

	if d.paused {
		return
	}
	if osstatus := C.AudioQueuePause(d.audioQueue); osstatus != C.noErr && d.err == nil {
		d.err = fmt.Errorf("oto: AudioQueuePause failed: %d", osstatus)
		return
	}
	d.paused = true
	d.lastPauseTime = time.Now()
}

func (d *driver) setError(err error) {
	d.m.Lock()
	defer d.m.Unlock()

	if theDriver.err != nil {
		return
	}
	theDriver.err = err
}

func (d *driver) tryWriteCanReturnWithoutWaiting() bool {
	return true
}

func setNotificationHandler(driver *driver) {
	C.oto_setNotificationHandler()
}

//export oto_setGlobalPause
func oto_setGlobalPause() {
	theDriver.pause()
}

//export oto_setGlobalResume
func oto_setGlobalResume() {
	theDriver.resume(true)
}

//export oto_setErrorByNotification
func oto_setErrorByNotification(s C.OSStatus, from *C.char) {
	gofrom := C.GoString(from)
	theDriver.setError(fmt.Errorf("oto: %s at notification failed: %d", gofrom, s))
}
