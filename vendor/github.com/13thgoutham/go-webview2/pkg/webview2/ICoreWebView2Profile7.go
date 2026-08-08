//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Profile7Vtbl struct {
	ICoreWebView2Profile6Vtbl
	AddBrowserExtension  ComProc
	GetBrowserExtensions ComProc
}

type ICoreWebView2Profile7 struct {
	Vtbl *ICoreWebView2Profile7Vtbl
}

func (i *ICoreWebView2Profile7) AddRef() uintptr {
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
func (i *ICoreWebView2Profile7) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Profile) GetICoreWebView2Profile7() *ICoreWebView2Profile7 {
	var result *ICoreWebView2Profile7

	iidICoreWebView2Profile7 := NewGUID("{7b4c7906-a1aa-4cb4-b723-db09f813d541}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Profile7)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Profile7) AddBrowserExtension(extensionFolderPath string, handler *ICoreWebView2ProfileAddBrowserExtensionCompletedHandler) error {

	// Convert string 'extensionFolderPath' to *uint16
	_extensionFolderPath, err := UTF16PtrFromString(extensionFolderPath)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.AddBrowserExtension.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_extensionFolderPath)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Profile7) GetBrowserExtensions(handler *ICoreWebView2ProfileGetBrowserExtensionsCompletedHandler) error {

	hr, _, _ := i.Vtbl.GetBrowserExtensions.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
