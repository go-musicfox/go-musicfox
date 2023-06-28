package openssl

import (
	"crypto/md5"
	"encoding/hex"
)

// Md5 Calculate the md5 hash of a string
func Md5(str string) []byte {
	h := md5.New()
	_, _ = h.Write([]byte(str))
	return h.Sum(nil)
}

// Md5ToString Calculate the md5 hash of a string, return hex string
func Md5ToString(str string) string {
	return hex.EncodeToString(Md5(str))
}
