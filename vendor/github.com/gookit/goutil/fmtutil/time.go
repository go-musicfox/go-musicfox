package fmtutil

import (
	"github.com/gookit/goutil/basefn"
)

// HowLongAgo format a seconds, get how lang ago
func HowLongAgo(sec int64) string {
	return basefn.HowLongAgo(sec)
}
