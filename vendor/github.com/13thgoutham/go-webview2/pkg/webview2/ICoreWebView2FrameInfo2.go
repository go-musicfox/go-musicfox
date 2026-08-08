//go:build windows

package webview2

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type ICoreWebView2FrameInfo2Vtbl struct {
	ICoreWebView2FrameInfoVtbl
	GetParentFrameInfo ComProc
	GetFrameId         ComProc
	GetFrameKind       ComProc
}

type ICoreWebView2FrameInfo2 struct {
	Vtbl *ICoreWebView2FrameInfo2Vtbl
}

func (i *ICoreWebView2FrameInfo2) AddRef() uintptr {
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
func (i *ICoreWebView2FrameInfo2) Release() uint32 {
	return i.Vtbl.CallRelease(unsafe.Pointer(i))
}

func (i *ICoreWebView2FrameInfo) GetICoreWebView2FrameInfo2() *ICoreWebView2FrameInfo2 {
	var result *ICoreWebView2FrameInfo2

	iidICoreWebView2FrameInfo2 := NewGUID("{56f85cfa-72c4-11ee-b962-0242ac120002}")
	// The HRESULT is deliberately not returned, because changing the signature of all 82 of these
	// accessors is an API break. It is E_NOINTERFACE whenever the installed WebView2 Runtime is
	// older than this interface, which is the normal case rather than an exotic one -- and then
	// result stays nil and the CALLER's next method call dereferences it. Callers must nil-check.
	//
	// This also leaks a reference on success: QueryInterface AddRefs, and no Release is generated.
	// Use Vtbl.CallRelease(unsafe.Pointer(x)) via the embedded IUnknownVtbl when finished.
	_, _, _ = i.Vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2FrameInfo2)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (i *ICoreWebView2FrameInfo2) GetParentFrameInfo() (*ICoreWebView2FrameInfo, error) {

	var frameInfo *ICoreWebView2FrameInfo

	hr, _, _ := i.Vtbl.GetParentFrameInfo.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&frameInfo)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return nil, syscall.Errno(hr)
	}
	return frameInfo, nil
}

func (i *ICoreWebView2FrameInfo2) GetFrameId() (uint32, error) {

	var id uint32

	hr, _, _ := i.Vtbl.GetFrameId.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&id)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return id, nil
}

func (i *ICoreWebView2FrameInfo2) GetFrameKind() (COREWEBVIEW2_FRAME_KIND, error) {

	var kind COREWEBVIEW2_FRAME_KIND

	hr, _, _ := i.Vtbl.GetFrameKind.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&kind)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, syscall.Errno(hr)
	}
	return kind, nil
}
