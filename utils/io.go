package utils

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
)

type readerFunc func(p []byte) (n int, err error)

func (f readerFunc) Read(p []byte) (n int, err error) {
	return f(p)
}

// CopyClose 可中断的流复制
func CopyClose(ctx context.Context, dst io.Writer, src io.ReadCloser) (int64, error) {
	defer src.Close()
	return io.Copy(dst, readerFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	}))
}

// WaitForNBytes 等待r中满足N个byte
func WaitForNBytes(r io.ReadSeeker, N int, interval time.Duration, retryTime int) (err error) {
	var (
		t []byte
		i int
	)
	for i = 0; i < retryTime; i++ {
		t = make([]byte, N)
		_, err = io.ReadFull(r, t)
		_, _ = r.Seek(0, 0)
		if err != io.EOF {
			break
		}
		<-time.After(interval)
	}
	if i >= retryTime {
		err = errors.Errorf("Reader is less than %d bytes", N)
	}
	return err
}
