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

//go:build nintendosdk

package oto

// #cgo !darwin LDFLAGS: -Wl,-unresolved-symbols=ignore-all
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
//
// typedef void (*oto_OnReadCallbackType)(float* buf, size_t length);
//
// void oto_OpenAudio(int sample_rate, int channel_num, oto_OnReadCallbackType on_read_callback, int buffer_size_in_bytes);
//
// void oto_OnReadCallback(float* buf, size_t length);
// static void oto_OpenProxy(int sample_rate, int channel_num, int buffer_size_in_bytes) {
//   oto_OpenAudio(sample_rate, channel_num, oto_OnReadCallback, buffer_size_in_bytes);
// }
import "C"

import (
	"unsafe"

	"github.com/ebitengine/oto/v3/internal/mux"
)

//export oto_OnReadCallback
func oto_OnReadCallback(buf *C.float, length C.size_t) {
	theContext.mux.ReadFloat32s(unsafe.Slice((*float32)(unsafe.Pointer(buf)), length))
}

type context struct {
	mux *mux.Mux
}

var theContext *context

func newContext(sampleRate int, channelCount int, format mux.Format, bufferSizeInBytes int) (*context, chan struct{}, error) {
	ready := make(chan struct{})
	close(ready)

	c := &context{
		mux: mux.New(sampleRate, channelCount, format),
	}
	theContext = c
	C.oto_OpenProxy(C.int(sampleRate), C.int(channelCount), C.int(bufferSizeInBytes))

	return c, ready, nil
}

func (c *context) Suspend() error {
	// Do nothing so far.
	return nil
}

func (c *context) Resume() error {
	// Do nothing so far.
	return nil
}

func (c *context) Err() error {
	return nil
}
