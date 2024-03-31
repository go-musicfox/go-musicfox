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
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	avAudioSessionErrorCodeCannotStartPlaying    = 0x21706c61 // '!pla'
	avAudioSessionErrorCodeCannotInterruptOthers = 0x21696e74 // '!int'
	avAudioSessionErrorCodeSiriIsRecording       = 0x73697269 // 'siri'
)

const (
	kAudioFormatLinearPCM = 0x6C70636D //'lpcm'
)

const (
	kAudioFormatFlagIsFloat = 1 << 0 // 0x1
)

type _AudioStreamBasicDescription struct {
	mSampleRate       float64
	mFormatID         uint32
	mFormatFlags      uint32
	mBytesPerPacket   uint32
	mFramesPerPacket  uint32
	mBytesPerFrame    uint32
	mChannelsPerFrame uint32
	mBitsPerChannel   uint32
	mReserved         uint32
}

type _AudioQueueRef uintptr

type _AudioTimeStamp uintptr

type _AudioStreamPacketDescription struct {
	mStartOffset            int64
	mVariableFramesInPacket uint32
	mDataByteSize           uint32
}

type _AudioQueueBufferRef *_AudioQueueBuffer

type _AudioQueueBuffer struct {
	mAudioDataBytesCapacity uint32
	mAudioData              uintptr // void*
	mAudioDataByteSize      uint32
	mUserData               uintptr // void*

	mPacketDescriptionCapacity uint32
	mPacketDescriptions        *_AudioStreamPacketDescription
	mPacketDescriptionCount    uint32
}

type _AudioQueueOutputCallback func(inUserData unsafe.Pointer, inAQ _AudioQueueRef, inBuffer _AudioQueueBufferRef)

func initializeAPI() error {
	toolbox, err := purego.Dlopen("/System/Library/Frameworks/AudioToolbox.framework/AudioToolbox", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	purego.RegisterLibFunc(&_AudioQueueNewOutput, toolbox, "AudioQueueNewOutput")
	purego.RegisterLibFunc(&_AudioQueueAllocateBuffer, toolbox, "AudioQueueAllocateBuffer")
	purego.RegisterLibFunc(&_AudioQueueEnqueueBuffer, toolbox, "AudioQueueEnqueueBuffer")
	purego.RegisterLibFunc(&_AudioQueueStart, toolbox, "AudioQueueStart")
	purego.RegisterLibFunc(&_AudioQueuePause, toolbox, "AudioQueuePause")
	return nil
}

var _AudioQueueNewOutput func(inFormat *_AudioStreamBasicDescription, inCallbackProc _AudioQueueOutputCallback, inUserData unsafe.Pointer, inCallbackRunLoop uintptr, inCallbackRunLoopMod uintptr, inFlags uint32, outAQ *_AudioQueueRef) uintptr

var _AudioQueueAllocateBuffer func(inAQ _AudioQueueRef, inBufferByteSize uint32, outBuffer *_AudioQueueBufferRef) uintptr

var _AudioQueueEnqueueBuffer func(inAQ _AudioQueueRef, inBuffer _AudioQueueBufferRef, inNumPacketDescs uint32, inPackets []_AudioStreamPacketDescription) uintptr

var _AudioQueueStart func(inAQ _AudioQueueRef, inStartTime *_AudioTimeStamp) uintptr

var _AudioQueuePause func(inAQ _AudioQueueRef) uintptr
