//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2EnvironmentVtbl struct {
	IUnknownVtbl
	CreateCoreWebView2Controller     ComProc
	CreateWebResourceResponse        ComProc
	GetBrowserVersionString          ComProc
	AddNewBrowserVersionAvailable    ComProc
	RemoveNewBrowserVersionAvailable ComProc
}

type ICoreWebView2Environment struct {
	Vtbl *ICoreWebView2EnvironmentVtbl
}

func (i *ICoreWebView2Environment) AddRef() uintptr {
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
func (i *ICoreWebView2Environment) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Environment) CreateCoreWebView2Controller(parentWindow HWND, handler *ICoreWebView2CreateCoreWebView2ControllerCompletedHandler) error {

	hr, _, _ := i.Vtbl.CreateCoreWebView2Controller.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(parentWindow),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Environment) CreateWebResourceResponse(content *IStream, statusCode int32, reasonPhrase string, headers string) (*ICoreWebView2WebResourceResponse, error) {

	// Convert string 'reasonPhrase' to *uint16
	_reasonPhrase, err := UTF16PtrFromString(reasonPhrase)
	if err != nil {
		return nil, err
	}
	// Convert string 'headers' to *uint16
	_headers, err := UTF16PtrFromString(headers)
	if err != nil {
		return nil, err
	}
	var response *ICoreWebView2WebResourceResponse

	hr, _, _ := i.Vtbl.CreateWebResourceResponse.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(content)),
		uintptr(statusCode),
		uintptr(unsafe.Pointer(_reasonPhrase)),
		uintptr(unsafe.Pointer(_headers)),
		uintptr(unsafe.Pointer(&response)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return response, nil
}

func (i *ICoreWebView2Environment) GetBrowserVersionString() (string, error) {
	// Create *uint16 to hold result
	var _versionInfo *uint16

	hr, _, _ := i.Vtbl.GetBrowserVersionString.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_versionInfo)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return "", syscall.Errno(hr)
	}
	// Get result and cleanup
	versionInfo := UTF16PtrToString(_versionInfo)
	CoTaskMemFree(unsafe.Pointer(_versionInfo))
	return versionInfo, nil
}

func (i *ICoreWebView2Environment) AddNewBrowserVersionAvailable(eventHandler *ICoreWebView2NewBrowserVersionAvailableEventHandler) (EventRegistrationToken, error) {

	var token EventRegistrationToken

	hr, _, _ := i.Vtbl.AddNewBrowserVersionAvailable.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return EventRegistrationToken{}, syscall.Errno(hr)
	}
	return token, nil
}

func (i *ICoreWebView2Environment) RemoveNewBrowserVersionAvailable(token EventRegistrationToken) error {

	hr, _, _ := i.Vtbl.RemoveNewBrowserVersionAvailable.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(*(*uint64)(unsafe.Pointer(&token))),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
