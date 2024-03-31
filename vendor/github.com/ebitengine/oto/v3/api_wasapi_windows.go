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
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ole32 = windows.NewLazySystemDLL("ole32")
)

var (
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
)

type _REFERENCE_TIME int64

var (
	uuidIAudioClient2       = windows.GUID{0x726778cd, 0xf60a, 0x4eda, [...]byte{0x82, 0xde, 0xe4, 0x76, 0x10, 0xcd, 0x78, 0xaa}}
	uuidIAudioRenderClient  = windows.GUID{0xf294acfc, 0x3146, 0x4483, [...]byte{0xa7, 0xbf, 0xad, 0xdc, 0xa7, 0xc2, 0x60, 0xe2}}
	uuidIMMDeviceEnumerator = windows.GUID{0xa95664d2, 0x9614, 0x4f35, [...]byte{0xa7, 0x46, 0xde, 0x8d, 0xb6, 0x36, 0x17, 0xe6}}
	uuidMMDeviceEnumerator  = windows.GUID{0xbcde0395, 0xe52f, 0x467c, [...]byte{0x8e, 0x3d, 0xc4, 0x57, 0x92, 0x91, 0x69, 0x2e}}
)

const (
	_AUDCLNT_STREAMFLAGS_AUTOCONVERTPCM = 0x80000000
	_AUDCLNT_STREAMFLAGS_EVENTCALLBACK  = 0x00040000
	_AUDCLNT_STREAMFLAGS_NOPERSIST      = 0x00080000
	_COINIT_APARTMENTTHREADED           = 0x2
	_COINIT_MULTITHREADED               = 0
	_REFTIMES_PER_SEC                   = 10000000
	_SPEAKER_FRONT_CENTER               = 0x4
	_SPEAKER_FRONT_LEFT                 = 0x1
	_SPEAKER_FRONT_RIGHT                = 0x2
	_WAVE_FORMAT_EXTENSIBLE             = 0xfffe
)

var (
	_KSDATAFORMAT_SUBTYPE_IEEE_FLOAT = windows.GUID{0x00000003, 0x0000, 0x0010, [...]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}
	_KSDATAFORMAT_SUBTYPE_PCM        = windows.GUID{0x00000001, 0x0000, 0x0010, [...]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}
)

type _AUDCLNT_ERR uint32

const (
	_AUDCLNT_E_DEVICE_INVALIDATED    _AUDCLNT_ERR = 0x88890004
	_AUDCLNT_E_NOT_INITIALIZED       _AUDCLNT_ERR = 0x88890001
	_AUDCLNT_E_RESOURCES_INVALIDATED _AUDCLNT_ERR = 0x88890026
)

func isAudclntErr(hresult uint32) bool {
	return hresult&0xffff0000 == (1<<31)|(windows.FACILITY_AUDCLNT<<16)
}

func (e _AUDCLNT_ERR) Error() string {
	switch e {
	case _AUDCLNT_E_DEVICE_INVALIDATED:
		return "AUDCLNT_E_DEVICE_INVALIDATED"
	case _AUDCLNT_E_RESOURCES_INVALIDATED:
		return "AUDCLNT_E_RESOURCES_INVALIDATED"
	default:
		return fmt.Sprintf("AUDCLNT_ERR(%d)", e)
	}
}

type _AUDCLNT_SHAREMODE int32

const (
	_AUDCLNT_SHAREMODE_SHARED    _AUDCLNT_SHAREMODE = 0
	_AUDCLNT_SHAREMODE_EXCLUSIVE _AUDCLNT_SHAREMODE = 1
)

type _AUDCLNT_STREAMOPTIONS int32

const (
	_AUDCLNT_STREAMOPTIONS_NONE         _AUDCLNT_STREAMOPTIONS = 0x0
	_AUDCLNT_STREAMOPTIONS_RAW          _AUDCLNT_STREAMOPTIONS = 0x1
	_AUDCLNT_STREAMOPTIONS_MATCH_FORMAT _AUDCLNT_STREAMOPTIONS = 0x2
	_AUDCLNT_STREAMOPTIONS_AMBISONICS   _AUDCLNT_STREAMOPTIONS = 0x4
)

