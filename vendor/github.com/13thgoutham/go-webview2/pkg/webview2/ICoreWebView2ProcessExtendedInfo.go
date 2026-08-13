//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2ProcessExtendedInfoVtbl struct {
	IUnknownVtbl
	GetProcessInfo          ComProc
	GetAssociatedFrameInfos ComProc
}

type ICoreWebView2ProcessExtendedInfo struct {
	Vtbl *ICoreWebView2ProcessExtendedInfoVtbl
}

func (i *ICoreWebView2ProcessExtendedInfo) AddRef() uintptr {
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
func (i *ICoreWebView2ProcessExtendedInfo) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2ProcessExtendedInfo) GetProcessInfo() (*ICoreWebView2ProcessInfo, error) {

	var processInfo *ICoreWebView2ProcessInfo

	hr, _, _ := i.Vtbl.GetProcessInfo.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&processInfo)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return processInfo, nil
}

func (i *ICoreWebView2ProcessExtendedInfo) GetAssociatedFrameInfos() (*ICoreWebView2FrameInfoCollection, error) {

	var frames *ICoreWebView2FrameInfoCollection

	hr, _, _ := i.Vtbl.GetAssociatedFrameInfos.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&frames)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return frames, nil
}
