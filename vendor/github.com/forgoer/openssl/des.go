package openssl

import (
	"crypto/des"
)

// DesECBEncrypt
func DesECBEncrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ECBEncrypt(block, src, padding)
}

// DesECBDecrypt
func DesECBDecrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return ECBDecrypt(block, src, padding)
}

// DesCBCEncrypt
func DesCBCEncrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCEncrypt(block, src, iv, padding)
}

// DesCBCDecrypt
func DesCBCDecrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCDecrypt(block, src, iv, padding)
}
