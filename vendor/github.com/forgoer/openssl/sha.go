package openssl

import (
	"crypto/hmac"
	"crypto/sha1"
)

// Sha1 Calculate the sha1 hash of a string
func Sha1(str string) []byte {
	h := sha1.New()
	_, _ = h.Write([]byte(str))
	return h.Sum(nil)
}

func HmacSha1(key string, data string) []byte {
	mac := hmac.New(sha1.New, []byte(key))
	_, _ = mac.Write([]byte(data))

	return mac.Sum(nil)
}
