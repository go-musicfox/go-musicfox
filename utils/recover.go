package utils

import (
	"runtime/debug"
)

func Recover(ignore bool) (hasCaught bool) {
	err := recover()
	if err != nil {
		Logger().Printf("catch panic, err: %+v, stack: %s", err, debug.Stack())
		if ignore {
			hasCaught = true
			return
		}
		panic(err)
	}
	return
}

func PanicRecoverWrapper(ignorePanic bool, f func()) {
	defer Recover(ignorePanic)
	f()
}
