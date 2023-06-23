package openssl

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// AesECBEncrypt
func AesECBEncrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := AesNewCipher(key)
	if err != nil {
		return nil, err
	}
	return ECBEncrypt(block, src, padding)
}

// AesECBDecrypt
func AesECBDecrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := AesNewCipher(key)
	if err != nil {
		return nil, err
	}

	return ECBDecrypt(block, src, padding)
}

// AesCBCEncrypt
func AesCBCEncrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := AesNewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCEncrypt(block, src, iv, padding)
}

// AesCBCDecrypt
func AesCBCDecrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := AesNewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCDecrypt(block, src, iv, padding)
}

// AesNewCipher creates and returns a new AES cipher.Block.
// it will automatically pad the length of the key.
func AesNewCipher(key []byte) (cipher.Block, error) {
	return aes.NewCipher(aesKeyPending(key))
}

// aesKeyPending The length of the key can be 16/24/32 characters (128/192/256 bits)
func aesKeyPending(key []byte) []byte {
	k := len(key)
	count := 0
	switch true {
	case k <= 16:
		count = 16 - k
	case k <= 24:
		count = 24 - k
	case k <= 32:
		count = 32 - k
	default:
		return key[:32]
	}
	if count == 0 {
		return key
	}

	return append(key, bytes.Repeat([]byte{0}, count)...)
}
