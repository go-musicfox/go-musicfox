// +build !windows

package winloader

import "errors"

// MakePEBEntryForModule is a hack that inserts an entry for the loaded module
// into the PEB loader data to make it appear to Windows functions.
func MakePEBEntryForModule() error {
	return errors.New("platform not supported")
}

// GetProcessHInstance gets the HINSTANCE for the current process.
func GetProcessHInstance() (uintptr, error) {
	return errors.New("platform not supported")
}
