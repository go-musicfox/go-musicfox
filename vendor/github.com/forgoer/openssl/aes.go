package openssl

import (
	"crypto/aes"
)

// AesECBEncrypt
func AesECBEncrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ECBEncrypt(block, src, padding)
}

// AesECBDecrypt
func AesECBDecrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return ECBDecrypt(block, src, padding)
}

// AesCBCEncrypt
func AesCBCEncrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCEncrypt(block, src, iv, padding)
}

// AesCBCDecrypt
func AesCBCDecrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCDecrypt(block, src, iv, padding)
}
