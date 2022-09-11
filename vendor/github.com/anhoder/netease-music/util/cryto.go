package util

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/forgoer/openssl"
	"math/big"
	"strings"
)

var iv = []byte("0102030405060708")
var presetKey = []byte("0CoJUm6Qyw8W8jud")
var linuxapiKey = []byte("rFgB&h#%2?^eDg:Q")
var stdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var publicKey = []byte("-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDgtQn2JZ34ZC28NWYpAUd98iZ37BUrX/aKzmFbt7clFSs6sXqHauqKWqdtLkF2KexO40H1YTX8z2lSgBBOAxLsvaklV8k4cBFK9snQXE9/DDaFt6Rr7iVZMldczhC0JNgTz+SHXT6CBHuX3e9SdB1Ua44oncaTWz7OBGLbCiK45wIDAQAB\n-----END PUBLIC KEY-----")
var eapiKey = []byte("e82ckenh8dichen8")

func aesEncryptCBC(buffer []byte, key []byte, ivv []byte) []byte {
	dst, _ := openssl.AesCBCEncrypt(buffer, key, ivv, openssl.PKCS7_PADDING)
	return dst
	// base64 解码
	//fmt.Println(base64.StdEncoding.EncodeToString(dst))

	// 解密
	//dst, _ = openssl.AesCBCDecrypt(dst, presetKey, iv, openssl.PKCS7_PADDING)
	//fmt.Println(string(dst)) // 123456
}

func aesEncryptECB(buffer []byte, key []byte) []byte {
	dst, _ := openssl.AesECBEncrypt(buffer, key, openssl.PKCS7_PADDING)
	return dst
	//fmt.Println(base64.StdEncoding.EncodeToString(dst))  // yXVUkR45PFz0UfpbDB8/ew==
	// hex 解码
	//fmt.Println(hex.EncodeToString(dst))

	//解密
	//dst, _ = openssl.AesECBDecrypt(dst, linuxapiKey, openssl.PKCS7_PADDING)
	//fmt.Println(string(dst)) // 123456
}

func NewLen16Rand() ([]byte, []byte) {
	randByte := make([]byte, 16)
	randByteReverse := make([]byte, 16)
	for i := 0; i < 16; i++ {
		result, _ := rand.Int(rand.Reader, big.NewInt(62))
		randByte[i] = stdChars[result.Int64()]
		randByteReverse[15-i] = stdChars[result.Int64()]
	}
	return randByte, randByteReverse
}

func aesEncrypt(buffer []byte, mod string, key []byte, ivv []byte) []byte {
	if mod == "cbc" {
		return aesEncryptCBC(buffer, key, ivv)
	} else if mod == "ecb" {
		return aesEncryptECB(buffer, key)
	}
	return nil
}

func rsaEncrypt(buffer []byte, key []byte) []byte {
	buffers := make([]byte, 128-16, 128)
	buffers = append(buffers, buffer...)
	block, _ := pem.Decode(key)
	if block == nil {
		return nil
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)

	// 加密 因为网易采用的是无padding加密故直接进行计算
	c := new(big.Int).SetBytes([]byte(buffers))
	encryptedBytes := c.Exp(c, big.NewInt(int64(pub.E)), pub.N).Bytes()
	return encryptedBytes
	////加密
	//a,err:=rsa.EncryptPKCS1v15(rand.Reader, pub, buffers)
	//if err!=nil{
	//	fmt.Println(err.Error())
	//}
	//return a
}

func Weapi(data map[string]string) map[string]string {
	text, _ := json.Marshal(data)
	//fmt.Println(string(text))
	secretKey, reSecretKey := NewLen16Rand()
	//secretKey,_=hex.DecodeString("3554324955624839667a7679634f3372")
	//reSecretKey,_=hex.DecodeString("72334f6379767a663948625549325435")
	weapiType := make(map[string]string, 2)
	weapiType["params"] = base64.StdEncoding.EncodeToString(aesEncrypt([]byte(base64.StdEncoding.EncodeToString(aesEncrypt(text, "cbc", presetKey, iv))), "cbc", reSecretKey, iv))
	weapiType["encSecKey"] = hex.EncodeToString(rsaEncrypt(secretKey, publicKey))
	return weapiType
}

func Linuxapi(data map[string]interface{}) map[string]string {
	text, _ := json.Marshal(data)
	linuxapiType := make(map[string]string, 1)
	linuxapiType["eparams"] = strings.ToUpper(hex.EncodeToString(aesEncrypt(text, "ecb", linuxapiKey, nil)))
	return linuxapiType
}

func Eapi(url string, data map[string]interface{}) map[string]string {
	textByte, _ := json.Marshal(data)
	message := "nobody" + url + "use" + string(textByte) + "md5forencrypt"
	h := md5.New()
	h.Write([]byte(message))
	digest := hex.EncodeToString(h.Sum(nil))
	dd := url + "-36cd479b6b5-" + string(textByte) + "-36cd479b6b5-" + digest
	eapiType := make(map[string]string, 1)
	eapiType["params"] = strings.ToUpper(hex.EncodeToString(aesEncrypt([]byte(dd), "ecb", eapiKey, nil)))
	return eapiType
}
