//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2NonClientRegionChangedEventArgsVtbl struct {
	IUnknownVtbl
	GetRegionKind ComProc
}

type ICoreWebView2NonClientRegionChangedEventArgs struct {
	Vtbl *ICoreWebView2NonClientRegionChangedEventArgsVtbl
}

func (i *ICoreWebView2NonClientRegionChangedEventArgs) AddRef() uintptr {
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
func (i *ICoreWebView2NonClientRegionChangedEventArgs) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2NonClientRegionChangedEventArgs) GetRegionKind() (COREWEBVIEW2_NON_CLIENT_REGION_KIND, error) {

	var value COREWEBVIEW2_NON_CLIENT_REGION_KIND

	hr, _, _ := i.Vtbl.GetRegionKind.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}