type _AUDIO_STREAM_CATEGORY int32

const (
	_AudioCategory_Other                  _AUDIO_STREAM_CATEGORY = 0
	_AudioCategory_ForegroundOnlyMedia    _AUDIO_STREAM_CATEGORY = 1
	_AudioCategory_BackgroundCapableMedia _AUDIO_STREAM_CATEGORY = 2
	_AudioCategory_Communications         _AUDIO_STREAM_CATEGORY = 3
	_AudioCategory_Alerts                 _AUDIO_STREAM_CATEGORY = 4
	_AudioCategory_SoundEffects           _AUDIO_STREAM_CATEGORY = 5
	_AudioCategory_GameEffects            _AUDIO_STREAM_CATEGORY = 6
	_AudioCategory_GameMedia              _AUDIO_STREAM_CATEGORY = 7
	_AudioCategory_GameChat               _AUDIO_STREAM_CATEGORY = 8
	_AudioCategory_Speech                 _AUDIO_STREAM_CATEGORY = 9
	_AudioCategory_Movie                  _AUDIO_STREAM_CATEGORY = 10
	_AudioCategory_Media                  _AUDIO_STREAM_CATEGORY = 11
)

type _CLSCTX int32

const (
	_CLSCTX_INPROC_SERVER  _CLSCTX = 0x00000001
	_CLSCTX_INPROC_HANDLER _CLSCTX = 0x00000002
	_CLSCTX_LOCAL_SERVER   _CLSCTX = 0x00000004
	_CLSCTX_REMOTE_SERVER  _CLSCTX = 0x00000010
	_CLSCTX_ALL                    = _CLSCTX_INPROC_SERVER | _CLSCTX_INPROC_HANDLER | _CLSCTX_LOCAL_SERVER | _CLSCTX_REMOTE_SERVER
)

type _EDataFlow int32

const (
	eRender _EDataFlow = 0
)

type _ERole int32

const (
	eConsole _ERole = 0
)

type _WIN32_ERR uint32

const (
	_E_NOTFOUND _WIN32_ERR = 0x80070490
)

func isWin32Err(hresult uint32) bool {
	return hresult&0xffff0000 == (1<<31)|(windows.FACILITY_WIN32<<16)
}

func (e _WIN32_ERR) Error() string {
	switch e {
	case _E_NOTFOUND:
		return "E_NOTFOUND"
	default:
		return fmt.Sprintf("HRESULT(%d)", e)
	}
}

type _AudioClientProperties struct {
	cbSize     uint32
	bIsOffload int32
	eCategory  _AUDIO_STREAM_CATEGORY
	Options    _AUDCLNT_STREAMOPTIONS
}

type _PROPVARIANT struct {
	// TODO: Implmeent this
}

type _WAVEFORMATEXTENSIBLE struct {
	wFormatTag      uint16
	nChannels       uint16
	nSamplesPerSec  uint32
	nAvgBytesPerSec uint32
	nBlockAlign     uint16
	wBitsPerSample  uint16
	cbSize          uint16
	Samples         uint16 // union
	dwChannelMask   uint32
	SubFormat       windows.GUID
}

func _CoCreateInstance(rclsid *windows.GUID, pUnkOuter unsafe.Pointer, dwClsContext uint32, riid *windows.GUID) (unsafe.Pointer, error) {
	var v unsafe.Pointer
	r, _, _ := procCoCreateInstance.Call(uintptr(unsafe.Pointer(rclsid)), uintptr(pUnkOuter), uintptr(dwClsContext), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(&v)))
	runtime.KeepAlive(rclsid)
	runtime.KeepAlive(riid)
	if uint32(r) != uint32(windows.S_OK) {
		return nil, fmt.Errorf("oto: CoCreateInstance failed: HRESULT(%d)", uint32(r))
	}
	return v, nil
}

type _IAudioClient2 struct {
	vtbl *_IAudioClient2_Vtbl
}

type _IAudioClient2_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	Initialize          uintptr
	GetBufferSize       uintptr
	GetStreamLatency    uintptr
	GetCurrentPadding   uintptr
	IsFormatSupported   uintptr
	GetMixFormat        uintptr
	GetDevicePeriod     uintptr
	Start               uintptr
	Stop                uintptr
	Reset               uintptr
	SetEventHandle      uintptr
	GetService          uintptr
	IsOffloadCapable    uintptr
	SetClientProperties uintptr
	GetBufferSizeLimits uintptr
}

