// Package sysutil provide some system util functions. eg: sysenv, exec, user, process
package sysutil

import (
	"os"
	"path"
	"path/filepath"
)

// Workdir get
func Workdir() string {
	dir, _ := os.Getwd()
	return dir
}

// BinDir get
func BinDir() string {
	binFile := os.Args[0]
	return path.Dir(binFile)
}

// BinFile get
func BinFile() string {
	return os.Args[0]
}

// Open file or url address
func Open(fileOrUrl string) error {
	return OpenURL(fileOrUrl)
}

// OpenBrowser file or url address
func OpenBrowser(fileOrUrl string) error {
	return OpenURL(fileOrUrl)
}

// OpenFile opens new browser window for the file path.
func OpenFile(path string) error {
	fpath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return OpenURL("file://" + fpath)
}
