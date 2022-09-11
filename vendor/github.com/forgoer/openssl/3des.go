package openssl

import "crypto/des"

// Des3ECBEncrypt
func Des3ECBEncrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	return ECBEncrypt(block, src, padding)
}

// Des3ECBDecrypt
func Des3ECBDecrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	return ECBDecrypt(block, src, padding)
}

// Des3CBCEncrypt
func Des3CBCEncrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCEncrypt(block, src, iv, padding)
}

// Des3CBCDecrypt
func Des3CBCDecrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCDecrypt(block, src, iv, padding)
}
