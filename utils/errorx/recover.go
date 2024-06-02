package errorx

import (
	"log/slog"
	"runtime/debug"
)

func Recover(ignore bool) (hasCaught bool) {
	err := recover()
	if err != nil {
		slog.Error("catch panic", slog.Any("error", err), slog.Any("stack", debug.Stack()))
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

func Go(f func(), ignorePanic ...bool) {
	var ignore bool
	if len(ignorePanic) > 0 {
		ignore = ignorePanic[0]
	}
	go PanicRecoverWrapper(ignore, f)
}

func WaitGoStart(f func(), ignorePanic ...bool) {
	var wait = make(chan struct{})
	Go(func() {
		Go(f, ignorePanic...)
		wait <- struct{}{}
	}, ignorePanic...)
	<-wait
}
