package strutil

import "strings"

// Value string
type Value string

// StrVal string. alias of Value
type StrVal = Value

// Set value
func (s *Value) Set(val string) error {
	*s = Value(val)
	return nil
}

// IsEmpty check
func (s *Value) IsEmpty() bool {
	return string(*s) == ""
}

// IsStartWith prefix
func (s *Value) IsStartWith(sub string) bool {
	return strings.HasPrefix(string(*s), sub)
}

// HasPrefix prefix
func (s *Value) HasPrefix(sub string) bool {
	return strings.HasPrefix(string(*s), sub)
}

// IsEndWith suffix
func (s *Value) IsEndWith(sub string) bool {
	return strings.HasSuffix(string(*s), sub)
}

// HasSuffix suffix
func (s *Value) HasSuffix(sub string) bool {
	return strings.HasSuffix(string(*s), sub)
}

// Bytes string to bytes
func (s *Value) Bytes() []byte {
	return []byte(*s)
}

// Val string
func (s *Value) Val() string {
	return string(*s)
}

// Int convert
func (s *Value) Int() int {
	return QuietInt(string(*s))
}

// Int64 convert
func (s *Value) Int64() int64 {
	return QuietInt64(string(*s))
}

// Bool convert
func (s *Value) Bool() bool {
	return QuietBool(string(*s))
}

// Value string
func (s *Value) String() string {
	return string(*s)
}

// Split string
func (s *Value) Split(sep string) []string {
	return strings.Split(string(*s), sep)
}

// SplitN string
func (s *Value) SplitN(sep string, n int) []string {
	return strings.SplitN(string(*s), sep, n)
}

// TrimSpace string and return new
func (s *Value) TrimSpace() Value {
	return Value(strings.TrimSpace(string(*s)))
}
