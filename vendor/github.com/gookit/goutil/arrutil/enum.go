package arrutil

import (
	"strconv"
	"strings"
)

// Ints type
type Ints []int

// String to string
func (is Ints) String() string {
	ss := make([]string, len(is))
	for i, iv := range is {
		ss[i] = strconv.Itoa(iv)
	}
	return strings.Join(ss, ",")
}

// Has given element
func (is Ints) Has(i int) bool {
	for _, iv := range is {
		if i == iv {
			return true
		}
	}
	return false
}

// Strings type
type Strings []string

// String to string
func (ss Strings) String() string {
	return strings.Join(ss, ",")
}

// Join to string
func (ss Strings) Join(sep string) string {
	return strings.Join(ss, sep)
}

// Has given element
func (ss Strings) Has(sub string) bool {
	for _, s := range ss {
		if s == sub {
			return true
		}
	}
	return false
}
