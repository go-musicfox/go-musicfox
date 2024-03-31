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

package oto

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	winmm = windows.NewLazySystemDLL("winmm")
)

var (
	procWaveOutOpen            = winmm.NewProc("waveOutOpen")
	procWaveOutClose           = winmm.NewProc("waveOutClose")
	procWaveOutPrepareHeader   = winmm.NewProc("waveOutPrepareHeader")
	procWaveOutUnprepareHeader = winmm.NewProc("waveOutUnprepareHeader")
	procWaveOutWrite           = winmm.NewProc("waveOutWrite")
)

type _WAVEHDR struct {
	lpData          uintptr
	dwBufferLength  uint32
	dwBytesRecorded uint32
	dwUser          uintptr
	dwFlags         uint32
	dwLoops         uint32
	lpNext          uintptr
	reserved        uintptr
}

type _WAVEFORMATEX struct {
	wFormatTag      uint16
	nChannels       uint16
	nSamplesPerSec  uint32
	nAvgBytesPerSec uint32
	nBlockAlign     uint16
	wBitsPerSample  uint16
	cbSize          uint16
}

const (
	_WAVE_FORMAT_IEEE_FLOAT = 3
	_WHDR_INQUEUE           = 16
)

type _MMRESULT uint

const (
	_MMSYSERR_NOERROR       _MMRESULT = 0
	_MMSYSERR_ERROR         _MMRESULT = 1
	_MMSYSERR_BADDEVICEID   _MMRESULT = 2
	_MMSYSERR_ALLOCATED     _MMRESULT = 4
	_MMSYSERR_INVALIDHANDLE _MMRESULT = 5
	_MMSYSERR_NODRIVER      _MMRESULT = 6
	_MMSYSERR_NOMEM         _MMRESULT = 7
	_WAVERR_BADFORMAT       _MMRESULT = 32
	_WAVERR_STILLPLAYING    _MMRESULT = 33
	_WAVERR_UNPREPARED      _MMRESULT = 34
	_WAVERR_SYNC            _MMRESULT = 35
)

func (m _MMRESULT) Error() string {
	switch m {
	case _MMSYSERR_NOERROR:
		return "MMSYSERR_NOERROR"
	case _MMSYSERR_ERROR:
		return "MMSYSERR_ERROR"
	case _MMSYSERR_BADDEVICEID:
		return "MMSYSERR_BADDEVICEID"
	case _MMSYSERR_ALLOCATED:
		return "MMSYSERR_ALLOCATED"
	case _MMSYSERR_INVALIDHANDLE:
		return "MMSYSERR_INVALIDHANDLE"
	case _MMSYSERR_NODRIVER:
		return "MMSYSERR_NODRIVER"
	case _MMSYSERR_NOMEM:
		return "MMSYSERR_NOMEM"
	case _WAVERR_BADFORMAT:
		return "WAVERR_BADFORMAT"
	case _WAVERR_STILLPLAYING:
		return "WAVERR_STILLPLAYING"
	case _WAVERR_UNPREPARED:
		return "WAVERR_UNPREPARED"
	case _WAVERR_SYNC:
		return "WAVERR_SYNC"
	}
	return fmt.Sprintf("MMRESULT (%d)", m)
}

func waveOutOpen(f *_WAVEFORMATEX, callback uintptr) (uintptr, error) {
	const (
		waveMapper       = 0xffffffff
		callbackFunction = 0x30000
	)
	var w uintptr
	var fdwOpen uintptr
	if callback != 0 {
		fdwOpen |= callbackFunction
	}
	r, _, e := procWaveOutOpen.Call(uintptr(unsafe.Pointer(&w)), waveMapper, uintptr(unsafe.Pointer(f)),
		callback, 0, fdwOpen)
	runtime.KeepAlive(f)
	if _MMRESULT(r) != _MMSYSERR_NOERROR {
		if e != nil && e != windows.ERROR_SUCCESS {
			return 0, fmt.Errorf("oto: waveOutOpen failed: %w", e)
		}
		return 0, fmt.Errorf("oto: waveOutOpen failed: %w", _MMRESULT(r))
	}
	return w, nil
}

func waveOutClose(hwo uintptr) error {
	r, _, e := procWaveOutClose.Call(hwo)
	if _MMRESULT(r) != _MMSYSERR_NOERROR {
		if e != nil && e != windows.ERROR_SUCCESS {
			return fmt.Errorf("oto: waveOutClose failed: %w", e)
		}
		return fmt.Errorf("oto: waveOutClose failed: %w", _MMRESULT(r))
	}
	return nil
}

func waveOutPrepareHeader(hwo uintptr, pwh *_WAVEHDR) error {
	r, _, e := procWaveOutPrepareHeader.Call(hwo, uintptr(unsafe.Pointer(pwh)), unsafe.Sizeof(_WAVEHDR{}))
	runtime.KeepAlive(pwh)
	if _MMRESULT(r) != _MMSYSERR_NOERROR {
		if e != nil && e != windows.ERROR_SUCCESS {
			return fmt.Errorf("oto: waveOutPrepareHeader failed: %w", e)
		}
		return fmt.Errorf("oto: waveOutPrepareHeader failed: %w", _MMRESULT(r))
	}
	return nil
}

func waveOutUnprepareHeader(hwo uintptr, pwh *_WAVEHDR) error {
	r, _, e := procWaveOutUnprepareHeader.Call(hwo, uintptr(unsafe.Pointer(pwh)), unsafe.Sizeof(_WAVEHDR{}))
	runtime.KeepAlive(pwh)
	if _MMRESULT(r) != _MMSYSERR_NOERROR {
		if e != nil && e != windows.ERROR_SUCCESS {
			return fmt.Errorf("oto: waveOutUnprepareHeader failed: %w", e)
		}
		return fmt.Errorf("oto: waveOutUnprepareHeader failed: %w", _MMRESULT(r))
	}
	return nil
}

func waveOutWrite(hwo uintptr, pwh *_WAVEHDR) error {
	r, _, e := procWaveOutWrite.Call(hwo, uintptr(unsafe.Pointer(pwh)), unsafe.Sizeof(_WAVEHDR{}))
	runtime.KeepAlive(pwh)
	if _MMRESULT(r) != _MMSYSERR_NOERROR {
		if e != nil && e != windows.ERROR_SUCCESS {
			return fmt.Errorf("oto: waveOutWrite failed: %w", e)
		}
		return fmt.Errorf("oto: waveOutWrite failed: %w", _MMRESULT(r))
	}
	return nil
}
