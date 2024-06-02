package util

import (
	"math/rand"
	"time"
)

// RandStringRunes 返回随机字符串
func RandStringRunes(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}

func StringOr(a, b string, others ...string) string {
	if a != "" {
		return a
	}
	if b != "" {
		return b
	}
	for _, v := range others {
		if v != "" {
			return v
		}
	}
	return ""
}

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
