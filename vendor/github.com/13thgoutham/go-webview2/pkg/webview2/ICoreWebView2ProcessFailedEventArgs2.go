//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2ProcessFailedEventArgs2Vtbl struct {
	ICoreWebView2ProcessFailedEventArgsVtbl
	GetReason                     ComProc
	GetExitCode                   ComProc
	GetProcessDescription         ComProc
	GetFrameInfosForFailedProcess ComProc
}

type ICoreWebView2ProcessFailedEventArgs2 struct {
	Vtbl *ICoreWebView2ProcessFailedEventArgs2Vtbl
}

func (i *ICoreWebView2ProcessFailedEventArgs2) AddRef() uintptr {
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
func (i *ICoreWebView2ProcessFailedEventArgs2) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2ProcessFailedEventArgs) GetICoreWebView2ProcessFailedEventArgs2() *ICoreWebView2ProcessFailedEventArgs2 {
	var result *ICoreWebView2ProcessFailedEventArgs2

	iidICoreWebView2ProcessFailedEventArgs2 := NewGUID("{4dab9422-46fa-4c3e-a5d2-41d2071d3680}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2ProcessFailedEventArgs2)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2ProcessFailedEventArgs2) GetReason() (COREWEBVIEW2_PROCESS_FAILED_REASON, error) {

	var reason COREWEBVIEW2_PROCESS_FAILED_REASON

	hr, _, _ := i.Vtbl.GetReason.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&reason)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return reason, nil
}

func (i *ICoreWebView2ProcessFailedEventArgs2) GetExitCode() (int32, error) {

	var exitCode int32

	hr, _, _ := i.Vtbl.GetExitCode.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&exitCode)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return exitCode, nil
}

func (i *ICoreWebView2ProcessFailedEventArgs2) GetProcessDescription() (string, error) {
	// Create *uint16 to hold result
	var _processDescription *uint16

	hr, _, _ := i.Vtbl.GetProcessDescription.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_processDescription)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return "", syscall.Errno(hr)
	}
	// Get result and cleanup
	processDescription := UTF16PtrToString(_processDescription)
	CoTaskMemFree(unsafe.Pointer(_processDescription))
	return processDescription, nil
}

func (i *ICoreWebView2ProcessFailedEventArgs2) GetFrameInfosForFailedProcess() (*ICoreWebView2FrameInfoCollection, error) {

	var frames *ICoreWebView2FrameInfoCollection

	hr, _, _ := i.Vtbl.GetFrameInfosForFailedProcess.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&frames)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return frames, nil
}
