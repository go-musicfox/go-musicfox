// Package stdutil provide some standard util functions for go.
package stdutil

import (
	"fmt"
)

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
func Panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}
