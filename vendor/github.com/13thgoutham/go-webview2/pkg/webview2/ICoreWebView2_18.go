//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2_18Vtbl struct {
	ICoreWebView2_17Vtbl
	AddLaunchingExternalUriScheme    ComProc
	RemoveLaunchingExternalUriScheme ComProc
}

type ICoreWebView2_18 struct {
	Vtbl *ICoreWebView2_18Vtbl
}

func (i *ICoreWebView2_18) AddRef() uintptr {
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
func (i *ICoreWebView2_18) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2) GetICoreWebView2_18() *ICoreWebView2_18 {
	var result *ICoreWebView2_18

	iidICoreWebView2_18 := NewGUID("{7a626017-28be-49b2-b865-3ba2b3522d90}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_18)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2_18) AddLaunchingExternalUriScheme(eventHandler *ICoreWebView2LaunchingExternalUriSchemeEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddLaunchingExternalUriScheme.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2_18) RemoveLaunchingExternalUriScheme(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveLaunchingExternalUriScheme.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
