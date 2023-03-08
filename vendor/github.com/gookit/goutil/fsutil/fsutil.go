// Package fsutil Filesystem util functions, quick create, read and write file. eg: file and dir check, operate
package fsutil

import (
	"io"
	"net/http"
	"os"
)

const (
	// MimeSniffLen sniff Length, use for detect file mime type
	MimeSniffLen = 512
)

// OSTempFile create a temp file on os.TempDir()
//
// Usage:
//
//	fsutil.OSTempFile("example.*.txt")
func OSTempFile(pattern string) (*os.File, error) {
	return os.CreateTemp(os.TempDir(), pattern)
}

// TempFile is like os.CreateTemp, but can custom temp dir.
//
// Usage:
//
//	fsutil.TempFile("", "example.*.txt")
func TempFile(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

// OSTempDir creates a new temp dir on os.TempDir and return the temp dir path
//
// Usage:
//
//	fsutil.OSTempDir("example.*")
func OSTempDir(pattern string) (string, error) {
	return os.MkdirTemp(os.TempDir(), pattern)
}

// TempDir creates a new temp dir and return the temp dir path
//
// Usage:
//
//	fsutil.TempDir("", "example.*")
//	fsutil.TempDir("testdata", "example.*")
func TempDir(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// MimeType get File Mime Type name. eg "image/png"
func MimeType(path string) (mime string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	return ReaderMimeType(file)
}

// ReaderMimeType get the io.Reader mimeType
//
// Usage:
//
//	file, err := os.Open(filepath)
//	if err != nil {
//		return
//	}
//	mime := ReaderMimeType(file)
func ReaderMimeType(r io.Reader) (mime string) {
	var buf [MimeSniffLen]byte
	n, _ := io.ReadFull(r, buf[:])
	if n == 0 {
		return ""
	}

	return http.DetectContentType(buf[:n])
}
