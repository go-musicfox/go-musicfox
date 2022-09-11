package strutil

import (
	"encoding/base64"
	"net/url"
	"strings"
	"text/template"
)

var (
	// EscapeJS escape javascript string
	EscapeJS = template.JSEscapeString
	// EscapeHTML escape html string
	EscapeHTML = template.HTMLEscapeString
)

// Base64 encode
func Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// B64Encode base64 encode
func B64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

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
