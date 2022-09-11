package sysutil

import (
	"os"
	"path"
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
