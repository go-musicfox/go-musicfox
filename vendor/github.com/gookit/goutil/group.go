package goutil

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ErrGroup is a collection of goroutines working on subtasks that
// are part of the same overall task.
//
// Refers:
//
//	https://github.com/neilotoole/errgroup
//	https://github.com/fatih/semgroup
type ErrGroup struct {
	*errgroup.Group
}

// NewCtxErrGroup instance
func NewCtxErrGroup(ctx context.Context, limit ...int) (*ErrGroup, context.Context) {
	egg, ctx1 := errgroup.WithContext(ctx)
	if len(limit) > 0 && limit[0] > 0 {
		egg.SetLimit(limit[0])
	}

	eg := &ErrGroup{Group: egg}
	return eg, ctx1
}

// NewErrGroup instance
func NewErrGroup(limit ...int) *ErrGroup {
	eg := &ErrGroup{Group: new(errgroup.Group)}

	if len(limit) > 0 && limit[0] > 0 {
		eg.SetLimit(limit[0])
	}
	return eg
}

// Add one or more handler at once
func (g *ErrGroup) Add(handlers ...func() error) {
	for _, handler := range handlers {
		g.Go(handler)
	}
}
