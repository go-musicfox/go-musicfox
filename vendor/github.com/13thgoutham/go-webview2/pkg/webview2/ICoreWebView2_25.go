//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2_25Vtbl struct {
	ICoreWebView2_24Vtbl
	AddSaveAsUIShowing    ComProc
	RemoveSaveAsUIShowing ComProc
	ShowSaveAsUI          ComProc
}

type ICoreWebView2_25 struct {
	Vtbl *ICoreWebView2_25Vtbl
}

func (i *ICoreWebView2_25) AddRef() uintptr {
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
func (i *ICoreWebView2_25) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2) GetICoreWebView2_25() *ICoreWebView2_25 {
	var result *ICoreWebView2_25

	iidICoreWebView2_25 := NewGUID("{b5a86092-df50-5b4f-a17b-6c8f8b40b771}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_25)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2_25) AddSaveAsUIShowing(eventHandler *ICoreWebView2SaveAsUIShowingEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddSaveAsUIShowing.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2_25) RemoveSaveAsUIShowing(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveSaveAsUIShowing.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2_25) ShowSaveAsUI(handler *ICoreWebView2ShowSaveAsUICompletedHandler) error {

	hr, _, _ := i.Vtbl.ShowSaveAsUI.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
