//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2RegionRectCollectionViewVtbl struct {
	IUnknownVtbl
	GetCount        ComProc
	GetValueAtIndex ComProc
}

type ICoreWebView2RegionRectCollectionView struct {
	Vtbl *ICoreWebView2RegionRectCollectionViewVtbl
}

func (i *ICoreWebView2RegionRectCollectionView) AddRef() uintptr {
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
func (i *ICoreWebView2RegionRectCollectionView) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2RegionRectCollectionView) GetCount() (uint32, error) {

	var value uint32

	hr, _, _ := i.Vtbl.GetCount.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2RegionRectCollectionView) GetValueAtIndex(index uint32) (RECT, error) {

	var value RECT

	hr, _, _ := i.Vtbl.GetValueAtIndex.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(index),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return RECT{}, syscall.Errno(hr)
	}
	return value, nil
}
