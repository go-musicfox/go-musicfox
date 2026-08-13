//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Settings3Vtbl struct {
	ICoreWebView2Settings2Vtbl
	GetAreBrowserAcceleratorKeysEnabled ComProc
	PutAreBrowserAcceleratorKeysEnabled ComProc
}

type ICoreWebView2Settings3 struct {
	Vtbl *ICoreWebView2Settings3Vtbl
}

func (i *ICoreWebView2Settings3) AddRef() uintptr {
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
func (i *ICoreWebView2Settings3) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Settings) GetICoreWebView2Settings3() *ICoreWebView2Settings3 {
	var result *ICoreWebView2Settings3

	iidICoreWebView2Settings3 := NewGUID("{fdb5ab74-af33-4854-84f0-0a631deb5eba}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Settings3)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Settings3) GetAreBrowserAcceleratorKeysEnabled() (bool, error) {
	// Create int32 to hold bool result
	var _value int32

	hr, _, _ := i.Vtbl.GetAreBrowserAcceleratorKeysEnabled.Call(
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

func (i *ICoreWebView2Settings3) PutAreBrowserAcceleratorKeysEnabled(value bool) error {

	hr, _, _ := i.Vtbl.PutAreBrowserAcceleratorKeysEnabled.Call(
		uintptr(unsafe.Pointer(i)),
		boolToUintptr(value),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
