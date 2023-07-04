package core

import (
	"fmt"
	"os"
	"path"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/ants/v2"
)

var pool, _ = ants.NewPool(1000000)

func calc(folder string) (total int64, e error) {
	entrys, err := os.ReadDir(folder)
	if err != nil {
		return 0, err
	}
	entrysLen := len(entrys)
	if entrysLen == 0 {
		return 0, nil
	}
	var wg sync.WaitGroup
	wg.Add(entrysLen)

	fileSize := int64(0)
	for i := 0; i < entrysLen; i++ {
		entry := entrys[i]
		if entry.IsDir() {
			pool.Submit(func() {
				defer wg.Done()
				size, err := calc(path.Join(folder, entry.Name()))
				if err != nil {
					panic(err)
				}
				atomic.AddInt64(&total, size)
			})
			continue
		}
		// Normal files
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}
		fileSize += info.Size()
		wg.Done()
	}
	wg.Wait()
	total += fileSize
	return total, nil
}

// Parallel execution, fast enough
func Parallel(folder string) (total int64, e error) {
	// catch panic
	defer func() {
		if err := recover(); err != nil {
			e = fmt.Errorf("%v", err)
		}
	}()

	total, err := calc(folder)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func looseCalc(folder string) (total int64) {
	entrys, err := os.ReadDir(folder)
	if err != nil {
		return 0
	}
	entrysLen := len(entrys)
	if entrysLen == 0 {
		return 0
	}
	var wg sync.WaitGroup
	wg.Add(entrysLen)

	fileSize := int64(0)
	for i := 0; i < entrysLen; i++ {
		entry := entrys[i]

		if entry.IsDir() {
			pool.Submit(func() {
				defer wg.Done()
				size := looseCalc(path.Join(folder, entry.Name()))
				atomic.AddInt64(&total, size)
			})
			continue
		}
		// Normal files
		info, err := entry.Info()
		if err == nil {
			fileSize += info.Size()
		}
		wg.Done()
	}
	wg.Wait()
	total += fileSize
	return total
}

func LooseParallel(folder string) (total int64) {
	return looseCalc(folder)
}
