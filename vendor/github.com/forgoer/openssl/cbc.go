package openssl

import (
	"bytes"
	"crypto/cipher"
)

// CBCEncrypt
func CBCEncrypt(block cipher.Block, src, iv []byte, padding string) ([]byte, error) {
	blockSize := block.BlockSize()
	src = Padding(padding, src, blockSize)

	encryptData := make([]byte, len(src))

	if len(iv) != block.BlockSize() {
		// auto pad length to block size
		iv = cbcIVPending(iv, block.BlockSize())
		//return nil, errors.New("CBCEncrypt: IV length must equal block size")
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptData, src)

	return encryptData, nil
}

// CBCDecrypt
func CBCDecrypt(block cipher.Block, src, iv []byte, padding string) ([]byte, error) {

	dst := make([]byte, len(src))

	if len(iv) != block.BlockSize() {
		// auto pad length to block size
		iv = cbcIVPending(iv, block.BlockSize())
		//return nil, errors.New("CBCDecrypt: IV length must equal block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, src)

	return UnPadding(padding, dst)
}

// cbcIVPending auto pad length to block size
func cbcIVPending(iv []byte, blockSize int) []byte {
	k := len(iv)
	if k < blockSize {
		return append(iv, bytes.Repeat([]byte{0}, blockSize-k)...)
	} else if k > blockSize {
		return iv[0:blockSize]
	}

	return iv
}
