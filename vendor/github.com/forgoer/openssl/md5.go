package openssl

import (
	"crypto/md5"
)

// Md5 Calculate the md5 hash of a string
func Md5(str string) []byte {
	h := md5.New()
	_, _ = h.Write([]byte(str))
	return h.Sum(nil)
}
