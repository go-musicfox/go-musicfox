// Copyright 2021 The Oto Authors
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

//go:build !android && !darwin && !js && !windows && !nintendosdk

package oto

// #cgo pkg-config: alsa
//
// #include <alsa/asoundlib.h>
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/oto/v3/internal/mux"
)

type context struct {
	channelCount int

	suspended bool

	handle *C.snd_pcm_t

	cond *sync.Cond

	mux *mux.Mux
	err atomicError

	ready chan struct{}
}

var theContext *context

func alsaError(name string, err C.int) error {
	return fmt.Errorf("oto: ALSA error at %s: %s", name, C.GoString(C.snd_strerror(err)))
}

func deviceCandidates() []string {
	const getAllDevices = -1

	cPCMInterfaceName := C.CString("pcm")
	defer C.free(unsafe.Pointer(cPCMInterfaceName))

	var hints *unsafe.Pointer
	err := C.snd_device_name_hint(getAllDevices, cPCMInterfaceName, &hints)
	if err != 0 {
		return []string{"default", "plug:default"}
	}
	defer C.snd_device_name_free_hint(hints)

	var devices []string

	cIoHintName := C.CString("IOID")
	defer C.free(unsafe.Pointer(cIoHintName))
	cNameHintName := C.CString("NAME")
	defer C.free(unsafe.Pointer(cNameHintName))

	for it := hints; *it != nil; it = (*unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(it)) + unsafe.Sizeof(uintptr(0)))) {
		io := C.snd_device_name_get_hint(*it, cIoHintName)
		defer func() {
			if io != nil {
				C.free(unsafe.Pointer(io))
			}
		}()
		if C.GoString(io) == "Input" {
			continue
		}

		name := C.snd_device_name_get_hint(*it, cNameHintName)
		defer func() {
			if name != nil {
				C.free(unsafe.Pointer(name))
			}
		}()
		if name == nil {
			continue
		}
		goName := C.GoString(name)
		if goName == "null" {
			continue
		}
		if goName == "default" {
			continue
		}
		devices = append(devices, goName)
	}

	devices = append([]string{"default", "plug:default"}, devices...)

	return devices
}

func newContext(sampleRate int, channelCount int, format mux.Format, bufferSizeInBytes int) (*context, chan struct{}, error) {
	c := &context{
		channelCount: channelCount,
		cond:         sync.NewCond(&sync.Mutex{}),
		mux:          mux.New(sampleRate, channelCount, format),
		ready:        make(chan struct{}),
	}
	theContext = c

	go func() {
		defer close(c.ready)

		// Open a default ALSA audio device for blocking stream playback
		type openError struct {
			device string
			err    C.int
		}
		var openErrs []openError
		var found bool

		for _, name := range deviceCandidates() {
			cname := C.CString(name)
			defer C.free(unsafe.Pointer(cname))
			if err := C.snd_pcm_open(&c.handle, cname, C.SND_PCM_STREAM_PLAYBACK, 0); err < 0 {
				openErrs = append(openErrs, openError{
					device: name,
					err:    err,
				})
				continue
			}
			found = true
			break
		}
		if !found {
			var msgs []string
			for _, e := range openErrs {
				msgs = append(msgs, fmt.Sprintf("%q: %s", e.device, C.GoString(C.snd_strerror(e.err))))
			}
			c.err.TryStore(fmt.Errorf("oto: ALSA error at snd_pcm_open: %s", strings.Join(msgs, ", ")))
			return
		}

		// TODO: Should snd_pcm_hw_params_set_periods be called explicitly?
		const periods = 2
		var periodSize C.snd_pcm_uframes_t
		if bufferSizeInBytes != 0 {
			periodSize = C.snd_pcm_uframes_t(bufferSizeInBytes / (channelCount * 4 * periods))
		} else {
			periodSize = C.snd_pcm_uframes_t(1024)
		}
		bufferSize := periodSize * periods
		if err := c.alsaPcmHwParams(sampleRate, channelCount, &bufferSize, &periodSize); err != nil {
			c.err.TryStore(err)
			return
		}

		go func() {
			buf32 := make([]float32, int(periodSize)*channelCount)
			for {
				if !c.readAndWrite(buf32) {
					return
				}
			}
		}()
	}()

	return c, c.ready, nil
}

