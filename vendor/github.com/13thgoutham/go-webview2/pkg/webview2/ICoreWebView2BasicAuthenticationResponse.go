//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2BasicAuthenticationResponseVtbl struct {
	IUnknownVtbl
	GetUserName ComProc
	PutUserName ComProc
	GetPassword ComProc
	PutPassword ComProc
}

type ICoreWebView2BasicAuthenticationResponse struct {
	Vtbl *ICoreWebView2BasicAuthenticationResponseVtbl
}

func (i *ICoreWebView2BasicAuthenticationResponse) AddRef() uintptr {
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
func (i *ICoreWebView2BasicAuthenticationResponse) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2BasicAuthenticationResponse) GetUserName() (string, error) {
	// Create *uint16 to hold result
	var _userName *uint16

	hr, _, _ := i.Vtbl.GetUserName.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_userName)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return "", syscall.Errno(hr)
	}
	// Get result and cleanup
	userName := UTF16PtrToString(_userName)
	CoTaskMemFree(unsafe.Pointer(_userName))
	return userName, nil
}

func (i *ICoreWebView2BasicAuthenticationResponse) PutUserName(userName string) error {

	// Convert string 'userName' to *uint16
	_userName, err := UTF16PtrFromString(userName)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.PutUserName.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_userName)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2BasicAuthenticationResponse) GetPassword() (string, error) {
	// Create *uint16 to hold result
	var _password *uint16

	hr, _, _ := i.Vtbl.GetPassword.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_password)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return "", syscall.Errno(hr)
	}
	// Get result and cleanup
	password := UTF16PtrToString(_password)
	CoTaskMemFree(unsafe.Pointer(_password))
	return password, nil
}

func (i *ICoreWebView2BasicAuthenticationResponse) PutPassword(password string) error {

	// Convert string 'password' to *uint16
	_password, err := UTF16PtrFromString(password)
	if err != nil {
		return err
	}

	hr, _, _ := i.Vtbl.PutPassword.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_password)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return syscall.Errno(hr)
	}
	return nil
}
