// Package goutil 💪 Useful utils for Go: int, string, array/slice, map, error, time, format, CLI, ENV, filesystem,
// system, testing, debug and more.
package goutil

import (
	"fmt"

	"github.com/gookit/goutil/stdutil"
)

// Value alias of stdutil.Value
type Value = stdutil.Value

// Go is a basic promise implementation: it wraps calls a function in a goroutine
// and returns a channel which will later return the function's return value.
func Go(f func() error) error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()
	return <-ch
}

// PanicIfErr if error is not empty, will panic
func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// PanicErr if error is not empty, will panic
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

// MustOK if error is not empty, will panic
func MustOK(err error) {
	if err != nil {
		panic(err)
	}
}

// Panicf format panic message use fmt.Sprintf
func Panicf(format string, v ...any) {
	panic(fmt.Sprintf(format, v...))
}

// FuncName get func name
func FuncName(f any) string {
	return stdutil.FuncName(f)
}

// PkgName get current package name. alias of stdutil.PkgName()
//
// Usage:
//
//	funcName := goutil.FuncName(fn)
//	pgkName := goutil.PkgName(funcName)
func PkgName(funcName string) string {
	return stdutil.PkgName(funcName)
}
