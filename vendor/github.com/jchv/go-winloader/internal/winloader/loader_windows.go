package winloader

import (
	"syscall"

	"github.com/jchv/go-winloader/internal/loader"
	"golang.org/x/sys/windows"
)

// Proc is a Proc implementation using a native procedure.
type Proc uintptr

// Call calls a native procedure.
func (p Proc) Call(a ...uint64) (r1, r2 uint64, lastErr error) {
	var n1, n2 uintptr
	switch len(a) {
	case 0:
		n1, n2, lastErr = syscall.Syscall(uintptr(p), 0, 0, 0, 0)
	case 1:
		n1, n2, lastErr = syscall.Syscall(uintptr(p), 1, uintptr(a[0]), 0, 0)
	case 2:
		n1, n2, lastErr = syscall.Syscall(uintptr(p), 2, uintptr(a[0]), uintptr(a[1]), 0)
	case 3:
		n1, n2, lastErr = syscall.Syscall(uintptr(p), 3, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]))
	case 4:
		n1, n2, lastErr = syscall.Syscall6(uintptr(p), 4, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), 0, 0)
	case 5:
		n1, n2, lastErr = syscall.Syscall6(uintptr(p), 5, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), 0)
	case 6:
		n1, n2, lastErr = syscall.Syscall6(uintptr(p), 6, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]))
	case 7:
		n1, n2, lastErr = syscall.Syscall9(uintptr(p), 7, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), 0, 0)
	case 8:
		n1, n2, lastErr = syscall.Syscall9(uintptr(p), 8, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), 0)
	case 9:
		n1, n2, lastErr = syscall.Syscall9(uintptr(p), 9, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]))
	case 10:
		n1, n2, lastErr = syscall.Syscall12(uintptr(p), 10, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]), uintptr(a[9]), 0, 0)
	case 11:
		n1, n2, lastErr = syscall.Syscall12(uintptr(p), 11, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]), uintptr(a[9]), uintptr(a[10]), 0)
	case 12:
		n1, n2, lastErr = syscall.Syscall12(uintptr(p), 12, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]), uintptr(a[9]), uintptr(a[10]), uintptr(a[11]))
	case 13:
		n1, n2, lastErr = syscall.Syscall15(uintptr(p), 13, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]), uintptr(a[9]), uintptr(a[10]), uintptr(a[11]), uintptr(a[12]), 0, 0)
	case 14:
		n1, n2, lastErr = syscall.Syscall15(uintptr(p), 14, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]), uintptr(a[9]), uintptr(a[10]), uintptr(a[11]), uintptr(a[12]), uintptr(a[13]), 0)
	case 15:
		n1, n2, lastErr = syscall.Syscall15(uintptr(p), 15, uintptr(a[0]), uintptr(a[1]), uintptr(a[2]), uintptr(a[3]), uintptr(a[4]), uintptr(a[5]), uintptr(a[6]), uintptr(a[7]), uintptr(a[8]), uintptr(a[9]), uintptr(a[10]), uintptr(a[11]), uintptr(a[12]), uintptr(a[13]), uintptr(a[14]))
	default:
		panic("too many arguments")
	}
	return uint64(n1), uint64(n2), lastErr
}

// Addr gets the native address of the procedure.
func (p Proc) Addr() uint64 {
	return uint64(p)
}

// Library is a module implementation using the native Windows loader.
type Library windows.Handle

// Proc implements loader.Module
func (l Library) Proc(name string) loader.Proc {
	proc, _ := windows.GetProcAddress(windows.Handle(l), name)
	if proc == 0 {
		return nil
	}
	return Proc(proc)
}

// Ordinal implements loader.Module
func (l Library) Ordinal(ordinal uint64) loader.Proc {
	proc, _ := windows.GetProcAddressByOrdinal(windows.Handle(l), uintptr(ordinal))
	if proc == 0 {
		return nil
	}
	return Proc(proc)
}

// Free implements loader.Module
func (l Library) Free() error {
	return windows.FreeLibrary(windows.Handle(l))
}

// Loader is a loader that uses the native Windows library loader.
type Loader struct{}

// Load loads a module into memory.
func (Loader) Load(libname string) (loader.Module, error) {
	handle, err := windows.LoadLibrary(libname)
	if err != nil {
		return nil, err
	}
	return Library(handle), nil
}
