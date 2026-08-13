//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2ProcessFailedEventArgs3Vtbl struct {
	ICoreWebView2ProcessFailedEventArgs2Vtbl
	GetFailureSourceModulePath ComProc
}

type ICoreWebView2ProcessFailedEventArgs3 struct {
	Vtbl *ICoreWebView2ProcessFailedEventArgs3Vtbl
}

func (i *ICoreWebView2ProcessFailedEventArgs3) AddRef() uintptr {
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
func (i *ICoreWebView2ProcessFailedEventArgs3) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2ProcessFailedEventArgs) GetICoreWebView2ProcessFailedEventArgs3() *ICoreWebView2ProcessFailedEventArgs3 {
	var result *ICoreWebView2ProcessFailedEventArgs3

	iidICoreWebView2ProcessFailedEventArgs3 := NewGUID("{ab667428-094d-5fd1-b480-8b4c0fdbdf2f}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2ProcessFailedEventArgs3)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2ProcessFailedEventArgs3) GetFailureSourceModulePath() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.GetFailureSourceModulePath.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return "", syscall.Errno(hr)
	}
	// Get result and cleanup
	value := UTF16PtrToString(_value)
	CoTaskMemFree(unsafe.Pointer(_value))
	return value, nil
}
