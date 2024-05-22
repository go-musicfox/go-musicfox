// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint:all
package adaptive

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/foundation"
)

const SignatureAdaptiveMediaSourceAdvancedSettings string = "rc(Windows.Media.Streaming.Adaptive.AdaptiveMediaSourceAdvancedSettings;{55db1680-1aeb-47dc-aa08-9a11610ba45a})"

type AdaptiveMediaSourceAdvancedSettings struct {
	ole.IUnknown
}

func (impl *AdaptiveMediaSourceAdvancedSettings) GetAllSegmentsIndependent() (bool, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiAdaptiveMediaSourceAdvancedSettings))
	defer itf.Release()
	v := (*iAdaptiveMediaSourceAdvancedSettings)(unsafe.Pointer(itf))
	return v.GetAllSegmentsIndependent()
}

func (impl *AdaptiveMediaSourceAdvancedSettings) SetAllSegmentsIndependent(value bool) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiAdaptiveMediaSourceAdvancedSettings))
	defer itf.Release()
	v := (*iAdaptiveMediaSourceAdvancedSettings)(unsafe.Pointer(itf))
	return v.SetAllSegmentsIndependent(value)
}

func (impl *AdaptiveMediaSourceAdvancedSettings) GetDesiredBitrateHeadroomRatio() (*foundation.IReference, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiAdaptiveMediaSourceAdvancedSettings))
	defer itf.Release()
	v := (*iAdaptiveMediaSourceAdvancedSettings)(unsafe.Pointer(itf))
	return v.GetDesiredBitrateHeadroomRatio()
}

func (impl *AdaptiveMediaSourceAdvancedSettings) SetDesiredBitrateHeadroomRatio(value *foundation.IReference) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiAdaptiveMediaSourceAdvancedSettings))
	defer itf.Release()
	v := (*iAdaptiveMediaSourceAdvancedSettings)(unsafe.Pointer(itf))
	return v.SetDesiredBitrateHeadroomRatio(value)
}

func (impl *AdaptiveMediaSourceAdvancedSettings) GetBitrateDowngradeTriggerRatio() (*foundation.IReference, error) {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiAdaptiveMediaSourceAdvancedSettings))
	defer itf.Release()
	v := (*iAdaptiveMediaSourceAdvancedSettings)(unsafe.Pointer(itf))
	return v.GetBitrateDowngradeTriggerRatio()
}

func (impl *AdaptiveMediaSourceAdvancedSettings) SetBitrateDowngradeTriggerRatio(value *foundation.IReference) error {
	itf := impl.MustQueryInterface(ole.NewGUID(GUIDiAdaptiveMediaSourceAdvancedSettings))
	defer itf.Release()
	v := (*iAdaptiveMediaSourceAdvancedSettings)(unsafe.Pointer(itf))
	return v.SetBitrateDowngradeTriggerRatio(value)
}

const GUIDiAdaptiveMediaSourceAdvancedSettings string = "55db1680-1aeb-47dc-aa08-9a11610ba45a"
const SignatureiAdaptiveMediaSourceAdvancedSettings string = "{55db1680-1aeb-47dc-aa08-9a11610ba45a}"

type iAdaptiveMediaSourceAdvancedSettings struct {
	ole.IInspectable
}

type iAdaptiveMediaSourceAdvancedSettingsVtbl struct {
	ole.IInspectableVtbl

	GetAllSegmentsIndependent       uintptr
	SetAllSegmentsIndependent       uintptr
	GetDesiredBitrateHeadroomRatio  uintptr
	SetDesiredBitrateHeadroomRatio  uintptr
	GetBitrateDowngradeTriggerRatio uintptr
	SetBitrateDowngradeTriggerRatio uintptr
}

func (v *iAdaptiveMediaSourceAdvancedSettings) VTable() *iAdaptiveMediaSourceAdvancedSettingsVtbl {
	return (*iAdaptiveMediaSourceAdvancedSettingsVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *iAdaptiveMediaSourceAdvancedSettings) GetAllSegmentsIndependent() (bool, error) {
	var out bool
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetAllSegmentsIndependent,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out bool
	)

	if hr != 0 {
		return false, ole.NewError(hr)
	}

	return out, nil
}

func (v *iAdaptiveMediaSourceAdvancedSettings) SetAllSegmentsIndependent(value bool) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetAllSegmentsIndependent,
		uintptr(unsafe.Pointer(v)),                // this
		uintptr(*(*byte)(unsafe.Pointer(&value))), // in bool
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iAdaptiveMediaSourceAdvancedSettings) GetDesiredBitrateHeadroomRatio() (*foundation.IReference, error) {
	var out *foundation.IReference
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetDesiredBitrateHeadroomRatio,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out foundation.IReference
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}

func (v *iAdaptiveMediaSourceAdvancedSettings) SetDesiredBitrateHeadroomRatio(value *foundation.IReference) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetDesiredBitrateHeadroomRatio,
		uintptr(unsafe.Pointer(v)),     // this
		uintptr(unsafe.Pointer(value)), // in foundation.IReference
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}

func (v *iAdaptiveMediaSourceAdvancedSettings) GetBitrateDowngradeTriggerRatio() (*foundation.IReference, error) {
	var out *foundation.IReference
	hr, _, _ := syscall.SyscallN(
		v.VTable().GetBitrateDowngradeTriggerRatio,
		uintptr(unsafe.Pointer(v)),    // this
		uintptr(unsafe.Pointer(&out)), // out foundation.IReference
	)

	if hr != 0 {
		return nil, ole.NewError(hr)
	}

	return out, nil
}

func (v *iAdaptiveMediaSourceAdvancedSettings) SetBitrateDowngradeTriggerRatio(value *foundation.IReference) error {
	hr, _, _ := syscall.SyscallN(
		v.VTable().SetBitrateDowngradeTriggerRatio,
		uintptr(unsafe.Pointer(v)),     // this
		uintptr(unsafe.Pointer(value)), // in foundation.IReference
	)

	if hr != 0 {
		return ole.NewError(hr)
	}

	return nil
}