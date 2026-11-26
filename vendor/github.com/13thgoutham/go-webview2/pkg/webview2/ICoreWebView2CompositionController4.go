//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2CompositionController4Vtbl struct {
	ICoreWebView2CompositionController3Vtbl
	GetNonClientRegionAtPoint    ComProc
	QueryNonClientRegion         ComProc
	AddNonClientRegionChanged    ComProc
	RemoveNonClientRegionChanged ComProc
}

type ICoreWebView2CompositionController4 struct {
	Vtbl *ICoreWebView2CompositionController4Vtbl
}

func (i *ICoreWebView2CompositionController4) AddRef() uintptr {
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
func (i *ICoreWebView2CompositionController4) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2CompositionController) GetICoreWebView2CompositionController4() *ICoreWebView2CompositionController4 {
	var result *ICoreWebView2CompositionController4

	iidICoreWebView2CompositionController4 := NewGUID("{7C367B9B-3D2B-450F-9E58-D61A20F486AA}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2CompositionController4)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2CompositionController4) GetNonClientRegionAtPoint(point POINT) (COREWEBVIEW2_NON_CLIENT_REGION_KIND, error) {

	var value COREWEBVIEW2_NON_CLIENT_REGION_KIND

	hr, _, _ := i.Vtbl.GetNonClientRegionAtPoint.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&point))),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2CompositionController4) QueryNonClientRegion(kind COREWEBVIEW2_NON_CLIENT_REGION_KIND) (*ICoreWebView2RegionRectCollectionView, error) {

	var rects *ICoreWebView2RegionRectCollectionView

	hr, _, _ := i.Vtbl.QueryNonClientRegion.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(kind),
		uintptr(unsafe.Pointer(&rects)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return rects, nil
}

func (i *ICoreWebView2CompositionController4) AddNonClientRegionChanged(eventHandler *ICoreWebView2NonClientRegionChangedEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddNonClientRegionChanged.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2CompositionController4) RemoveNonClientRegionChanged(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveNonClientRegionChanged.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
