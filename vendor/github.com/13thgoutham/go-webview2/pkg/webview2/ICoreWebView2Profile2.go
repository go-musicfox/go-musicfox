//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"math"
	"syscall"
	"unsafe"
)

type ICoreWebView2Profile2Vtbl struct {
	ICoreWebView2ProfileVtbl
	ClearBrowsingData            ComProc
	ClearBrowsingDataInTimeRange ComProc
	ClearBrowsingDataAll         ComProc
}

type ICoreWebView2Profile2 struct {
	Vtbl *ICoreWebView2Profile2Vtbl
}

func (i *ICoreWebView2Profile2) AddRef() uintptr {
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
func (i *ICoreWebView2Profile2) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Profile) GetICoreWebView2Profile2() *ICoreWebView2Profile2 {
	var result *ICoreWebView2Profile2

	iidICoreWebView2Profile2 := NewGUID("{fa740d4b-5eae-4344-a8ad-74be31925397}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Profile2)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Profile2) ClearBrowsingData(dataKinds COREWEBVIEW2_BROWSING_DATA_KINDS, handler *ICoreWebView2ClearBrowsingDataCompletedHandler) error {

	hr, _, _ := i.Vtbl.ClearBrowsingData.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(dataKinds),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Profile2) ClearBrowsingDataInTimeRange(dataKinds COREWEBVIEW2_BROWSING_DATA_KINDS, startTime float64, endTime float64, handler *ICoreWebView2ClearBrowsingDataCompletedHandler) error {

	hr, _, _ := i.Vtbl.ClearBrowsingDataInTimeRange.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(dataKinds),
		uintptr(math.Float64bits(startTime)),
		uintptr(math.Float64bits(endTime)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Profile2) ClearBrowsingDataAll(handler *ICoreWebView2ClearBrowsingDataCompletedHandler) error {

	hr, _, _ := i.Vtbl.ClearBrowsingDataAll.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
