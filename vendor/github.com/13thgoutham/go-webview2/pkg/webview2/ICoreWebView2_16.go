//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2_16Vtbl struct {
	ICoreWebView2_15Vtbl
	Print            ComProc
	ShowPrintUI      ComProc
	PrintToPdfStream ComProc
}

type ICoreWebView2_16 struct {
	Vtbl *ICoreWebView2_16Vtbl
}

func (i *ICoreWebView2_16) AddRef() uintptr {
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
func (i *ICoreWebView2_16) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2) GetICoreWebView2_16() *ICoreWebView2_16 {
	var result *ICoreWebView2_16

	iidICoreWebView2_16 := NewGUID("{0EB34DC9-9F91-41E1-8639-95CD5943906B}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_16)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2_16) Print(printSettings *ICoreWebView2PrintSettings, handler *ICoreWebView2PrintCompletedHandler) error {

	hr, _, _ := i.Vtbl.Print.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(printSettings)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2_16) ShowPrintUI(printDialogKind COREWEBVIEW2_PRINT_DIALOG_KIND) error {

	hr, _, _ := i.Vtbl.ShowPrintUI.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(printDialogKind),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2_16) PrintToPdfStream(printSettings *ICoreWebView2PrintSettings, handler *ICoreWebView2PrintToPdfStreamCompletedHandler) error {

	hr, _, _ := i.Vtbl.PrintToPdfStream.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(printSettings)),
		uintptr(unsafe.Pointer(handler)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
