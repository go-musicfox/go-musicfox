//go:build !linux && !freebsd && !netbsd && !openbsd && !windows && !darwin && !illumos && !js
// +build !linux,!freebsd,!netbsd,!openbsd,!windows,!darwin,!illumos,!js

package beeep

// Alert displays a desktop notification and plays a beep.
func Alert(title, message, appIcon string) error {
	return ErrUnsupported
}
