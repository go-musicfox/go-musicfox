// Package byteutil Provide some bytes utils functions or structs
package byteutil

import (
	"crypto/md5"
	"fmt"
)

// Md5 Generate a 32-bit md5 bytes
func Md5(src any) []byte {
	h := md5.New()

	if s, ok := src.(string); ok {
		h.Write([]byte(s))
	} else {
		h.Write([]byte(fmt.Sprint(src)))
	}
	return h.Sum(nil)
}
