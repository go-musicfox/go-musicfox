package slogx

import "testing"

func TestFlushDoesNotPanic(t *testing.T) {
	Flush()
}
