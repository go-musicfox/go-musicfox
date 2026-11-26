//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2EnvironmentOptions7Vtbl struct {
	IUnknownVtbl
	GetChannelSearchKind ComProc
	PutChannelSearchKind ComProc
	GetReleaseChannels   ComProc
	PutReleaseChannels   ComProc
}

type ICoreWebView2EnvironmentOptions7 struct {
	Vtbl *ICoreWebView2EnvironmentOptions7Vtbl
}

func (i *ICoreWebView2EnvironmentOptions7) AddRef() uintptr {
	refCounter, _, _ := i.Vtbl.AddRef.Call(uintptr(unsafe.Pointer(i)))
	return refCounter
}

// Release drops one reference and returns the new count.
//
// AddRef was generated for all 252 interfaces and Release for none, which left every caller of a
// Get<Interface>() accessor leaking: QueryInterface AddRefs on success and there was no matching
// call to make, short of reaching through the embedded IUnknownVtbl for CallRelease. Additive, so
// no existing caller changes.
//
// Not generated for handler interfaces: those are objects WE implement and hand to WebView2, so
// their lifetime is the Go object's, and calling through the vtable would re-enter our own impl.
func (i *ICoreWebView2EnvironmentOptions7) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2EnvironmentOptions7) GetChannelSearchKind() (COREWEBVIEW2_CHANNEL_SEARCH_KIND, error) {

	var value COREWEBVIEW2_CHANNEL_SEARCH_KIND

	hr, _, _ := i.Vtbl.GetChannelSearchKind.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2EnvironmentOptions7) PutChannelSearchKind(value COREWEBVIEW2_CHANNEL_SEARCH_KIND) error {

	hr, _, _ := i.Vtbl.PutChannelSearchKind.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(value),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2EnvironmentOptions7) GetReleaseChannels() (COREWEBVIEW2_RELEASE_CHANNELS, error) {

	var value COREWEBVIEW2_RELEASE_CHANNELS

	hr, _, _ := i.Vtbl.GetReleaseChannels.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2EnvironmentOptions7) PutReleaseChannels(value COREWEBVIEW2_RELEASE_CHANNELS) error {

	hr, _, _ := i.Vtbl.PutReleaseChannels.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(value),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