func (i *_IAudioClient2) GetBufferSize() (uint32, error) {
	var numBufferFrames uint32
	r, _, _ := syscall.Syscall(i.vtbl.GetBufferSize, 2, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&numBufferFrames)), 0)
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return 0, fmt.Errorf("oto: IAudioClient2::GetBufferSize failed: %w", _AUDCLNT_ERR(r))
		}
		return 0, fmt.Errorf("oto: IAudioClient2::GetBufferSize failed: HRESULT(%d)", uint32(r))
	}
	return numBufferFrames, nil
}

func (i *_IAudioClient2) GetCurrentPadding() (uint32, error) {
	var numPaddingFrames uint32
	r, _, _ := syscall.Syscall(i.vtbl.GetCurrentPadding, 2, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&numPaddingFrames)), 0)
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return 0, fmt.Errorf("oto: IAudioClient2::GetCurrentPadding failed: %w", _AUDCLNT_ERR(r))
		}
		return 0, fmt.Errorf("oto: IAudioClient2::GetCurrentPadding failed: HRESULT(%d)", uint32(r))
	}
	return numPaddingFrames, nil
}

func (i *_IAudioClient2) GetDevicePeriod() (_REFERENCE_TIME, _REFERENCE_TIME, error) {
	var defaultDevicePeriod _REFERENCE_TIME
	var minimumDevicePeriod _REFERENCE_TIME
	r, _, _ := syscall.Syscall(i.vtbl.GetDevicePeriod, 3, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&defaultDevicePeriod)), uintptr(unsafe.Pointer(&minimumDevicePeriod)))
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return 0, 0, fmt.Errorf("oto: IAudioClient2::GetDevicePeriod failed: %w", _AUDCLNT_ERR(r))
		}
		return 0, 0, fmt.Errorf("oto: IAudioClient2::GetDevicePeriod failed: HRESULT(%d)", uint32(r))
	}
	return defaultDevicePeriod, minimumDevicePeriod, nil
}

func (i *_IAudioClient2) GetService(riid *windows.GUID) (unsafe.Pointer, error) {
	var v unsafe.Pointer
	r, _, _ := syscall.Syscall(i.vtbl.GetService, 3, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(&v)))
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return nil, fmt.Errorf("oto: IAudioClient2::GetService failed: %w", _AUDCLNT_ERR(r))
		}
		return nil, fmt.Errorf("oto: IAudioClient2::GetService failed: HRESULT(%d)", uint32(r))
	}
	return v, nil
}

func (i *_IAudioClient2) Initialize(shareMode _AUDCLNT_SHAREMODE, streamFlags uint32, hnsBufferDuration _REFERENCE_TIME, hnsPeriodicity _REFERENCE_TIME, pFormat *_WAVEFORMATEXTENSIBLE, audioSessionGuid *windows.GUID) error {
	var r uintptr
	if unsafe.Sizeof(uintptr(0)) == 8 {
		// 64bits
		r, _, _ = syscall.Syscall9(i.vtbl.Initialize, 7, uintptr(unsafe.Pointer(i)),
			uintptr(shareMode), uintptr(streamFlags), uintptr(hnsBufferDuration),
			uintptr(hnsPeriodicity), uintptr(unsafe.Pointer(pFormat)), uintptr(unsafe.Pointer(audioSessionGuid)),
			0, 0)
	} else {
		// 32bits
		r, _, _ = syscall.Syscall9(i.vtbl.Initialize, 9, uintptr(unsafe.Pointer(i)),
			uintptr(shareMode), uintptr(streamFlags), uintptr(hnsBufferDuration),
			uintptr(hnsBufferDuration>>32), uintptr(hnsPeriodicity), uintptr(hnsPeriodicity>>32),
			uintptr(unsafe.Pointer(pFormat)), uintptr(unsafe.Pointer(audioSessionGuid)))
	}
	runtime.KeepAlive(pFormat)
	runtime.KeepAlive(audioSessionGuid)
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return fmt.Errorf("oto: IAudioClient2::Initialize failed: %w", _AUDCLNT_ERR(r))
		}
		return fmt.Errorf("oto: IAudioClient2::Initialize failed: HRESULT(%d)", uint32(r))
	}
	return nil
}

