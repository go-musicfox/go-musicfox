package openssl

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// Sha1 Calculate the sha1 hash of a string
func Sha1(str string) []byte {
	h := sha1.New()
	_, _ = h.Write([]byte(str))
	return h.Sum(nil)
}

// HmacSha1 Calculate the sha1 hash of a string using the HMAC method
func HmacSha1(key string, data string) []byte {
	mac := hmac.New(sha1.New, []byte(key))
	_, _ = mac.Write([]byte(data))

	return mac.Sum(nil)
}

// HmacSha1ToString Calculate the sha1 hash of a string using the HMAC method, outputs lowercase hexits
func HmacSha1ToString(key string, data string) string {
	return hex.EncodeToString(HmacSha1(key, data))
}

// Sha256 Calculate the sha256 hash of a string
func Sha256(str string) []byte {
	h := sha256.New()
	_, _ = h.Write([]byte(str))
	return h.Sum(nil)
}

// HmacSha256 Calculate the sha256 hash of a string using the HMAC method
func HmacSha256(key string, data string) []byte {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(data))

	return mac.Sum(nil)
}

// HmacSha256ToString Calculate the sha256 hash of a string using the HMAC method, outputs lowercase hexits
func HmacSha256ToString(key string, data string) string {
	return hex.EncodeToString(HmacSha256(key, data))
}
