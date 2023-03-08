package stdutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// WaitCloseSignals for some huang program.
func WaitCloseSignals(closer io.Closer) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
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

// SignalHandler returns an actor, i.e. an execute and interrupt func, that
// terminates with SignalError when the process receives one of the provided
// signals, or the parent context is canceled.
//
// from https://github.com/oklog/run/blob/master/actors.go
func SignalHandler(ctx context.Context, signals ...os.Signal) (execute func() error, interrupt func(error)) {
	ctx, cancel := context.WithCancel(ctx)
	return func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, signals...)
			defer signal.Stop(c)
			select {
			case sig := <-c:
				return SignalError{Signal: sig}
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func(error) {
			cancel()
		}
}

// SignalError is returned by the signal handler's execute function
// when it terminates due to a received signal.
type SignalError struct {
	Signal os.Signal
}

// Error implements the error interface.
func (e SignalError) Error() string {
	return fmt.Sprintf("received signal %s", e.Signal)
}
