// Package stdutil provide some standard util functions for go.
package stdutil

import (
	"fmt"
	"runtime"
)

// DiscardE discard error
func DiscardE(_ error) {}

// PanicIfErr if error is not empty
func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// PanicIf if error is not empty
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// Panicf format panic message use fmt.Sprintf
func Panicf(format string, v ...any) {
	panic(fmt.Sprintf(format, v...))
}

// GoVersion get go runtime version. eg: "1.18.2"
func GoVersion() string {
	return runtime.Version()[2:]
}