func (i *_IAudioClient2) IsFormatSupported(shareMode _AUDCLNT_SHAREMODE, pFormat *_WAVEFORMATEXTENSIBLE) (*_WAVEFORMATEXTENSIBLE, error) {
	var closestMatch *_WAVEFORMATEXTENSIBLE
	r, _, _ := syscall.Syscall6(i.vtbl.IsFormatSupported, 4, uintptr(unsafe.Pointer(i)),
		uintptr(shareMode), uintptr(unsafe.Pointer(pFormat)), uintptr(unsafe.Pointer(&closestMatch)),
		0, 0)
	if uint32(r) != uint32(windows.S_OK) {
		if uint32(r) == uint32(windows.S_FALSE) {
			var r _WAVEFORMATEXTENSIBLE
			if closestMatch != nil {
				r = *closestMatch
				windows.CoTaskMemFree(unsafe.Pointer(closestMatch))
			}
			return &r, nil
		}
		if isAudclntErr(uint32(r)) {
			return nil, fmt.Errorf("oto: IAudioClient2::IsFormatSupported failed: %w", _AUDCLNT_ERR(r))
		}
		return nil, fmt.Errorf("oto: IAudioClient2::IsFormatSupported failed: HRESULT(%d)", uint32(r))
	}
	return nil, nil
}

func (i *_IAudioClient2) Release() uint32 {
	r, _, _ := syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	return uint32(r)
}

func (i *_IAudioClient2) SetClientProperties(pProperties *_AudioClientProperties) error {
	r, _, _ := syscall.Syscall(i.vtbl.SetClientProperties, 2, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(pProperties)), 0)
	runtime.KeepAlive(pProperties)
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return fmt.Errorf("oto: IAudioClient2::SetClientProperties failed: %w", _AUDCLNT_ERR(r))
		}
		return fmt.Errorf("oto: IAudioClient2::SetClientProperties failed: HRESULT(%d)", uint32(r))
	}
	return nil
}

func (i *_IAudioClient2) SetEventHandle(eventHandle windows.Handle) error {
	r, _, _ := syscall.Syscall(i.vtbl.SetEventHandle, 2, uintptr(unsafe.Pointer(i)), uintptr(eventHandle), 0)
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return fmt.Errorf("oto: IAudioClient2::SetEventHandle failed: %w", _AUDCLNT_ERR(r))
		}
		return fmt.Errorf("oto: IAudioClient2::SetEventHandle failed: HRESULT(%d)", uint32(r))
	}
	return nil
}

func (i *_IAudioClient2) Start() error {
	r, _, _ := syscall.Syscall(i.vtbl.Start, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return fmt.Errorf("oto: IAudioClient2::Start failed: %w", _AUDCLNT_ERR(r))
		}
		return fmt.Errorf("oto: IAudioClient2::Start failed: HRESULT(%d)", uint32(r))
	}
	return nil
}

func (i *_IAudioClient2) Stop() (bool, error) {
	r, _, _ := syscall.Syscall(i.vtbl.Stop, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	if uint32(r) != uint32(windows.S_OK) && uint32(r) != uint32(windows.S_FALSE) {
		if isAudclntErr(uint32(r)) {
			return false, fmt.Errorf("oto: IAudioClient2::Stop failed: %w", _AUDCLNT_ERR(r))
		}
		return false, fmt.Errorf("oto: IAudioClient2::Stop failed: HRESULT(%d)", uint32(r))
	}
	return uint32(r) == uint32(windows.S_OK), nil
}

type _IAudioRenderClient struct {
	vtbl *_IAudioRenderClient_Vtbl
}

type _IAudioRenderClient_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetBuffer     uintptr
	ReleaseBuffer uintptr
}

