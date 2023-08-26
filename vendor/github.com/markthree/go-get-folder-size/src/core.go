package core

import (
	"io/fs"
	"os"
	"path"
)

type EntrySizeChan chan int64

func ValuePromise[T any](ch chan T) (func(T), func() T) {
	resolve := func(result T) {
		ch <- result
	}
	value := func() T {
		return <-ch
	}
	return resolve, value
}

func StatusPromise[T any](ch chan T, err chan error) (func(T), func(error), func() (T, error)) {
	resolve := func(result T) {
		ch <- result
	}
	reject := func(errValue error) {
		err <- errValue
	}
	status := func() (T, error) {
		select {
		case value := <-ch:
			return value, nil
		case errValue := <-err:
			var zero T
			return zero, errValue
		}
	}
	return resolve, reject, status
}

func Invoke(folder string) (size int64, e error) {
	gracefulExit := func(err error) {
		e = err
		size = 0
	}

	entrys, err := os.ReadDir(folder)
	if err != nil {
		gracefulExit(err)
		return
	}
	entrysLen := len(entrys)
	if entrysLen == 0 {
		gracefulExit(nil)
		return
	}
	errChan := make(chan error)
	sizeChan := make(EntrySizeChan, entrysLen)

	resolve, reject, status := StatusPromise(sizeChan, errChan)

	for i := 0; i < entrysLen; i++ {
		go func(entry fs.DirEntry) {
			if entry.IsDir() {
				subFolderSize, err := Invoke(path.Join(folder, entry.Name()))
				if err != nil {
					reject(err)
					return
				}
				resolve(subFolderSize)
				return
			}

			info, err := entry.Info()
			if err != nil {
				reject(err)
				return
			}
			resolve(info.Size())
		}(entrys[i])
	}

	for i := 0; i < entrysLen; i++ {
		newSize, newErr := status()
		if newErr != nil {
			gracefulExit(newErr)
			return
		}
		size += newSize
	}
	return size, nil
}

func LooseInvoke(folder string) int64 {
	size := int64(0)

	entrys, err := os.ReadDir(folder)
	if err != nil {
		return 0
	}
	entrysLen := len(entrys)
	if entrysLen == 0 {
		return 0
	}
	sizeChan := make(EntrySizeChan, entrysLen)
	resolve, value := ValuePromise(sizeChan)

	for i := 0; i < entrysLen; i++ {
		go func(entry fs.DirEntry) {
			if entry.IsDir() {
				subFolderSize := LooseInvoke(path.Join(folder, entry.Name()))
				resolve(subFolderSize)
				return
			}
			info, err := entry.Info()
			if err != nil {
				resolve(0)
				return
			}
			resolve(info.Size())
		}(entrys[i])
	}
	for i := 0; i < entrysLen; i++ {
		size += value()
	}
	return size
}
