//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2Environment14Vtbl struct {
	ICoreWebView2Environment13Vtbl
	CreateWebFileSystemFileHandle      ComProc
	CreateWebFileSystemDirectoryHandle ComProc
	CreateObjectCollection             ComProc
}

type ICoreWebView2Environment14 struct {
	Vtbl *ICoreWebView2Environment14Vtbl
}

func (i *ICoreWebView2Environment14) AddRef() uintptr {
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
func (i *ICoreWebView2Environment14) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2Environment) GetICoreWebView2Environment14() *ICoreWebView2Environment14 {
	var result *ICoreWebView2Environment14

	iidICoreWebView2Environment14 := NewGUID("{a5e9fad9-c875-59da-9bd7-473aa5ca1cef}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2Environment14)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2Environment14) CreateWebFileSystemFileHandle(path string, permission COREWEBVIEW2_FILE_SYSTEM_HANDLE_PERMISSION) (*ICoreWebView2FileSystemHandle, error) {

	// Convert string 'path' to *uint16
	_path, err := UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	var value *ICoreWebView2FileSystemHandle

	hr, _, _ := i.Vtbl.CreateWebFileSystemFileHandle.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_path)),
		uintptr(permission),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2Environment14) CreateWebFileSystemDirectoryHandle(path string, permission COREWEBVIEW2_FILE_SYSTEM_HANDLE_PERMISSION) (*ICoreWebView2FileSystemHandle, error) {

	// Convert string 'path' to *uint16
	_path, err := UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	var value *ICoreWebView2FileSystemHandle

	hr, _, _ := i.Vtbl.CreateWebFileSystemDirectoryHandle.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_path)),
		uintptr(permission),
		uintptr(unsafe.Pointer(&value)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return value, nil
}

func (i *ICoreWebView2Environment14) CreateObjectCollection(length uint32, items **IUnknown) (*ICoreWebView2ObjectCollection, error) {

	var objectCollection *ICoreWebView2ObjectCollection

	hr, _, _ := i.Vtbl.CreateObjectCollection.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(length),
		uintptr(unsafe.Pointer(items)),
		uintptr(unsafe.Pointer(&objectCollection)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return objectCollection, nil
}
