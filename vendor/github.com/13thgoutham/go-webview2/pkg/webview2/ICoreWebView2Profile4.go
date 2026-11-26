//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Profile4Vtbl struct {
	ICoreWebView2Profile3Vtbl
	SetPermissionState              ComProc
	GetNonDefaultPermissionSettings ComProc
}

type ICoreWebView2Profile4 struct {
	Vtbl *ICoreWebView2Profile4Vtbl
}

func (i *ICoreWebView2Profile4) AddRef() uintptr {
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
func (i *ICoreWebView2Profile4) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Profile) GetICoreWebView2Profile4() *ICoreWebView2Profile4 {
	var result *ICoreWebView2Profile4

	iidICoreWebView2Profile4 := NewGUID("{8f4ae680-192e-4ec8-833a-21cfadaef628}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Profile4)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Profile4) SetPermissionState(PermissionKind COREWEBVIEW2_PERMISSION_KIND, origin string, State COREWEBVIEW2_PERMISSION_STATE, handler *ICoreWebView2SetPermissionStateCompletedHandler) error {

	// Convert string 'origin' to *uint16
	_origin, err := UTF16PtrFromString(origin)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.SetPermissionState.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(PermissionKind),
		uintptr(unsafe.Pointer(_origin)),
		uintptr(State),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Profile4) GetNonDefaultPermissionSettings(handler *ICoreWebView2GetNonDefaultPermissionSettingsCompletedHandler) error {

	hr, _, _ := i.Vtbl.GetNonDefaultPermissionSettings.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
