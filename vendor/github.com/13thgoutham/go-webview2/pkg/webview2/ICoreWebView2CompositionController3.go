//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2CompositionController3Vtbl struct {
	ICoreWebView2CompositionController2Vtbl
	DragEnter ComProc
	DragLeave ComProc
	DragOver  ComProc
	Drop      ComProc
}

type ICoreWebView2CompositionController3 struct {
	Vtbl *ICoreWebView2CompositionController3Vtbl
}

func (i *ICoreWebView2CompositionController3) AddRef() uintptr {
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
func (i *ICoreWebView2CompositionController3) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2CompositionController) GetICoreWebView2CompositionController3() *ICoreWebView2CompositionController3 {
	var result *ICoreWebView2CompositionController3

	iidICoreWebView2CompositionController3 := NewGUID("{9570570e-4d76-4361-9ee1-f04d0dbdfb1e}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2CompositionController3)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2CompositionController3) DragEnter(dataObject *IDataObject, keyState uint32, point POINT) (uint32, error) {

	var effect uint32

	hr, _, _ := i.Vtbl.DragEnter.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(dataObject)),
		uintptr(keyState),
		uintptr(*(*uint64)(unsafe.Pointer(&point))),
		uintptr(unsafe.Pointer(&effect)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return effect, nil
}

func (i *ICoreWebView2CompositionController3) DragLeave() error {

	hr, _, _ := i.Vtbl.DragLeave.Call(
		uintptr(unsafe.Pointer(i)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2CompositionController3) DragOver(keyState uint32, point POINT) (uint32, error) {

	var effect uint32

	hr, _, _ := i.Vtbl.DragOver.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(keyState),
		uintptr(*(*uint64)(unsafe.Pointer(&point))),
		uintptr(unsafe.Pointer(&effect)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return effect, nil
}

func (i *ICoreWebView2CompositionController3) Drop(dataObject *IDataObject, keyState uint32, point POINT) (uint32, error) {

	var effect uint32

	hr, _, _ := i.Vtbl.Drop.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(dataObject)),
		uintptr(keyState),
		uintptr(*(*uint64)(unsafe.Pointer(&point))),
		uintptr(unsafe.Pointer(&effect)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return effect, nil
}
