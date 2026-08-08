//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2ServerCertificateErrorDetectedEventArgsVtbl struct {
	IUnknownVtbl
	GetErrorStatus       ComProc
	GetRequestUri        ComProc
	GetServerCertificate ComProc
	GetAction            ComProc
	PutAction            ComProc
	GetDeferral          ComProc
}

type ICoreWebView2ServerCertificateErrorDetectedEventArgs struct {
	Vtbl *ICoreWebView2ServerCertificateErrorDetectedEventArgsVtbl
}

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) AddRef() uintptr {
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
func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) GetErrorStatus() (COREWEBVIEW2_WEB_ERROR_STATUS, error) {

	var value COREWEBVIEW2_WEB_ERROR_STATUS

	hr, _, _ := i.Vtbl.GetErrorStatus.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) GetRequestUri() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.GetRequestUri.Call(
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

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) GetServerCertificate() (*ICoreWebView2Certificate, error) {

	var value *ICoreWebView2Certificate

	hr, _, _ := i.Vtbl.GetServerCertificate.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) GetAction() (COREWEBVIEW2_SERVER_CERTIFICATE_ERROR_ACTION, error) {

	var value COREWEBVIEW2_SERVER_CERTIFICATE_ERROR_ACTION

	hr, _, _ := i.Vtbl.GetAction.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) PutAction(value COREWEBVIEW2_SERVER_CERTIFICATE_ERROR_ACTION) error {

	hr, _, _ := i.Vtbl.PutAction.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(value),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2ServerCertificateErrorDetectedEventArgs) GetDeferral() (*ICoreWebView2Deferral, error) {

	var deferral *ICoreWebView2Deferral

	hr, _, _ := i.Vtbl.GetDeferral.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&deferral)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return deferral, nil
}
