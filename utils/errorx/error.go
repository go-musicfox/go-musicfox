// Package errorx 提供错误与 panic 处理工具：panic 捕获记录到日志、
// goroutine 安全启动与崩溃提示。
package errorx

import (
	"fmt"
)

func Must(err error) {
	if err != nil {
		panic(fmt.Sprintf("caught err: %v", err))
	}
}

func Must1[T any](a T, err error) T {
	Must(err)
	return a
}

func Must2[T, S any](a T, b S, err error) (T, S) {
	Must(err)
	return a, b
}
