//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Environment2Vtbl struct {
	ICoreWebView2EnvironmentVtbl
	CreateWebResourceRequest ComProc
}

type ICoreWebView2Environment2 struct {
	Vtbl *ICoreWebView2Environment2Vtbl
}

func (i *ICoreWebView2Environment2) AddRef() uintptr {
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
func (i *ICoreWebView2Environment2) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Environment) GetICoreWebView2Environment2() *ICoreWebView2Environment2 {
	var result *ICoreWebView2Environment2

	iidICoreWebView2Environment2 := NewGUID("{41f3632b-5ef4-404f-ad82-2d606c5a9a21}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Environment2)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Environment2) CreateWebResourceRequest(uri string, Method string, postData *IStream, Headers string) (*ICoreWebView2WebResourceRequest, error) {

	// Convert string 'uri' to *uint16
	_uri, err := UTF16PtrFromString(uri)
	if err != nil {
		return nil, err
	}
	// Convert string 'Method' to *uint16
	_Method, err := UTF16PtrFromString(Method)
	if err != nil {
		return nil, err
	}
	// Convert string 'Headers' to *uint16
	_Headers, err := UTF16PtrFromString(Headers)
	if err != nil {
		return nil, err
	}
	var value *ICoreWebView2WebResourceRequest

	hr, _, _ := i.Vtbl.CreateWebResourceRequest.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_uri)),
		uintptr(unsafe.Pointer(_Method)),
		uintptr(unsafe.Pointer(postData)),
		uintptr(unsafe.Pointer(_Headers)),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}
