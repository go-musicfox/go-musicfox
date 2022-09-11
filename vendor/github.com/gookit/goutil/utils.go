package goutil

import (
	"github.com/gookit/goutil/jsonutil"
)

// Go is a basic promise implementation: it wraps calls a function in a goroutine
// and returns a channel which will later return the function's return value.
func Go(f func() error) error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()
	return <-ch
}

// Filling filling a model from submitted data
// form 提交过来的数据结构体
// model 定义表模型的数据结构体
// 相当于是在合并两个结构体(data 必须是 model 的子集)
func Filling(form interface{}, model interface{}) error {
	jsonBytes, _ := jsonutil.Encode(form)
	return jsonutil.Decode(jsonBytes, model)
}
