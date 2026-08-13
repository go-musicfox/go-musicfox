//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2WebMessageReceivedEventArgsVtbl struct {
	IUnknownVtbl
	GetSource                ComProc
	GetWebMessageAsJson      ComProc
	TryGetWebMessageAsString ComProc
}

type ICoreWebView2WebMessageReceivedEventArgs struct {
	Vtbl *ICoreWebView2WebMessageReceivedEventArgsVtbl
}

func (i *ICoreWebView2WebMessageReceivedEventArgs) AddRef() uintptr {
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
func (i *ICoreWebView2WebMessageReceivedEventArgs) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2WebMessageReceivedEventArgs) GetSource() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.GetSource.Call(
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

func (i *ICoreWebView2WebMessageReceivedEventArgs) GetWebMessageAsJson() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.GetWebMessageAsJson.Call(
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

func (i *ICoreWebView2WebMessageReceivedEventArgs) TryGetWebMessageAsString() (string, error) {
	// Create *uint16 to hold result
	var _value *uint16

	hr, _, _ := i.Vtbl.TryGetWebMessageAsString.Call(
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
