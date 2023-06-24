// Package comdef provide some common type or constant definitions
package comdef

import (
	"fmt"
	"io"
)

// ByteStringWriter interface
type ByteStringWriter interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
	fmt.Stringer
}

// StringWriteStringer interface
type StringWriteStringer interface {
	io.StringWriter
	fmt.Stringer
}

// StringMatcher interface
type StringMatcher interface {
	Match(s string) bool
}

// StringMatchFunc definition
type StringMatchFunc func(s string) bool

// Match satisfies the StringMatcher interface
func (fn StringMatchFunc) Match(s string) bool {
	return fn(s)
}

type (
	// MarshalFunc define
	MarshalFunc func(v any) ([]byte, error)

	// UnmarshalFunc define
	UnmarshalFunc func(bts []byte, ptr any) error
)
