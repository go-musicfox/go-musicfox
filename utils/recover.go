package utils

import "runtime/debug"

func Recover(ignore bool) {
	err := recover()
	if err != nil {
		Logger().Printf("catch panic, err: %+v, stack: %s", err, debug.Stack())
		if ignore {
			return
		}
		panic(err)
	}
}
