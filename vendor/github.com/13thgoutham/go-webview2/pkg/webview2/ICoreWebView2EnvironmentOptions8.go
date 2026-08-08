//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2EnvironmentOptions8Vtbl struct {
	IUnknownVtbl
	GetScrollBarStyle ComProc
	PutScrollBarStyle ComProc
}

type ICoreWebView2EnvironmentOptions8 struct {
	Vtbl *ICoreWebView2EnvironmentOptions8Vtbl
}

func (i *ICoreWebView2EnvironmentOptions8) AddRef() uintptr {
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
func (i *ICoreWebView2EnvironmentOptions8) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2EnvironmentOptions8) GetScrollBarStyle() (COREWEBVIEW2_SCROLLBAR_STYLE, error) {

	var value COREWEBVIEW2_SCROLLBAR_STYLE

	hr, _, _ := i.Vtbl.GetScrollBarStyle.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2EnvironmentOptions8) PutScrollBarStyle(value COREWEBVIEW2_SCROLLBAR_STYLE) error {

	hr, _, _ := i.Vtbl.PutScrollBarStyle.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(value),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
