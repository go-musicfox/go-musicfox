package strutil

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"net/url"
	"strings"
	"text/template"
)

//
// -------------------- escape --------------------
//

// EscapeJS escape javascript string
func EscapeJS(s string) string {
	return template.JSEscapeString(s)
}

// EscapeHTML escape html string
func EscapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}

// AddSlashes add slashes for the string.
func AddSlashes(s string) string {
	if ln := len(s); ln == 0 {
		return ""
	}

	var buf bytes.Buffer
	for _, char := range s {
		switch char {
		case '\'', '"', '\\':
			buf.WriteRune('\\')
		}
		buf.WriteRune(char)
	}

	return buf.String()
}

// StripSlashes strip slashes for the string.
func StripSlashes(s string) string {
	ln := len(s)
	if ln == 0 {
		return ""
	}

	var skip bool
	var buf bytes.Buffer

	for i, char := range s {
		if skip {
			skip = false
		} else if char == '\\' {
			if i+1 < ln && s[i+1] == '\\' {
				skip = true
			}
			continue
		}
		buf.WriteRune(char)
	}

	return buf.String()
}

//
// -------------------- encode --------------------
//

// URLEncode encode url string.
func URLEncode(s string) string {
	if pos := strings.IndexRune(s, '?'); pos > -1 { // escape query data
		return s[0:pos+1] + url.QueryEscape(s[pos+1:])
	}
	return s
}

// URLDecode decode url string.
func URLDecode(s string) string {
	if pos := strings.IndexRune(s, '?'); pos > -1 { // un-escape query data
		qy, err := url.QueryUnescape(s[pos+1:])
		if err == nil {
			return s[0:pos+1] + qy
		}
	}

	return s
}

//
// -------------------- base encode --------------------
//

// B32Encode base32 encode
func B32Encode(str string) string {
	return base32.StdEncoding.EncodeToString([]byte(str))
}

// B32Decode base32 decode
func B32Decode(str string) string {
	dec, _ := base32.StdEncoding.DecodeString(str)
	return string(dec)
}

// B64Encode base64 encode
func B64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// B64EncodeBytes base64 encode
func B64EncodeBytes(src []byte) []byte {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(buf, src)

	return buf
}

// B64Decode base64 decode
func B64Decode(str string) string {
	dec, _ := base64.StdEncoding.DecodeString(str)
	return string(dec)
}

// B64DecodeBytes base64 decode
func B64DecodeBytes(str string) []byte {
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(str)))
	n, _ := base64.StdEncoding.Decode(dbuf, []byte(str))

	return dbuf[:n]
}

// BaseEncoder interface
type BaseEncoder interface {
	Encode(dst []byte, src []byte)
	EncodeToString(src []byte) string
	Decode(dst []byte, src []byte) (n int, err error)
	DecodeString(s string) ([]byte, error)
}

// BaseType for base encoding
type BaseType uint8

// types for base encoding
const (
	BaseTypeStd BaseType = iota
	BaseTypeHex
	BaseTypeURL
	BaseTypeRawStd
	BaseTypeRawURL
)

// Encoding instance
func Encoding(base int, typ BaseType) BaseEncoder {
	if base == 32 {
		switch typ {
		case BaseTypeHex:
			return base32.HexEncoding
		default:
			return base32.StdEncoding
		}
	}

	// base 64
	switch typ {
	case BaseTypeURL:
		return base64.URLEncoding
	case BaseTypeRawURL:
		return base64.RawURLEncoding
	case BaseTypeRawStd:
		return base64.RawStdEncoding
	default:
		return base64.StdEncoding
	}
}
