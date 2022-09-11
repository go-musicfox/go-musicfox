package stdutil

import (
	"io"
	"os"
	"os/signal"
	"syscall"
)

// WaitCloseSignals for some huang program.
func WaitCloseSignals(closer io.Closer) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	<-signals

	return closer.Close()
}

// Go is a basic promise implementation: it wraps calls a function in a goroutine
// and returns a channel which will later return the function's return value.
func Go(f func() error) error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()

	return <-ch
}
