package utils

import (
	"fmt"
)

func Ignore(err error) (caught bool) {
	caught = err != nil
	return
}

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
