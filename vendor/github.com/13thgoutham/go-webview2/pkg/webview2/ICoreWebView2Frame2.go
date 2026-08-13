//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Frame2Vtbl struct {
	ICoreWebView2FrameVtbl
	AddNavigationStarting     ComProc
	RemoveNavigationStarting  ComProc
	AddContentLoading         ComProc
	RemoveContentLoading      ComProc
	AddNavigationCompleted    ComProc
	RemoveNavigationCompleted ComProc
	AddDOMContentLoaded       ComProc
	RemoveDOMContentLoaded    ComProc
	ExecuteScript             ComProc
	PostWebMessageAsJson      ComProc
	PostWebMessageAsString    ComProc
	AddWebMessageReceived     ComProc
	RemoveWebMessageReceived  ComProc
}

type ICoreWebView2Frame2 struct {
	Vtbl *ICoreWebView2Frame2Vtbl
}

func (i *ICoreWebView2Frame2) AddRef() uintptr {
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
func (i *ICoreWebView2Frame2) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Frame) GetICoreWebView2Frame2() *ICoreWebView2Frame2 {
	var result *ICoreWebView2Frame2

	iidICoreWebView2Frame2 := NewGUID("{7a6a5834-d185-4dbf-b63f-4a9bc43107d4}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Frame2)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Frame2) AddNavigationStarting(eventHandler *ICoreWebView2FrameNavigationStartingEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddNavigationStarting.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2Frame2) RemoveNavigationStarting(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveNavigationStarting.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) AddContentLoading(eventHandler *ICoreWebView2FrameContentLoadingEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddContentLoading.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2Frame2) RemoveContentLoading(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveContentLoading.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) AddNavigationCompleted(eventHandler *ICoreWebView2FrameNavigationCompletedEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddNavigationCompleted.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2Frame2) RemoveNavigationCompleted(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveNavigationCompleted.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) AddDOMContentLoaded(eventHandler *ICoreWebView2FrameDOMContentLoadedEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddDOMContentLoaded.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2Frame2) RemoveDOMContentLoaded(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveDOMContentLoaded.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) ExecuteScript(javaScript string, handler *ICoreWebView2ExecuteScriptCompletedHandler) error {

	// Convert string 'javaScript' to *uint16
	_javaScript, err := UTF16PtrFromString(javaScript)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.ExecuteScript.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_javaScript)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) PostWebMessageAsJson(webMessageAsJson string) error {

	// Convert string 'webMessageAsJson' to *uint16
	_webMessageAsJson, err := UTF16PtrFromString(webMessageAsJson)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.PostWebMessageAsJson.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_webMessageAsJson)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) PostWebMessageAsString(webMessageAsString string) error {

	// Convert string 'webMessageAsString' to *uint16
	_webMessageAsString, err := UTF16PtrFromString(webMessageAsString)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.PostWebMessageAsString.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_webMessageAsString)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Frame2) AddWebMessageReceived(handler *ICoreWebView2FrameWebMessageReceivedEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddWebMessageReceived.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2Frame2) RemoveWebMessageReceived(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveWebMessageReceived.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
