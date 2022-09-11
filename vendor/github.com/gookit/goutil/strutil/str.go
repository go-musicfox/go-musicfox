package strutil

import "strings"

// Str string
type Str string

// IsStartBy prefix
func (s Str) IsStartBy(sub string) bool {
	return strings.HasPrefix(string(s), sub)
}

// IsEndBy suffix
func (s Str) IsEndBy(sub string) bool {
	return strings.HasSuffix(string(s), sub)
}

// Bytes string to bytes
func (s Str) Bytes() []byte {
	return []byte(s)
}

// Get string
func (s Str) Get() string {
	return string(s)
}

// String string
func (s Str) String() string {
	return string(s)
}

// TrimSpace string
func (s Str) TrimSpace() Str {
	return Str(strings.TrimSpace(string(s)))
}
