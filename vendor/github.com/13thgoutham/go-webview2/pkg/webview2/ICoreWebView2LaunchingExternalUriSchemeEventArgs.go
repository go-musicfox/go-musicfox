//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2LaunchingExternalUriSchemeEventArgsVtbl struct {
	IUnknownVtbl
	GetUri              ComProc
	GetInitiatingOrigin ComProc
	GetIsUserInitiated  ComProc
	GetCancel           ComProc
	PutCancel           ComProc
	GetDeferral         ComProc
}

type ICoreWebView2LaunchingExternalUriSchemeEventArgs struct {
	Vtbl *ICoreWebView2LaunchingExternalUriSchemeEventArgsVtbl
}

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) AddRef() uintptr {
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
func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) GetUri() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.GetUri.Call(
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

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) GetInitiatingOrigin() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.GetInitiatingOrigin.Call(
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

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) GetIsUserInitiated() (bool, error) {
	// Create int32 to hold bool result
	var _value int32

	hr, _, _ := i.Vtbl.GetIsUserInitiated.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return false, syscall.Errno(hr)
	}
	// Get result and cleanup
	value := _value != 0
	return value, nil
}

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) GetCancel() (bool, error) {
	// Create int32 to hold bool result
	var _value int32

	hr, _, _ := i.Vtbl.GetCancel.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return false, syscall.Errno(hr)
	}
	// Get result and cleanup
	value := _value != 0
	return value, nil
}

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) PutCancel(value bool) error {

	hr, _, _ := i.Vtbl.PutCancel.Call(
		uintptr(unsafe.Pointer(i)),
		boolToUintptr(value),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2LaunchingExternalUriSchemeEventArgs) GetDeferral() (*ICoreWebView2Deferral, error) {

	var value *ICoreWebView2Deferral

	hr, _, _ := i.Vtbl.GetDeferral.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}
