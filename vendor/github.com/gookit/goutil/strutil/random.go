package strutil

import (
	"encoding/base64"
	mathRand "math/rand"
	"time"

	"github.com/gookit/goutil/byteutil"
)

// some consts string chars
const (
	Numbers  = "0123456789"
	HexChars = "0123456789abcdef"

	AlphaBet  = "abcdefghijklmnopqrstuvwxyz"
	AlphaBet1 = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"

	AlphaNum  = "abcdefghijklmnopqrstuvwxyz0123456789"
	AlphaNum2 = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphaNum3 = "0123456789AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"

	Base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Base64Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
)

// RandomChars generate give length random chars at `a-z`
func RandomChars(ln int) string {
	cs := make([]byte, ln)
	for i := 0; i < ln; i++ {
		// 1607400451937462000
		mathRand.Seed(time.Now().UnixNano())
		idx := mathRand.Intn(25) // 0 - 25
		cs[i] = AlphaBet[idx]
	}

	return string(cs)
}

// RandomCharsV2 generate give length random chars in `0-9a-z`
func RandomCharsV2(ln int) string {
	cs := make([]byte, ln)
	for i := 0; i < ln; i++ {
		// 1607400451937462000
		mathRand.Seed(time.Now().UnixNano())
		idx := mathRand.Intn(35) // 0 - 35
		cs[i] = AlphaNum[idx]
	}

	return string(cs)
}

// RandomCharsV3 generate give length random chars in `0-9a-zA-Z`
func RandomCharsV3(ln int) string {
	cs := make([]byte, ln)
	for i := 0; i < ln; i++ {
		// 1607400451937462000
		mathRand.Seed(time.Now().UnixNano())
		idx := mathRand.Intn(61) // 0 - 61
		cs[i] = AlphaNum2[idx]
	}

	return string(cs)
}

// RandomBytes generate
func RandomBytes(length int) ([]byte, error) {
	return byteutil.Random(length)
}

// RandomString generate.
//
// Example:
//
//		// this will give us a 44 byte, base64 encoded output
//		token, err := RandomString(32)
//		if err != nil {
//	    // Serve an appropriately vague error to the
//	    // user, but log the details internally.
//		}
func RandomString(length int) (string, error) {
	b, err := RandomBytes(length)
	return base64.URLEncoding.EncodeToString(b), err
}

// RandWithTpl generate random string with give template
func RandWithTpl(n int, letters string) string {
	if len(letters) == 0 {
		letters = AlphaNum2
	}

	ln := len(letters)
	cs := make([]byte, n)
	for i := 0; i < n; i++ {
		mathRand.Seed(int64(time.Now().Nanosecond()))
		idx := mathRand.Intn(ln)
		cs[i] = letters[idx]
	}

	return byteutil.String(cs)
}
