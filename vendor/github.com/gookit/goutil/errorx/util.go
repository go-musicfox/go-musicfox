package errorx

import (
	"errors"
	"fmt"
)

// Raw new a raw go error. alias of errors.New()
func Raw(msg string) error {
	return errors.New(msg)
}

// Rawf new a raw go error. alias of errors.New()
func Rawf(tpl string, vars ...any) error {
	return fmt.Errorf(tpl, vars...)
}

/*************************************************************
 * helper func for error
 *************************************************************/

// Cause returns the first cause error by call err.Cause().
// Otherwise, will returns current error.
func Cause(err error) error {
	if err == nil {
		return nil
	}

	if err, ok := err.(Causer); ok {
		return err.Cause()
	}
	return err
}

// Unwrap returns previous error by call err.Unwrap().
// Otherwise, will returns nil.
func Unwrap(err error) error {
	if err == nil {
		return nil
	}

	if err, ok := err.(Unwrapper); ok {
		return err.Unwrap()
	}
	return nil
}

// Previous alias of Unwrap()
func Previous(err error) error { return Unwrap(err) }

// ToErrorX convert check
func ToErrorX(err error) (ex *ErrorX, ok bool) {
	ex, ok = err.(*ErrorX)
	return
}

// Has check err has contains target, or err is eq target.
// alias of errors.Is()
func Has(err, target error) bool {
	return errors.Is(err, target)
}

// Is alias of errors.Is()
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// To try convert err to target, returns is result.
//
// NOTICE: target must be ptr and not nil. alias of errors.As()
//
// Usage:
//
//	var ex *errorx.ErrorX
//	err := doSomething()
//	if errorx.To(err, &ex) {
//		fmt.Println(ex.GoString())
//	}
func To(err error, target any) bool {
	return errors.As(err, target)
}

// As same of the To(), alias of errors.As()
//
// NOTICE: target must be ptr and not nil
func As(err error, target any) bool {
	return errors.As(err, target)
}
