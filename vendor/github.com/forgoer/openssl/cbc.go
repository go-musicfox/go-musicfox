package openssl

import (
	"crypto/cipher"
	"errors"
)

// CBCEncrypt
func CBCEncrypt(block cipher.Block, src, iv []byte, padding string) ([]byte, error) {
	blockSize := block.BlockSize()
	src = Padding(padding, src, blockSize)

	encryptData := make([]byte, len(src))

	if len(iv) != block.BlockSize() {
		return nil, errors.New("CBCEncrypt: IV length must equal block size")
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptData, src)

	return encryptData, nil
}

// CBCDecrypt
func CBCDecrypt(block cipher.Block, src, iv []byte, padding string) ([]byte, error) {

	dst := make([]byte, len(src))

	if len(iv) != block.BlockSize() {
		return nil, errors.New("CBCDecrypt: IV length must equal block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, src)

	return UnPadding(padding, dst)
}
