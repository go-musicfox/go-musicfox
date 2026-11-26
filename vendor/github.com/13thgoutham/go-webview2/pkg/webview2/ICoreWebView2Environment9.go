//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Environment9Vtbl struct {
	ICoreWebView2Environment8Vtbl
	CreateContextMenuItem ComProc
}

type ICoreWebView2Environment9 struct {
	Vtbl *ICoreWebView2Environment9Vtbl
}

func (i *ICoreWebView2Environment9) AddRef() uintptr {
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
func (i *ICoreWebView2Environment9) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Environment) GetICoreWebView2Environment9() *ICoreWebView2Environment9 {
	var result *ICoreWebView2Environment9

	iidICoreWebView2Environment9 := NewGUID("{f06f41bf-4b5a-49d8-b9f6-fa16cd29f274}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Environment9)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Environment9) CreateContextMenuItem(Label string, iconStream *IStream, Kind COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND) (*ICoreWebView2ContextMenuItem, error) {

	// Convert string 'Label' to *uint16
	_Label, err := UTF16PtrFromString(Label)
	if err != nil {
		return nil, err
	}
	var value *ICoreWebView2ContextMenuItem

	hr, _, _ := i.Vtbl.CreateContextMenuItem.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_Label)),
		uintptr(unsafe.Pointer(iconStream)),
		uintptr(Kind),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}
