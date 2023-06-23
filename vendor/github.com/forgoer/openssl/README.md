# Openssl encryption

[![Default](https://github.com/forgoer/openssl/workflows/build/badge.svg?branch=master)](https://github.com/forgoer/openssl/actions)
[![Coverage Status](https://coveralls.io/repos/github/forgoer/openssl/badge.svg?branch=master)](https://coveralls.io/github/forgoer/openssl?branch=master)

A functions wrapping of OpenSSL library for symmetric and asymmetric encryption and decryption

- [AES](#AES)
- [DES](#DES)
- [3DES](#DES)
- [RSA](#RSA)
- [HMAC-SHA](#HMAC-SHA)

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl/)

```
go get -u github.com/forgoer/openssl
```

## Usage

### AES

The length of the key can be 16/24/32 characters (128/192/256 bits)

AES-ECB:

```go 
src := []byte("123456")
key := []byte("1234567890123456")
dst , _ := openssl.AesECBEncrypt(src, key, openssl.PKCS7_PADDING)
fmt.Printf(base64.StdEncoding.EncodeToString(dst))  // yXVUkR45PFz0UfpbDB8/ew==

dst , _ = openssl.AesECBDecrypt(dst, key, openssl.PKCS7_PADDING)
fmt.Println(string(dst)) // 123456
```

AES-CBC:

```go
src := []byte("123456")
key := []byte("1234567890123456")
iv := []byte("1234567890123456")
dst , _ := openssl.AesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
fmt.Println(base64.StdEncoding.EncodeToString(dst)) // 1jdzWuniG6UMtoa3T6uNLA==

dst , _ = openssl.AesCBCDecrypt(dst, key, iv, openssl.PKCS7_PADDING)
fmt.Println(string(dst)) // 123456
```

### DES

The length of the key must be 8 characters (64 bits).

DES-ECB:

```go
openssl.DesECBEncrypt(src, key, openssl.PKCS7_PADDING)
openssl.DesECBDecrypt(src, key, openssl.PKCS7_PADDING)
```

DES-CBC:

```go
openssl.DesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
openssl.DesCBCDecrypt(src, key, iv, openssl.PKCS7_PADDING)
```

### 3DES

The length of the key must be 24 characters (192 bits).

3DES-ECB:

```go
openssl.Des3ECBEncrypt(src, key, openssl.PKCS7_PADDING)
openssl.Des3ECBDecrypt(src, key, openssl.PKCS7_PADDING)
```

3DES-CBC:

```go
openssl.Des3CBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
openssl.Des3CBCDecrypt(src, key, iv, openssl.PKCS7_PADDING)
```

### RSA

```go
openssl.RSAGenerateKey(bits int, out io.Writer)
openssl.RSAGeneratePublicKey(priKey []byte, out io.Writer)

openssl.RSAEncrypt(src, pubKey []byte) ([]byte, error)
openssl.RSADecrypt(src, priKey []byte) ([]byte, error)

openssl.RSASign(src []byte, priKey []byte, hash crypto.Hash) ([]byte, error)
openssl.RSAVerify(src, sign, pubKey []byte, hash crypto.Hash) error
```

### HMAC-SHA

```
// Sha1 Calculate the sha1 hash of a string
Sha1(str string) []byte

// HmacSha1 Calculate the sha1 hash of a string using the HMAC method
HmacSha1(key string, data string) []byte

// HmacSha1ToString Calculate the sha1 hash of a string using the HMAC method, outputs lowercase hexits
HmacSha1ToString(key string, data string) string

// Sha256 Calculate the sha256 hash of a string
Sha256(str string) []byte

// HmacSha256 Calculate the sha256 hash of a string using the HMAC method
HmacSha256(key string, data string) []byte

// HmacSha256ToString Calculate the sha256 hash of a string using the HMAC method, outputs lowercase hexits
HmacSha256ToString(key string, data string) string
```

## License

This project is licensed under the [Apache 2.0 license](LICENSE).

## Contact

If you have any issues or feature requests, please contact us. PR is welcomed.
- https://github.com/forgoer/openssl/issues