func (i *_IAudioRenderClient) GetBuffer(numFramesRequested uint32) (*byte, error) {
	var data *byte
	r, _, _ := syscall.Syscall(i.vtbl.GetBuffer, 3, uintptr(unsafe.Pointer(i)), uintptr(numFramesRequested), uintptr(unsafe.Pointer(&data)))
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return nil, fmt.Errorf("oto: IAudioRenderClient::GetBuffer failed: %w", _AUDCLNT_ERR(r))
		}
		return nil, fmt.Errorf("oto: IAudioRenderClient::GetBuffer failed: HRESULT(%d)", uint32(r))
	}
	return data, nil
}

func (i *_IAudioRenderClient) Release() uint32 {
	r, _, _ := syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	return uint32(r)
}

func (i *_IAudioRenderClient) ReleaseBuffer(numFramesWritten uint32, dwFlags uint32) error {
	r, _, _ := syscall.Syscall(i.vtbl.ReleaseBuffer, 3, uintptr(unsafe.Pointer(i)), uintptr(numFramesWritten), uintptr(dwFlags))
	if uint32(r) != uint32(windows.S_OK) {
		if isAudclntErr(uint32(r)) {
			return fmt.Errorf("oto: IAudioRenderClient::ReleaseBuffer failed: %w", _AUDCLNT_ERR(r))
		}
		return fmt.Errorf("oto: IAudioRenderClient::ReleaseBuffer failed: HRESULT(%d)", uint32(r))
	}
	return nil
}

type _IMMDevice struct {
	vtbl *_IMMDevice_Vtbl
}

type _IMMDevice_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	Activate          uintptr
	OpenPropertyStore uintptr
	GetId             uintptr
	GetState          uintptr
}

func (i *_IMMDevice) Activate(iid *windows.GUID, dwClsCtx uint32, pActivationParams *_PROPVARIANT) (unsafe.Pointer, error) {
	var v unsafe.Pointer
	r, _, _ := syscall.Syscall6(i.vtbl.Activate, 5, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iid)), uintptr(dwClsCtx), uintptr(unsafe.Pointer(pActivationParams)), uintptr(unsafe.Pointer(&v)), 0)
	runtime.KeepAlive(iid)
	runtime.KeepAlive(pActivationParams)
	if uint32(r) != uint32(windows.S_OK) {
		return nil, fmt.Errorf("oto: IMMDevice::Activate failed: HRESULT(%d)", uint32(r))
	}
	return v, nil
}

func (i *_IMMDevice) GetId() (string, error) {
	var strId *uint16
	r, _, _ := syscall.Syscall(i.vtbl.GetId, 2, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&strId)), 0)
	if uint32(r) != uint32(windows.S_OK) {
		return "", fmt.Errorf("oto: IMMDevice::GetId failed: HRESULT(%d)", uint32(r))
	}
	return windows.UTF16PtrToString(strId), nil
}

func (i *_IMMDevice) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

type _IMMDeviceEnumerator struct {
	vtbl *_IMMDeviceEnumerator_Vtbl
}

type _IMMDeviceEnumerator_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	EnumAudioEndpoints                     uintptr
	GetDefaultAudioEndpoint                uintptr
	GetDevice                              uintptr
	RegisterEndpointNotificationCallback   uintptr
	UnregisterEndpointNotificationCallback uintptr
}

func (i *_IMMDeviceEnumerator) GetDefaultAudioEndPoint(dataFlow _EDataFlow, role _ERole) (*_IMMDevice, error) {
	var endPoint *_IMMDevice
	r, _, _ := syscall.Syscall6(i.vtbl.GetDefaultAudioEndpoint, 4, uintptr(unsafe.Pointer(i)),
		uintptr(dataFlow), uintptr(role), uintptr(unsafe.Pointer(&endPoint)), 0, 0)
	if uint32(r) != uint32(windows.S_OK) {
		if isWin32Err(uint32(r)) {
			return nil, fmt.Errorf("oto: IMMDeviceEnumerator::GetDefaultAudioEndPoint failed: %w", _E_NOTFOUND)
		}
		return nil, fmt.Errorf("oto: IMMDeviceEnumerator::GetDefaultAudioEndPoint failed: HRESULT(%d)", uint32(r))
	}
	return endPoint, nil
}

func (i *_IMMDeviceEnumerator) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}
