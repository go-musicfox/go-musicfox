package jsonutil

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/scanner"
)

// WriteFile write data to JSON file
func WriteFile(filePath string, data interface{}) error {
	jsonBytes, err := Encode(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, jsonBytes, 0664)
}

// ReadFile Read JSON file data
func ReadFile(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	return json.NewDecoder(file).Decode(v)
}

// Pretty JSON string and return
func Pretty(v interface{}) (string, error) {
	out, err := json.MarshalIndent(v, "", "    ")
	return string(out), err
}

// Encode data to json bytes.
func Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// EncodePretty encode pretty JSON data to json bytes.
func EncodePretty(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}

// EncodeToWriter encode data to writer.
func EncodeToWriter(v interface{}, w io.Writer) error {
	return json.NewEncoder(w).Encode(v)
}

// EncodeUnescapeHTML data to json bytes. will close escape HTML
func EncodeUnescapeHTML(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode json bytes to data ptr.
func Decode(bts []byte, ptr interface{}) error {
	return json.Unmarshal(bts, ptr)
}

// DecodeString json string to data ptr.
func DecodeString(str string, ptr interface{}) error {
	return json.Unmarshal([]byte(str), ptr)
}

// DecodeReader decode JSON from io reader.
func DecodeReader(r io.Reader, ptr interface{}) error {
	return json.NewDecoder(r).Decode(ptr)
}

// `(?s:` enable match multi line
var jsonMLComments = regexp.MustCompile(`(?s:/\*.*?\*/\s*)`)

// StripComments strip comments for a JSON string
func StripComments(src string) string {
	// multi line comments
	if strings.Contains(src, "/*") {
		src = jsonMLComments.ReplaceAllString(src, "")
	}

	// single line comments
	if !strings.Contains(src, "//") {
		return strings.TrimSpace(src)
	}

	// strip inline comments
	var s scanner.Scanner

	s.Init(strings.NewReader(src))
	s.Filename = "comments"
	s.Mode ^= scanner.SkipComments // don't skip comments

	buf := new(bytes.Buffer)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		txt := s.TokenText()
		if !strings.HasPrefix(txt, "//") && !strings.HasPrefix(txt, "/*") {
			buf.WriteString(txt)
			// } else {
			// fmt.Printf("%s: %s\n", s.Position, txt)
		}
	}

	return buf.String()
}
