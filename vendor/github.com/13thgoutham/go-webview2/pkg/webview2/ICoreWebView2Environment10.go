//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Environment10Vtbl struct {
	ICoreWebView2Environment9Vtbl
	CreateCoreWebView2ControllerOptions                ComProc
	CreateCoreWebView2ControllerWithOptions            ComProc
	CreateCoreWebView2CompositionControllerWithOptions ComProc
}

type ICoreWebView2Environment10 struct {
	Vtbl *ICoreWebView2Environment10Vtbl
}

func (i *ICoreWebView2Environment10) AddRef() uintptr {
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
func (i *ICoreWebView2Environment10) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Environment) GetICoreWebView2Environment10() *ICoreWebView2Environment10 {
	var result *ICoreWebView2Environment10

	iidICoreWebView2Environment10 := NewGUID("{ee0eb9df-6f12-46ce-b53f-3f47b9c928e0}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Environment10)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Environment10) CreateCoreWebView2ControllerOptions() (*ICoreWebView2ControllerOptions, error) {

	var value *ICoreWebView2ControllerOptions

	hr, _, _ := i.Vtbl.CreateCoreWebView2ControllerOptions.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2Environment10) CreateCoreWebView2ControllerWithOptions(ParentWindow HWND, options *ICoreWebView2ControllerOptions, handler *ICoreWebView2CreateCoreWebView2ControllerCompletedHandler) error {

	hr, _, _ := i.Vtbl.CreateCoreWebView2ControllerWithOptions.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(ParentWindow),
		uintptr(unsafe.Pointer(options)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Environment10) CreateCoreWebView2CompositionControllerWithOptions(ParentWindow HWND, options *ICoreWebView2ControllerOptions, handler *ICoreWebView2CreateCoreWebView2CompositionControllerCompletedHandler) error {

	hr, _, _ := i.Vtbl.CreateCoreWebView2CompositionControllerWithOptions.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(ParentWindow),
		uintptr(unsafe.Pointer(options)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