func (c *context) alsaPcmHwParams(sampleRate, channelCount int, bufferSize, periodSize *C.snd_pcm_uframes_t) error {
	var params *C.snd_pcm_hw_params_t
	C.snd_pcm_hw_params_malloc(&params)
	defer C.free(unsafe.Pointer(params))

	if err := C.snd_pcm_hw_params_any(c.handle, params); err < 0 {
		return alsaError("snd_pcm_hw_params_any", err)
	}
	if err := C.snd_pcm_hw_params_set_access(c.handle, params, C.SND_PCM_ACCESS_RW_INTERLEAVED); err < 0 {
		return alsaError("snd_pcm_hw_params_set_access", err)
	}
	if err := C.snd_pcm_hw_params_set_format(c.handle, params, C.SND_PCM_FORMAT_FLOAT_LE); err < 0 {
		return alsaError("snd_pcm_hw_params_set_format", err)
	}
	if err := C.snd_pcm_hw_params_set_channels(c.handle, params, C.unsigned(channelCount)); err < 0 {
		return alsaError("snd_pcm_hw_params_set_channels", err)
	}
	if err := C.snd_pcm_hw_params_set_rate_resample(c.handle, params, 1); err < 0 {
		return alsaError("snd_pcm_hw_params_set_rate_resample", err)
	}
	sr := C.unsigned(sampleRate)
	if err := C.snd_pcm_hw_params_set_rate_near(c.handle, params, &sr, nil); err < 0 {
		return alsaError("snd_pcm_hw_params_set_rate_near", err)
	}
	if err := C.snd_pcm_hw_params_set_buffer_size_near(c.handle, params, bufferSize); err < 0 {
		return alsaError("snd_pcm_hw_params_set_buffer_size_near", err)
	}
	if err := C.snd_pcm_hw_params_set_period_size_near(c.handle, params, periodSize, nil); err < 0 {
		return alsaError("snd_pcm_hw_params_set_period_size_near", err)
	}
	if err := C.snd_pcm_hw_params(c.handle, params); err < 0 {
		return alsaError("snd_pcm_hw_params", err)
	}
	return nil
}

func (c *context) readAndWrite(buf32 []float32) bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for c.suspended && c.err.Load() == nil {
		c.cond.Wait()
	}
	if c.err.Load() != nil {
		return false
	}

	c.mux.ReadFloat32s(buf32)

	for len(buf32) > 0 {
		n := C.snd_pcm_writei(c.handle, unsafe.Pointer(&buf32[0]), C.snd_pcm_uframes_t(len(buf32)/c.channelCount))
		if n < 0 {
			n = C.long(C.snd_pcm_recover(c.handle, C.int(n), 1))
		}
		if n < 0 {
			c.err.TryStore(alsaError("snd_pcm_writei or snd_pcm_recover", C.int(n)))
			return false
		}
		buf32 = buf32[int(n)*c.channelCount:]
	}
	return true
}

func (c *context) Suspend() error {
	<-c.ready

	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	if err := c.err.Load(); err != nil {
		return err.(error)
	}

	c.suspended = true

	// Do not use snd_pcm_pause as not all devices support this.
	// Do not use snd_pcm_drop as this might hang (https://github.com/libsdl-org/SDL/blob/a5c610b0a3857d3138f3f3da1f6dc3172c5ea4a8/src/audio/alsa/SDL_alsa_audio.c#L478).
	return nil
}

func (c *context) Resume() error {
	<-c.ready

	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	if err := c.err.Load(); err != nil {
		return err.(error)
	}

	c.suspended = false
	c.cond.Signal()
	return nil
}

func (c *context) Err() error {
	if err := c.err.Load(); err != nil {
		return err.(error)
	}
	return nil
}
