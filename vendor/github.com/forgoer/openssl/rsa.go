package openssl

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

// RSAGenerateKey generate RSA private key
func RSAGenerateKey(bits int, out io.Writer) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)

	privateBlock := pem.Block{Type: "RSA PRIVATE KEY", Bytes: X509PrivateKey}

	return pem.Encode(out, &privateBlock)
}

// RSAGeneratePublicKey generate RSA public key
func RSAGeneratePublicKey(priKey []byte, out io.Writer) error {
	block, _ := pem.Decode(priKey)
	if block == nil{
		return errors.New("key is invalid format")
	}

	// x509 parse
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	publicKey := privateKey.PublicKey
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}

	publicBlock := pem.Block{Type: "RSA PUBLIC KEY", Bytes: X509PublicKey}

	return pem.Encode(out, &publicBlock)
}

// RSAEncrypt RSA encrypt
func RSAEncrypt(src, pubKey []byte) ([]byte, error) {
	block, _ := pem.Decode(pubKey)
	if block == nil{
		return nil, errors.New("key is invalid format")
	}

	// x509 parse
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("the kind of key is not a rsa.PublicKey")
	}
	// encrypt
	dst, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, src)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

// RSADecrypt RSA decrypt
func RSADecrypt(src, priKey []byte) ([]byte, error) {
	block, _ := pem.Decode(priKey)
	if block == nil{
		return nil, errors.New("key is invalid format")
	}

	// x509 parse
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	dst, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, src)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

// RSASign RSA sign, use crypto.SHA256
func RSASign(src []byte, priKey []byte) ([]byte, error) {
	block, _ := pem.Decode(priKey)
	if block == nil{
		return nil, errors.New("key is invalid format")
	}

	// x509 parse
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	_, err = hash.Write(src)
	if err != nil {
		return nil, err
	}

	bytes := hash.Sum(nil)
	sign, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, bytes)
	if err != nil {
		return nil, err
	}

	return sign, nil
}

// RSAVerify RSA Verify
func RSAVerify(src, sign, pubKey []byte) error {
	block, _ := pem.Decode(pubKey)
	if block == nil{
		return errors.New("key is invalid format")
	}

	// x509 parse
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return errors.New("the kind of key is not a rsa.PublicKey")
	}

	hash := sha256.New()
	_, err = hash.Write(src)
	if err != nil {
		return err
	}

	bytes := hash.Sum(nil)

	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, bytes, sign)
}
